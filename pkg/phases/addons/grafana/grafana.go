package grafana

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

type GrafanaConfig struct {
	Image        string
	SidecarImage string
	IngressHost  string
}

func NewGrafanaConfig(cfg *kubeadmapi.ClusterConfiguration) addons.Configer {
	repo := cfg.ImageRepository
	config := GrafanaConfig{
		Image:        images.GetGenericImage(repo, constants.Grafana, constants.DefaultGrafanaVersion),
		SidecarImage: images.GetGenericImage(repo, constants.K8sSidecar, constants.DefaultK8sSidecarVersion),
		// TODO: support customize domain
		IngressHost: "grafana.test.io",
	}
	return config
}

func (c GrafanaConfig) Name() string {
	return "grafana"
}

func (c GrafanaConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(GrafanaTempate, c)
}
