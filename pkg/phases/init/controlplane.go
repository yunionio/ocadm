package init

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/controlplane"
)

func NewOCControlPlanePhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  "oc-control-plane",
		Short: "Generates all static Pod manifest files necessary to establish the control plane",
		Phases: []workflow.Phase{
			{
				Name:           "all",
				Short:          "Generates all static Pod manifest files",
				InheritFlags:   getControlPlanePhaseFlags("all"),
				RunAllSiblings: true,
			},
			//newControlPlaneSubphase(constants.OnecloudKeystone),
		},
		Run: runControlPlanePhase,
	}
	return phase
}

func newControlPlaneSubphase(component string) workflow.Phase {
	phase := workflow.Phase{
		Name:         controlPlanePhaseProperties[component].name,
		Short:        controlPlanePhaseProperties[component].short,
		Run:          runControlPlaneSubphase(component),
		InheritFlags: getControlPlanePhaseFlags(component),
	}
	return phase
}

func getControlPlanePhaseFlags(name string) []string {
	flags := []string{
		options.CfgPath,
		options.ImageRepository,
	}
	return flags
}

func runControlPlanePhase(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("control-plane phase invoked with an invalid data struct")
	}

	fmt.Printf("[oc-control-plane] Using manifest folder %q\n", data.ManifestDir())
	return nil
}

func runControlPlaneSubphase(component string) func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		data, ok := c.(InitData)
		if !ok {
			return errors.New("oc-control-plane phase invoked with an invalid data struct")
		}
		cfg := data.OnecloudCfg()

		fmt.Printf("[oc-control-plane] Creating static Pod manifest for %q\n", component)
		return controlplane.CreateStaticPodFiles(data.ManifestDir(), &cfg.ClusterConfiguration, component)
	}
}
