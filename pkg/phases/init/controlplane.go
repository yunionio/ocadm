package init

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/ocadm/pkg/phases/controlplane"
)

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
