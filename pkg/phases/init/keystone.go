package init

import (
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/ocadm/pkg/phases/components/keystone"
)

// NewKeystonePhase creates a ocadm workflow phase that implements handing of keystone
func NewKeystonePhase() workflow.Phase {
	return keystone.KeystoneComponent.ToInstallPhase()
}
