package cmd

import (
	"io"

	"github.com/spf13/cobra"

	componentsphase "yunion.io/x/ocadm/pkg/phases/components"

	_ "yunion.io/x/ocadm/pkg/phases/components/init"
)

func NewCmdComponent(out io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "component",
		Short: "Manage onecloud components",
	}

	cmds.AddCommand(componentsphase.InstallCmd.GetCmd())
	cmds.AddCommand(componentsphase.UninstallCmd.GetCmd())

	return cmds
}
