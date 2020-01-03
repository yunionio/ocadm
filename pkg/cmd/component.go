package cmd

import (
	"io"

	"github.com/spf13/cobra"

	componentphase "yunion.io/x/ocadm/pkg/phases/component"
)

func NewCmdComponent(out io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "component",
		Short: "Manage onecloud extra components",
	}
	cmds.AddCommand(componentphase.EnableCmd.GetCmd())
	cmds.AddCommand(componentphase.DisableCmd.GetCmd())
	cmds.AddCommand(componentphase.NewCmdConfig(out))
	return cmds
}
