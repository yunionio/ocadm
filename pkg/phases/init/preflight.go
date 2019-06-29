package init

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	utilsexec "k8s.io/utils/exec"

	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/preflight"
)

func NewPreflightPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "preflight",
		Short: "Run pre-flight checks",
		Long:  "Run pre-flight checks for onecloud and kubeadm init.",
		Run:   runPreflight,
		InheritFlags: []string{
			options.CfgPath,
			options.KubeadmCfgPath,
			options.IgnorePreflightErrors,
		},
	}
}

func runPreflight(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("preflight phase invoked with an invalid data struct")
	}

	fmt.Println("[preflight] Running pre-flight checks")
	if err := preflight.RunInitNodeChecks(utilsexec.New(), data.OnecloudCfg(), data.Cfg(), data.IgnorePreflightErrors(), false, false); err != nil {
		return err
	}

	if !data.DryRun() {
		fmt.Println("[preflight] Pulling images required for setting up a OneCloud on Kubernetes cluster")
		fmt.Println("[preflight] This might take a minute or two, depending on the speed of your internet connection")
		fmt.Println("[preflight] You can also perform this action in beforehand using 'ocadm config images pull'")
		if err := preflight.RunPullImagesCheck(utilsexec.New(), data.OnecloudCfg(), data.Cfg(), data.IgnorePreflightErrors()); err != nil {
			return err
		}
	} else {
		fmt.Println("[preflight] Would pull the required images (like 'deployer config images pull')")
	}

	return nil
}
