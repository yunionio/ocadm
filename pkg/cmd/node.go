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
	"yunion.io/x/ocadm/pkg/phases/nodelabels"
)

func NewCmdNode(out io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "node",
		Short: "nodes management",
	}

	for _, cmd := range cmdSetNodeLables() {
		cmds.AddCommand(cmd)
	}

	return cmds
}

type nodeOptions interface {
	GetNodes() []string
	ClientSet() (*clientset.Clientset, error)
	GetLabels() map[string]string
	Command() string
	Short() string
	AddNodesFlags(*flag.FlagSet)
}

type nodesBaseData struct {
	nodes     []string
	clientSet *clientset.Clientset
}

func (h *nodesBaseData) GetNodes() []string {
	return h.nodes
}

func (h *nodesBaseData) ClientSet() (*clientset.Clientset, error) {
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

func (h *nodesBaseData) AddNodesFlags(flagSet *flag.FlagSet) {
	flagSet.StringArrayVar(
		&h.nodes, "node", h.nodes,
		"Node names to configure",
	)
}

func cmdSetNodeLables() []*cobra.Command {
	var cmds = make([]*cobra.Command, 0)
	for _, nodeOption := range []nodeOptions{&hostEnableData{}, &hostDisableData{}, &onecloudControllerEnableData{}, &onecloudControllerDisableData{}} {
		cmd := func(opt nodeOptions) *cobra.Command {
			runner := workflow.NewRunner()
			cmd := &cobra.Command{
				Use:   opt.Command(),
				Short: opt.Short(),
				PreRun: func(cmd *cobra.Command, args []string) {
					if len(opt.GetNodes()) == 0 {
						cmd.Help()
						kubeadmutil.CheckErr(errors.New("Need input nodes"))
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
			runner.AppendPhase(nodelabels.NodesSetLabels())
			runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
				return opt, nil
			})
			runner.BindToCommand(cmd)
			return cmd
		}(nodeOption)
		cmds = append(cmds, cmd)
	}
	return cmds
}
