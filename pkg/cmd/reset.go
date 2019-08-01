package cmd

import (
	"fmt"
	"io"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmapiv1beta2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/reset"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	utilruntime "k8s.io/kubernetes/cmd/kubeadm/app/util/runtime"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	configutil "yunion.io/x/ocadm/pkg/util/config"
)

var (
	iptablesCleanupInstructions = dedent.Dedent(`
		The reset process does not reset or clean up iptables rules or IPVS tables.
		If you wish to reset iptables, you must do so manually.
		For example:
		iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X

		If your cluster was setup to utilize IPVS, run ipvsadm --clear (or similar)
		to reset your system's IPVS tables.

		The reset process does not clean your kubeconfig files and you must remove them manually.
		Please, check the contents of the $HOME/.kube/config file.
	`)
)

// resetOptions defines all the options exposed via flags by kubeadm reset.
type resetOptions struct {
	certificatesDir       string
	criSocketPath         string
	forceReset            bool
	ignorePreflightErrors []string
	kubeconfigPath        string
}

// resetData defines all the runtime information used when running the kubeadm reset workflow;
// this data is shared across all the phases that are included in the workflow.
type resetData struct {
	certificatesDir       string
	client                clientset.Interface
	criSocketPath         string
	forceReset            bool
	ignorePreflightErrors sets.String
	inputReader           io.Reader
	outputWriter          io.Writer
	cfg                   *kubeadmapi.InitConfiguration
	dirsToClean           []string
}

// newResetOptions returns a struct ready for being used for creating cmd join flags.
func newResetOptions() *resetOptions {
	return &resetOptions{
		certificatesDir:       kubeadmapiv1beta2.DefaultCertificatesDir,
		forceReset:            false,
		kubeconfigPath:        constants.GetAdminKubeConfigPath(),
		ignorePreflightErrors: []string{},
	}
}

func newResetData(cmd *cobra.Command, options *resetOptions, in io.Reader, out io.Writer) (*resetData, error) {
	var ocCfg *apis.InitConfiguration
	var cfg *kubeadmapi.InitConfiguration

	client, err := getClientset(options.kubeconfigPath, false)
	if err == nil {
		klog.V(1).Infof("[reset] Loaded client set from kubeconfig file: %s", options.kubeconfigPath)
		ocCfg, err = configutil.FetchInitConfigurationFromCluster(client, out, "reset", false)
		if err != nil {
			klog.Warningf("[reset] Unable to fetch the kubeadm-config ConfigMap from cluster: %v", err)
		}
		cfg = &ocCfg.InitConfiguration
	} else {
		klog.V(1).Infof("[reset] Could not obtain a client set from the kubeconfig file: %s", options.kubeconfigPath)
	}

	ignorePreflightErrorsSet, err := validation.ValidateIgnorePreflightErrors(options.ignorePreflightErrors, ignorePreflightErrors(cfg))
	if err != nil {
		return nil, err
	}
	kubeadmutil.CheckErr(err)
	if cfg != nil {
		// Also set the union of pre-flight errors to InitConfiguration, to provide a consistent view of the runtime configuration:
		cfg.NodeRegistration.IgnorePreflightErrors = ignorePreflightErrorsSet.List()
	}

	var criSocketPath string
	if options.criSocketPath == "" {
		criSocketPath, err = resetDetectCRISocket(cfg)
		if err != nil {
			return nil, err
		}
		klog.V(1).Infof("[reset] Detected and using CRI socket: %s", criSocketPath)
	} else {
		criSocketPath = options.criSocketPath
		klog.V(1).Infof("[reset] Using specified CRI socket: %s", criSocketPath)
	}

	return &resetData{
		certificatesDir:       options.certificatesDir,
		client:                client,
		criSocketPath:         criSocketPath,
		forceReset:            options.forceReset,
		ignorePreflightErrors: ignorePreflightErrorsSet,
		inputReader:           in,
		outputWriter:          out,
		cfg:                   cfg,
	}, nil
}

func ignorePreflightErrors(cfg *kubeadmapi.InitConfiguration) []string {
	if cfg == nil {
		return []string{}
	}
	return cfg.NodeRegistration.IgnorePreflightErrors
}

// AddResetFlags adds reset flags
func AddResetFlags(flagSet *flag.FlagSet, resetOptions *resetOptions) {
	flagSet.StringVar(
		&resetOptions.certificatesDir, options.CertificatesDir, resetOptions.certificatesDir,
		`The path to the directory where the certificates are stored. If specified, clean this directory.`,
	)
	flagSet.BoolVarP(
		&resetOptions.forceReset, options.ForceReset, "f", false,
		"Reset the node without prompting for confirmation.",
	)

	options.AddKubeConfigFlag(flagSet, &resetOptions.kubeconfigPath)
	options.AddIgnorePreflightErrorsFlag(flagSet, &resetOptions.ignorePreflightErrors)
	cmdutil.AddCRISocketFlag(flagSet, &resetOptions.criSocketPath)
}

// NewCmdReset returns the "ocadm reset" command
func NewCmdReset(in io.Reader, out io.Writer, resetOptions *resetOptions) *cobra.Command {
	if resetOptions == nil {
		resetOptions = newResetOptions()
	}
	resetRunner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Run this to revert any changes made to this host by 'kubeadm init' or 'kubeadm join'",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := resetRunner.InitData(args)
			kubeadmutil.CheckErr(err)

			err = resetRunner.Run(args)
			kubeadmutil.CheckErr(err)

			// Then clean contents from the stateful kubelet, etcd and cni directories
			data := c.(*resetData)
			cleanDirs(data)

			// Output help text instructing user how to remove iptables rules
			fmt.Print(iptablesCleanupInstructions)
		},
	}

	AddResetFlags(cmd.Flags(), resetOptions)

	// initialize the workflow runner with the list of phases
	resetRunner.AppendPhase(phases.NewPreflightPhase())
	resetRunner.AppendPhase(phases.NewUpdateClusterStatus())
	resetRunner.AppendPhase(phases.NewRemoveETCDMemberPhase())
	resetRunner.AppendPhase(phases.NewCleanupNodePhase())

	// sets the data builder function, that will be used by the runner
	// both when running the entire workflow or single phases
	resetRunner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newResetData(cmd, resetOptions, in, out)
	})

	// binds the Runner to kubeadm init command by altering
	// command help, adding --skip-phases flag and by adding phases subcommands
	resetRunner.BindToCommand(cmd)

	return cmd
}

func cleanDirs(data *resetData) {
	fmt.Printf("[reset] Deleting contents of stateful directories: %v\n", data.dirsToClean)
	for _, dir := range data.dirsToClean {
		klog.V(1).Infof("[reset] Deleting content of %s", dir)
		phases.CleanDir(dir)
	}
}

// Cfg returns the InitConfiguration.
func (r *resetData) Cfg() *kubeadmapi.InitConfiguration {
	return r.cfg
}

// CertificatesDir returns the CertificatesDir.
func (r *resetData) CertificatesDir() string {
	return r.certificatesDir
}

// Client returns the Client for accessing the cluster.
func (r *resetData) Client() clientset.Interface {
	return r.client
}

// ForceReset returns the forceReset flag.
func (r *resetData) ForceReset() bool {
	return r.forceReset
}

// InputReader returns the io.reader used to read messages.
func (r *resetData) InputReader() io.Reader {
	return r.inputReader
}

// IgnorePreflightErrors returns the list of preflight errors to ignore.
func (r *resetData) IgnorePreflightErrors() sets.String {
	return r.ignorePreflightErrors
}

// AddDirsToClean add a list of dirs to the list of dirs that will be removed.
func (r *resetData) AddDirsToClean(dirs ...string) {
	r.dirsToClean = append(r.dirsToClean, dirs...)
}

// CRISocketPath returns the criSocketPath.
func (r *resetData) CRISocketPath() string {
	return r.criSocketPath
}

func resetDetectCRISocket(cfg *kubeadmapi.InitConfiguration) (string, error) {
	if cfg != nil {
		// first try to get the CRI socket from the cluster configuration
		return cfg.NodeRegistration.CRISocket, nil
	}

	// if this fails, try to detect it
	return utilruntime.DetectCRISocket()
}
