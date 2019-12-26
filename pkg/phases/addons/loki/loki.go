package loki

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

type LokiConfig struct {
	Image string
}

func NewLokiConfig(cfg *kubeadmapi.ClusterConfiguration) addons.Configer {
	repo := cfg.ImageRepository
	config := LokiConfig{
		Image: images.GetGenericImage(repo, constants.Loki, constants.DefaultLokiVersion),
	}
	return config
}

func (c LokiConfig) Name() string {
	return "loki"
}

func (c LokiConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(LokiTemplate, c)
}

type PromtailConfig struct {
	Image string
}

func NewPromtailConfig(cfg *kubeadmapi.ClusterConfiguration) addons.Configer {
	repo := cfg.ImageRepository
	promtailConfig := PromtailConfig{
		Image: images.GetGenericImage(repo, constants.Promtail, constants.DefaultPromtailVersion),
	}
	return promtailConfig
}

func (c PromtailConfig) Name() string {
	return "promtail"
}

func (c PromtailConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(PromtailTemplate, c)
}
