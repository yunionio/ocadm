package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	kubeadmcmd "k8s.io/kubernetes/cmd/kubeadm/app/cmd"
)

func NewOneCloudAdminCommand(in io.Reader, out, err io.Writer) (*cobra.Command, []func() *cobra.Command) {
	cmds := &cobra.Command{
		Use:   "ocadm",
		Short: "Deploy and manage onecloud services on kubernetes cluster",
	}

	kubeadm := func() *cobra.Command {
		return kubeadmcmd.NewKubeadmCommand(os.Stdin, os.Stdout, os.Stderr)
	}

	cmds.ResetFlags()

	cmds.AddCommand(NewCmdConfig(out))
	cmds.AddCommand(NewCmdInit(out, nil))
	cmds.AddCommand(NewCmdReset(in, out))

	commandFns := []func() *cobra.Command{
		kubeadm,
	}

	for i := range commandFns {
		cmds.AddCommand(commandFns[i]())
	}

	return cmds, commandFns
}
