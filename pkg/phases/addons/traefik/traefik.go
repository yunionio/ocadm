package traefik

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

type TraefikConfig struct {
	Image string
}

func NewTraefikConfig(cfg *kubeadmapi.ClusterConfiguration) addons.Configer {
	repo := cfg.ImageRepository
	config := TraefikConfig{Image: images.GetGenericImage(repo, constants.IngressControllerTraefik, constants.DefaultTraefikVersion)}
	return config
}

func (c TraefikConfig) Name() string {
	return "traefik"
}

func (c TraefikConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(TraefikTemplate, c)
}
