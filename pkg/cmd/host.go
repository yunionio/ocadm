package cmd

import (
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/phases/host"
)

func NewCmdHost(out io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "host",
		Short: "Onecloud host-agent management",
	}

	cmds.AddCommand(cmdHostEnable())

	return cmds
}

type hostEnableData struct {
	nodes     []string
	clientSet *clientset.Clientset
}

func (h *hostEnableData) GetNodes() []string {
	return h.nodes
}

func (h *hostEnableData) ClientSet() (*clientset.Clientset, error) {
	if h.clientSet != nil {
		return h.clientSet, nil
	}
	path := constants.GetAdminKubeConfigPath()
	client, err := kubeconfigutil.ClientSetFromFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "[preflight] couldn't create Kubernetes client")
	}
	h.clientSet = client
	return client, nil
}

func AddHostEnableFlags(flagSet *flag.FlagSet, o *hostEnableData) {
	flagSet.StringArrayVar(
		&o.nodes, "node", o.nodes,
		"Node names to enable host agent",
	)
}

func cmdHostEnable() *cobra.Command {
	opt := &hostEnableData{}
	runner := workflow.NewRunner()
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Run this command to enable host agent",
		PreRun: func(cmd *cobra.Command, args []string) {
			if opt.nodes == nil {
				cmd.Help()
				kubeadmutil.CheckErr(errors.New("Enable host need input nodes"))
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
	AddHostEnableFlags(cmd.Flags(), opt)
	runner.AppendPhase(host.NodesEnableHostAgent())

	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return opt, nil
	})
	runner.BindToCommand(cmd)

	return cmd
}
