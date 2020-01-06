package onecloudoperator

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/constants"

	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

const (
	DefaultOperatorVersion = constants.DefaultOperatorVersion
)

type OperatorConfig struct {
	Image     string
	Namespace string
}

func NewOperatorConfig(cfg *kubeadmapi.ClusterConfiguration, version string) addons.Configer {
	repo := cfg.ImageRepository
	if version == "" {
		version = DefaultOperatorVersion
	}
	config := OperatorConfig{
		Image:     images.GetGenericImage(repo, "onecloud-operator", version),
		Namespace: constants.OnecloudNamespace,
	}
	return config
}

func (c OperatorConfig) Name() string {
	return "onecloud-operator"
}

func (c OperatorConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(OperatorTemplate, c)
}
