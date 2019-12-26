package csi

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

type LocalPathProvisionerConfig struct {
	Image string
}

func NewLocalPathProvisionerConfig(cfg *kubeadmapi.ClusterConfiguration) addons.Configer {
	repo := cfg.ImageRepository
	config := LocalPathProvisionerConfig{
		Image: images.GetGenericImage(repo, constants.RancherLocalPathProvisioner, constants.DefaultLocalProvisionerVersion),
	}
	return config
}

func (c LocalPathProvisionerConfig) Name() string {
	return "local-path-provisioner"
}

func (c LocalPathProvisionerConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(LocalPathProvisioner, c)
}
