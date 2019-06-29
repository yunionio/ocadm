package util

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"

	"yunion.io/x/ocadm/pkg/apis/constants"
)

var (
	SplitYAMLDocuments       = kubeadmutil.SplitYAMLDocuments
	GroupVersionKindsHasKind = kubeadmutil.GroupVersionKindsHasKind
)

// GroupVersionKindsHasClusterConfiguration returns whether the following gvk slice contains a ClusterConfiguration object
func GroupVersionKindsHasClusterConfiguration(gvks ...schema.GroupVersionKind) bool {
	return GroupVersionKindsHasKind(gvks, constants.ClusterConfigurationKind)
}

// GroupVersionKindsHasInitConfiguration returns whether the following gvk slice contains a InitConfiguration object
func GroupVersionKindsHasInitConfiguration(gvks ...schema.GroupVersionKind) bool {
	return GroupVersionKindsHasKind(gvks, constants.InitConfigurationKind)
}
