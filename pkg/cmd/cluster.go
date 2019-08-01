package cmd

import (
	"io"

	"github.com/spf13/cobra"

	clusterphase "yunion.io/x/ocadm/pkg/phases/cluster"
)

func NewCmdCluster(out io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "cluster",
		Short: "Onecloud cluster management",
	}

	cmds.AddCommand(clusterphase.NewCmdCreate(out))

	return cmds
}
