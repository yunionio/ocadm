package init

import (
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/ocadm/pkg/phases/components/region"
)

// NewRegionPhase creates a ocadm workflow phase that implements handing of region
func NewRegionPhase() workflow.Phase {
	return region.RegionComponent.ToInstallPhase()
}

func NewSchedulerPhase() workflow.Phase {
	return region.SchedulerComponent.ToInstallPhase()
}
