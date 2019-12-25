package component

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"

	util "yunion.io/x/ocadm/pkg/util/onecloud"
)

func NewCmdWeb() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Manage web UI component",
	}
	NewWebCmd(cmd, os.Stdout).Bind()
	return cmd
}

type WebCmd struct {
	*baseCmd
}

func NewWebCmd(cmd *cobra.Command, out io.Writer) *WebCmd {
	return &WebCmd{
		baseCmd: newBaseCmd(cmd, out),
	}
}

func (w *WebCmd) Bind() {
	w.baseCmd.AddCmd(w.newSubCmd("use-ce", w.useCE))
	w.baseCmd.AddCmd(w.newSubCmd("use-ee", w.useEE))
}

func (w *WebCmd) useCE(data *componentsData, _ io.Writer) error {
	// change onecloud web component image
	// disable yunion-agent
	return w.updateOnecloudCluster(func(oc *onecloud.OnecloudCluster) *onecloud.OnecloudCluster {
		return util.SetOCUseCE(oc)
	})
}

func (w *WebCmd) useEE(data *componentsData, _ io.Writer) error {
	// change onecloud web component image to ee
	// enable yunion-agent
	return w.updateOnecloudCluster(func(oc *onecloud.OnecloudCluster) *onecloud.OnecloudCluster {
		return util.SetOCUseEE(oc)
	})
}
