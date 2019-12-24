package cmd

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/phases/baremetal"
	"yunion.io/x/ocadm/pkg/phases/cluster"
	"yunion.io/x/onecloud-operator/pkg/client/clientset/versioned"
)

func NewCmdBaremetal(out io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "baremetal",
		Short: "Onecloud baremetal-agent management",
	}

	cmds.AddCommand(cmdBaremetalEnable())
	cmds.AddCommand(cmdBaremetalDisable())
	return cmds
}

type baremetalEnableData struct {
	nodesBaseData
	listenInterface string
	client          versioned.Interface
}

func (d *baremetalEnableData) GetListenInterface() string {
	return d.listenInterface
}

func (d *baremetalEnableData) VersionedClient() (versioned.Interface, error) {
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

func AddBaremetalEnableFlags(flagSet *flag.FlagSet, o *baremetalEnableData) {
	o.AddNodesFlags(flagSet)
	flagSet.StringVar(
		&o.listenInterface, "listen-interface", o.listenInterface,
		"Listen interface nome of baremetal agent",
	)
}

func cmdBaremetalEnable() *cobra.Command {
	opt := &baremetalEnableData{}
	runner := workflow.NewRunner()
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Run this command to select node enable baremetal agent",
		PreRun: func(cmd *cobra.Command, args []string) {
			if opt.nodes == nil {
				cmd.Help()
				kubeadmutil.CheckErr(errors.New("Enable baremetal need input nodes"))
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, err := runner.InitData(args)
			kubeadmutil.CheckErr(err)

			err = runner.Run(args)
			kubeadmutil.CheckErr(err)
		},
		Args: cobra.NoArgs,
	}
	AddBaremetalEnableFlags(cmd.Flags(), opt)
	runner.AppendPhase(baremetal.NodesEnableBaremetalAgent())

	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return opt, nil
	})
	runner.BindToCommand(cmd)

	return cmd
}

func cmdBaremetalDisable() *cobra.Command {
	opt := &baremetalEnableData{}
	runner := workflow.NewRunner()
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Run this command to select node disable baremetal agent",
		PreRun: func(cmd *cobra.Command, args []string) {
			if opt.nodes == nil {
				cmd.Help()
				kubeadmutil.CheckErr(errors.New("Disable baremetal need input nodes"))
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, err := runner.InitData(args)
			kubeadmutil.CheckErr(err)

			err = runner.Run(args)
			kubeadmutil.CheckErr(err)
		},
		Args: cobra.NoArgs,
	}
	opt.AddNodesFlags(cmd.Flags())
	runner.AppendPhase(baremetal.NodesDisableBaremetalAgent())

	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return opt, nil
	})
	runner.BindToCommand(cmd)

	return cmd
}
