package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"yunion.io/x/ocadm/pkg/util/mysql"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmapiv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	kubeadmcmd "k8s.io/kubernetes/cmd/kubeadm/app/cmd"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	utilruntime "k8s.io/kubernetes/cmd/kubeadm/app/util/runtime"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	configutil "yunion.io/x/ocadm/pkg/util/config"
)

// NewCmdReset returns the "ocadm reset" command
func NewCmdReset(in io.Reader, out io.Writer) *cobra.Command {
	var certsDir string
	var ignorePreflightErrors []string
	var forceReset bool
	var criSocketPath string
	var client clientset.Interface
	kubeConfigFile := constants.GetAdminKubeConfigPath()

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Run this to revert any changes made to this host by 'ocadm init' or 'ocadm join'",
		Run: func(cmd *cobra.Command, args []string) {
			ignorePreflightErrorsSet, err := validation.ValidateIgnorePreflightErrors(ignorePreflightErrors)
			kubeadmutil.CheckErr(err)

			var cfg *apis.InitConfiguration
			client, err = getClientset(kubeConfigFile, false)
			if err == nil {
				klog.V(1).Infof("[reset] Loaded client set from kubeconfig file: %s", kubeConfigFile)
				cfg, err = configutil.FetchInitConfigurationFromCluster(client, os.Stdout, "reset", false)
				if err != nil {
					klog.Warningf("[reset] Unable to fetch the kubeadm-config ConfigMap from cluster: %v", err)
				}
			} else {
				klog.V(1).Infof("[reset] Could not obtain a client set from the kubeconfig file: %s", kubeConfigFile)
			}

			if criSocketPath == "" {
				criSocketPath, err = resetDetectCRISocket(cfg)
				kubeadmutil.CheckErr(err)
				klog.V(1).Infof("[reset] Detected and using CRI socket: %s", criSocketPath)
			}

			r, err := NewReset(in, ignorePreflightErrorsSet, forceReset, certsDir, criSocketPath)
			kubeadmutil.CheckErr(err)
			kubeadmutil.CheckErr(r.Run(out, client, cfg))
		},
	}

	options.AddIgnorePreflightErrorsFlag(cmd.PersistentFlags(), &ignorePreflightErrors)

	cmd.PersistentFlags().StringVar(
		&certsDir, "cert-dir", kubeadmapiv1beta1.DefaultCertificatesDir,
		"The path to the directory where the certificates are stored. If specified, clean this directory.",
	)

	cmd.PersistentFlags().BoolVarP(
		&forceReset, "force", "f", false,
		"Reset the node without prompting for confirmation.",
	)
	return cmd
}

// Reset defines struct used for kubeadm reset command
type Reset struct {
	kubeadmReset *kubeadmcmd.Reset
}

// NewReset instantiate Reset struct
func NewReset(in io.Reader, ignorePreflightErrors sets.String, forceReset bool, certsDir, criSocketPath string) (*Reset, error) {
	kubeadmReset, err := kubeadmcmd.NewReset(in, ignorePreflightErrors, forceReset, certsDir, criSocketPath)
	if err != nil {
		return nil, err
	}

	return &Reset{
		kubeadmReset: kubeadmReset,
	}, nil
}

// Run reverts any changes made to this host by "ocadm init" or "ocadm join".
func (r *Reset) Run(out io.Writer, client clientset.Interface, cfg *apis.InitConfiguration) error {
	if isControlPlane() && cfg != nil {
		if err := r.resetDB(&cfg.MysqlConnection); err != nil {
			return errors.Wrap(err, "reset databases")
		}
	}
	var kubeadmInitCfg *kubeadmapi.InitConfiguration
	if cfg != nil {
		kubeadmInitCfg = &cfg.InitConfiguration
	}
	if err := r.kubeadmReset.Run(out, client, kubeadmInitCfg); err != nil {
		return errors.Wrap(err, "kubeadm reset")
	}
	dirsToClean := []string{constants.OnecloudConfigDir, constants.OnecloudOptTmpDir}
	fmt.Printf("[reset] Deleting contents of stateful directories: %v\n", dirsToClean)
	for _, dir := range dirsToClean {
		klog.V(1).Infof("[reset] Deleting content of %s", dir)
		cleanDir(dir)
	}
	return nil
}

func (r *Reset) resetDB(connInfo *apis.MysqlConnection) error {
	conn, err := mysql.NewConnection(connInfo)
	if err != nil {
		return err
	}
	for _, db := range []string{constants.KeystoneDB, constants.RegionDB} {
		if err := conn.DropDatabase(db); err != nil {
			return errors.Wrapf(err, "drop db %s", db)
		}
	}
	return nil
}

func resetDetectCRISocket(cfg *apis.InitConfiguration) (string, error) {
	if cfg != nil {
		// first try to get the CRI socket from the cluster configuration
		return cfg.NodeRegistration.CRISocket, nil
	}

	// if this fails, try to detect it
	return utilruntime.DetectCRISocket()
}

// isControlPlane checks if a node is a control-plane node by looking up
// the kube-apiserver manifest file
func isControlPlane() bool {
	filepath := constants.GetStaticPodFilepath(constants.OnecloudKeystone, constants.GetStaticPodDirectory())
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

// cleanDir removes everything in a directory, but not the directory itself
func cleanDir(filePath string) error {
	// If the directory doesn't even exist there's nothing to do, and we do
	// not consider this an error
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	d, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(filePath, name)); err != nil {
			return err
		}
	}
	return nil
}
