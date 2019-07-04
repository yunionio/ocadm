package config

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"yunion.io/x/ocadm/pkg/apis/constants"
	ocadmscheme "yunion.io/x/ocadm/pkg/apis/scheme"
	apis "yunion.io/x/ocadm/pkg/apis/v1"
)

// FetchInitConfigurationFromCluster fetches configuration from a ConfigMap in the cluster
func FetchInitConfigurationFromCluster(client clientset.Interface, w io.Writer, logPrefix string, newControlPlane bool) (*apis.InitConfiguration, error) {
	fmt.Fprintf(w, "[%s] Reading configuration from the cluster...\n", logPrefix)
	fmt.Fprintf(w, "[%s] FYI: You can look at this config file with 'kubectl -n %s get cm %s -oyaml'\n", logPrefix, metav1.NamespaceSystem, constants.OnecloudAdminConfigConfigMap)

	kubeadmCfg, err := config.FetchInitConfigurationFromCluster(client, w, logPrefix, newControlPlane)
	if err != nil {
		return nil, err
	}

	// Also, the config map really should be OcadmConfigConfigMap...
	configMap, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(constants.OnecloudAdminConfigConfigMap, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config map")
	}

	// InitConfiguration is composed with data from different places
	initcfg := &apis.InitConfiguration{}

	// gets ClusterConfiguration from kubeadm-config
	clusterConfigurationData, ok := configMap.Data[constants.ClusterConfigurationConfigMapKey]
	if !ok {
		return nil, errors.Errorf("unexpected error when reading kubeadm-config ConfigMap: %s key value pair missing", constants.ClusterConfigurationConfigMapKey)
	}
	if err := runtime.DecodeInto(ocadmscheme.Codecs.UniversalDecoder(), []byte(clusterConfigurationData), &initcfg.ClusterConfiguration); err != nil {
		return nil, errors.Wrap(err, "failed to decode init configuration data")
	}

	initcfg.InitConfiguration = *kubeadmCfg
	if err := SetInitDynamicDefaults(initcfg); err != nil {
		return nil, errors.Wrap(err, "failed to set dynamic defaults")
	}
	return initcfg, nil
}
