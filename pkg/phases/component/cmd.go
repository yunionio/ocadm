package component

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"

	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
)

var (
	EnableCmd  = NewComponentActionCmd("enable")
	DisableCmd = NewComponentActionCmd("disable")
)

type baseCmd struct {
	cmd  *cobra.Command
	out  io.Writer
	data *componentsData
}

func newBaseCmd(cmd *cobra.Command, out io.Writer) *baseCmd {
	return &baseCmd{
		cmd: cmd,
		out: out,
	}
}

func (b *baseCmd) Init() error {
	data, err := newComponentsData(b.cmd, nil, nil, b.out)
	if err != nil {
		return err
	}
	b.data = data
	return nil
}

func (b *baseCmd) AddCmd(cmd *cobra.Command) {
	b.cmd.AddCommand(cmd)
}

func (b *baseCmd) newSubCmd(use string, f func(data *componentsData, out io.Writer) error) *cobra.Command {
	return &cobra.Command{
		Use: use,
		Run: func(_ *cobra.Command, _ []string) {
			kubeadmutil.CheckErr(b.Init())
			kubeadmutil.CheckErr(f(b.data, b.out))
		},
	}
}

func (b *baseCmd) updateOnecloudCluster(f func(*onecloud.OnecloudCluster) *onecloud.OnecloudCluster) error {
	oc := b.data.OnecloudCluster().DeepCopy()
	oc = f(oc)
	_, err := b.data.ClusterClient().OnecloudV1alpha1().OnecloudClusters(oc.GetNamespace()).Update(oc)
	return err
}

type ComponentActionCmd struct {
	cmd       *cobra.Command
	allSubCmd *cobra.Command

	action string
	phases []workflow.Phase
}

func NewComponentActionCmd(action string) *ComponentActionCmd {
	return &ComponentActionCmd{
		action: action,
		cmd: &cobra.Command{
			Use:   action,
			Short: fmt.Sprintf("%s components", action),
			RunE:  cmdutil.SubCmdRunE(action),
		},
	}
}

func (a *ComponentActionCmd) GetAction() string {
	return a.action
}

type SubCmd struct {
	Cmd   *cobra.Command
	Phase workflow.Phase
}

func (a *ComponentActionCmd) AddCmd(cmd *SubCmd) *ComponentActionCmd {
	a.cmd.AddCommand(cmd.Cmd)
	a.phases = append(a.phases, cmd.Phase)
	return a
}

func (a *ComponentActionCmd) CompleteAllSubCmd() *ComponentActionCmd {
	runner := workflow.NewRunner()
	cOpt := newComponentsOptions()

	a.allSubCmd = &cobra.Command{
		Use:   "all",
		Short: fmt.Sprintf("%s all components", a.action),
		Args:  cobra.NoArgs,
		Run:   runComponentFunc(runner),
	}
	a.cmd.AddCommand(a.allSubCmd)
	for _, p := range a.phases {
		runner.AppendPhase(p)
	}
	runComponentSetDataInitializer(runner, cOpt)
	return a
}

func (a *ComponentActionCmd) GetCmd() *cobra.Command {
	return a.cmd
}
