package config

import (
	"k8s.io/apimachinery/pkg/runtime"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	configutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	ocadmscheme "yunion.io/x/ocadm/pkg/apis/scheme"
	apis "yunion.io/x/ocadm/pkg/apis/v1"
)

var (
	VerifyAPIServerBindAddress = configutil.VerifyAPIServerBindAddress
)

// MarshalOcadmConfigObject marshals an Object registered in the ocadm scheme. If the object is a InitConfiguration or ClusterConfiguration, some extra logic is run
func MarshalOcadmConfigObject(obj runtime.Object) ([]byte, error) {
	switch internalcfg := obj.(type) {
	case *apis.InitConfiguration:
		return MarshalInitConfigurationToBytes(internalcfg, apis.SchemeGroupVersion)
	case *apis.ClusterConfiguration:
		return MarshalClusterConfigurationToBytes(internalcfg, apis.SchemeGroupVersion)
	default:
		return kubeadmutil.MarshalToYamlForCodecs(obj, apis.SchemeGroupVersion, ocadmscheme.Codecs)
	}
}
