package cmd

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"yunion.io/x/ocadm/pkg/util/kubectl"

	"yunion.io/x/ocadm/pkg/apis/constants"
	v1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/cluster"
	"yunion.io/x/ocadm/pkg/phases/longhorn"
	"yunion.io/x/onecloud-operator/pkg/client/clientset/versioned"
)

func NewCmdLonghorn(out io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "longhorn",
		Short: "Deployment longhorn management",
	}

	cmds.AddCommand(cmdLonghornEnable())
	return cmds
}

func cmdLonghornEnable() *cobra.Command {
	opt := &longHornConfig{}
	runner := workflow.NewRunner()
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Run this command to enable longhorn",
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, err := runner.InitData(args)
			kubeadmutil.CheckErr(err)

			err = runner.Run(args)
			kubeadmutil.CheckErr(err)
		},
		Args: cobra.NoArgs,
	}
	AddLonghornFlags(cmd.Flags(), opt)
	runner.AppendPhase(longhorn.InstallLonghornPhase())

	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		_, err := opt.VersionedClient()
		if err != nil {
			return nil, err
		}
		err = opt.VerifyNode()
		if err != nil {
			return nil, err
		}
		opt.kubeconfigPath = kubeadmconstants.GetAdminKubeConfigPath()
		_, err = opt.KubectlClient()
		if err != nil {
			return nil, err
		}
		return opt, nil
	})
	runner.BindToCommand(cmd)
	return cmd
}

type longHornConfig struct {
	nodesBaseData
	DataPath                   string
	OverProvisioningPercentage int
	ReplicaCount               int
	ImageRepository            string
	client                     versioned.Interface
	kubectlClient              *kubectl.Client
	kubeconfigPath             string
}

func AddLonghornFlags(flagSet *flag.FlagSet, o *longHornConfig) {
	// longhorn configuration
	flagSet.StringVar(
		&o.DataPath, options.LonghornDataPath,
		constants.LonghornDefaultDataPath, "Longhorn data path, default /var/lib/longhorn",
	)
	flagSet.IntVar(
		&o.OverProvisioningPercentage,
		options.LonghornOverProvisioningPercentage,
		constants.LonghornDefaultOverProvisioningPercentage,
		"Longhorn disk over provisioning percentage, default 100",
	)
	//flagSet.IntVar(
	//	&o.ReplicaCount, options.LonghornReplicaCount,
	//	constants.LonghornDefaultReplicaCount, "Longhorn replica count, default 3",
	//)
	flagSet.StringVar(
		&o.ImageRepository, options.ImageRepository,
		v1.DefaultImageRepository, "Image repository",
	)
	o.AddNodesFlags(flagSet)
}

func (d *longHornConfig) GetImageRepository() string {
	return d.ImageRepository
}

func (d *longHornConfig) VersionedClient() (versioned.Interface, error) {
	if d.client != nil {
		return d.client, nil
	}

	var tlsBootstrapCfg *clientcmdapi.Config
	var err error

	kubeConfigFile := constants.GetAdminKubeConfigPath()
	if _, err := os.Stat(kubeConfigFile); err == nil {
		// use the admin.conf as tlsBootstrapCfg, that is the kubeconfig file used for reading the ocadm-config during dicovery
		klog.V(1).Infof("[preflight] found %s. Use it for skipping discovery", kubeConfigFile)
		tlsBootstrapCfg, err = clientcmd.LoadFromFile(kubeConfigFile)
		if err != nil {
			return nil, errors.Wrapf(err, "Error loading %s", kubeConfigFile)
		}
	} else {
		return nil, err
	}

	if tlsBootstrapCfg == nil {
		return nil, errors.Errorf("Not found valid %s, please run this command at controlplane", kubeConfigFile)
	}
	cli, err := cluster.NewClusterClient(tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}
	d.client = cli
	return cli, nil
}

func (d *longHornConfig) LonghornConfig() *longhorn.LonghornConfig {
	return &longhorn.LonghornConfig{
		DataPath:                    d.DataPath,
		OverProviosioningPercentage: d.OverProvisioningPercentage,
		ReplicaCount:                d.ReplicaCount,
	}
}

func (d *longHornConfig) VerifyNode() error {
	cli, err := d.ClientSet()
	if err != nil {
		return err
	}
	if len(d.nodes) == 0 {
		return nil
	} else {
		for i := 0; i < len(d.nodes); i++ {
			_, err = cli.CoreV1().Nodes().Get(d.nodes[i], metav1.GetOptions{})
			if err != nil {
				return errors.Wrapf(err, "get node %s", d.nodes[i])
			}
		}
		return nil
	}
}

// KubeConfigPath returns the path to the kubeconfig file to use for connecting to Kubernetes
func (d *longHornConfig) KubeConfigPath() string {
	return d.kubeconfigPath
}

func (d *longHornConfig) KubectlClient() (*kubectl.Client, error) {
	if d.kubectlClient != nil {
		return d.kubectlClient, nil
	}
	cli, err := kubectl.NewClientFormKubeconfigFile(d.KubeConfigPath())
	if err != nil {
		return nil, err
	}
	d.kubectlClient = cli
	return d.kubectlClient, nil
}
