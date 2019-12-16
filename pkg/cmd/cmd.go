package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

func NewOneCloudAdminCommand(in io.Reader, out, err io.Writer) (*cobra.Command, []func() *cobra.Command) {
	cmds := &cobra.Command{
		Use:   "ocadm",
		Short: "Deploy and manage onecloud services on kubernetes cluster",
	}

	/*kubeadm := func() *cobra.Command {
		return kubeadmcmd.NewKubeadmCommand(os.Stdin, os.Stdout, os.Stderr)
	}*/

	cmds.ResetFlags()

	cmds.AddCommand(NewCmdConfig(out))
	cmds.AddCommand(NewCmdInit(out, nil))
	cmds.AddCommand(NewCmdJoin(out, nil))
	cmds.AddCommand(NewCmdReset(in, out, nil))
	cmds.AddCommand(NewCmdToken(out, err))
	cmds.AddCommand(NewCmdCluster(out))
	cmds.AddCommand(NewCmdComponent(out))
	cmds.AddCommand(NewCmdHost(out))
	cmds.AddCommand(NewCmdBaremetal(out))

	commandFns := []func() *cobra.Command{
		//kubeadm,
	}

	for i := range commandFns {
		cmds.AddCommand(commandFns[i]())
	}

	return cmds, commandFns
}
