package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/klog"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"

	"yunion.io/x/pkg/util/version"
)

// NewCmdVersion provides the version information of ocadm.
func NewCmdVersion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of ocadm",
		Run: func(cmd *cobra.Command, args []string) {
			err := RunVersion(out, cmd)
			kubeadmutil.CheckErr(err)
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output format; available options are 'json' and 'short'")
	return cmd
}

// RunVersion provides the version information of ocadm in format depending on arguments
// specified in cobra.Command
func RunVersion(out io.Writer, cmd *cobra.Command) error {
	klog.V(1).Infoln("[version] retrieving version info")
	const flag = "output"
	of, err := cmd.Flags().GetString(flag)
	if err != nil {
		klog.Fatalf("error accessing flag %s for command %s: %v", flag, cmd.Name(), err)
	}
	switch of {
	case "":
		fmt.Fprintf(out, "ocadm version: %#v\n", version.Get())
	case "short":
		fmt.Fprintf(out, "%s\n", version.GetShortString())
	case "json":
		fmt.Fprintln(out, version.GetJsonString())
	default:
		return errors.Errorf("invalid output format: %s", of)
	}
	return nil
}
