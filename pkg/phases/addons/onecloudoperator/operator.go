package onecloudoperator

import (
	"fmt"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/constants"

	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
	"yunion.io/x/ocadm/pkg/util/kubectl"
)

const (
	DefaultOperatorVersion = "latest"
)

type OperatorConfig struct {
	Image     string
	Namespace string
}

func (c OperatorConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(OperatorTemplate, c)
}

func EnsureOnecloudOperatorAddon(cfg *kubeadmapi.ClusterConfiguration, kubectlCli *kubectl.Client) error {
	repo := cfg.ImageRepository
	config := OperatorConfig{
		Image:     images.GetGenericImage(repo, "onecloud-operator", DefaultOperatorVersion),
		Namespace: constants.OnecloudNamespace,
	}
	manifest, err := config.GenerateYAML()
	if err != nil {
		return err
	}
	if err := kubectlCli.Apply(manifest); err != nil {
		return err
	}
	fmt.Println("[oc-addons] Applied essential addon: onecloud-operator")
	return nil
}

type LocalPathProvisionerConfig struct {
	Image string
}

func (c LocalPathProvisionerConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(LocalPathProvisioner, c)
}

func EnsureLocalPathProvisionerAddon(cfg *kubeadmapi.ClusterConfiguration, client *kubectl.Client) error {
	repo := cfg.ImageRepository
	config := LocalPathProvisionerConfig{Image: images.GetGenericImage(repo, constants.RancherLocalPathProvisioner, constants.DefaultLocalProvisionerVersion)}
	manifest, err := config.GenerateYAML()
	if err != nil {
		return err
	}
	if err := client.Apply(manifest); err != nil {
		return err
	}
	fmt.Println("[oc-addons] Applied essential addon: local-path-provisioner")
	return nil
}

type TraefikConfig struct {
	Image string
}

func (c TraefikConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(TraefikTemplate, c)
}

func EnsureIngressTraefikAddon(cfg *kubeadmapi.ClusterConfiguration, client *kubectl.Client) error {
	repo := cfg.ImageRepository
	config := TraefikConfig{Image: images.GetGenericImage(repo, constants.IngressControllerTraefik, constants.DefaultTraefikVersion)}
	manifest, err := config.GenerateYAML()
	if err != nil {
		return err
	}
	if err := client.Apply(manifest); err != nil {
		return err
	}
	fmt.Println("[oc-addons] Applied essential addon: traefik")
	return nil
}
