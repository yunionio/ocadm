package uploadconfig

import (
	"fmt"
	"io/ioutil"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	configutil "yunion.io/x/ocadm/pkg/util/config"
)

func UploadConfiguration(cfg *apis.InitConfiguration, client clientset.Interface) error {
	fmt.Printf("[oc-upload-config] storing the configuration used in ConfigMap %q in the %q Namespace\n", constants.OnecloudAdminConfigConfigMap, metav1.NamespaceSystem)

	// Prepare the ClusterConfiguration for upload
	// The components store their config in their own ConfigMaps,
	// We don't want to mutate the cfg itself, so create a copy of it using .DeepCopy of it first
	clusterConfigurationToUpload := cfg.ClusterConfiguration.DeepCopy()

	// Marshal the ClusterConfiguration into YAML
	clusterConfigurationYaml, err := configutil.MarshalOcadmConfigObject(clusterConfigurationToUpload)
	if err != nil {
		return err
	}

	authInfo, err := ioutil.ReadFile(occonfig.AdminConfigFilePath())
	if err != nil {
		return err
	}

	err = apiclient.CreateOrUpdateConfigMap(client, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.OnecloudAdminConfigConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			constants.ClusterConfigurationConfigMapKey: string(clusterConfigurationYaml),
			constants.ClusterAdminAuthConfigMapKey:     string(authInfo),
		},
	})
	if err != nil {
		return err
	}
	return nil
}
