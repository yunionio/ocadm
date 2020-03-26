package component

import (
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

func init() {
	registerComponentCmds()
}

func registerComponentCmds() {
	components := []IComponent{
		ItsmComponent,
	}
	addSubCmds(components...)
}

func addSubCmds(cs ...IComponent) {
	for _, c := range cs {
		addCmdSubCmd(EnableCmd, c.ToEnableCmd(), c.ToEnablePhase())
		addCmdSubCmd(DisableCmd, c.ToDisableCmd(), c.ToDisablePhase())
	}
	EnableCmd.CompleteAllSubCmd()
	DisableCmd.CompleteAllSubCmd()
}

func addCmdSubCmd(actionCmd *ComponentActionCmd, cmd *cobra.Command, phase workflow.Phase) {
	actionCmd.AddCmd(&SubCmd{
		Cmd:   cmd,
		Phase: phase,
	})
}
