package components

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
)

var (
	InstallCmd   *ComponentActionCmd
	UninstallCmd *ComponentActionCmd
)

func init() {
	InstallCmd = NewComponentActionCmd("install")
	UninstallCmd = NewComponentActionCmd("uninstall")
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
