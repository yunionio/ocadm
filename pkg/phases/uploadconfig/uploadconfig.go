package uploadconfig

import (
	"fmt"
	"io/ioutil"

	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/phases/uploadconfig"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	rbachelper "k8s.io/kubernetes/pkg/apis/rbac/v1"

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

	// Ensure that the NodesKubeadmConfigClusterRoleName exists
	err = apiclient.CreateOrUpdateRole(client, &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uploadconfig.NodesKubeadmConfigClusterRoleName,
			Namespace: metav1.NamespaceSystem,
		},
		Rules: []rbac.PolicyRule{
			rbachelper.NewRule("get").Groups("").Resources("configmaps").Names(kubeadmconstants.KubeadmConfigConfigMap).RuleOrDie(),
			rbachelper.NewRule("get").Groups("").Resources("configmaps").Names(constants.OnecloudAdminConfigConfigMap).RuleOrDie(),
		},
	})
	if err != nil {
		return err
	}

	// Binds the NodesKubeadmConfigClusterRoleName to all the bootstrap tokens
	// that are members of the system:bootstrappers:kubeadm:default-node-token group
	// and to all nodes
	return apiclient.CreateOrUpdateRoleBinding(client, &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uploadconfig.NodesKubeadmConfigClusterRoleName,
			Namespace: metav1.NamespaceSystem,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "Role",
			Name:     uploadconfig.NodesKubeadmConfigClusterRoleName,
		},
		Subjects: []rbac.Subject{
			{
				Kind: rbac.GroupKind,
				Name: kubeadmconstants.NodeBootstrapTokenAuthGroup,
			},
			{
				Kind: rbac.GroupKind,
				Name: kubeadmconstants.NodesGroup,
			},
		},
	})
}
