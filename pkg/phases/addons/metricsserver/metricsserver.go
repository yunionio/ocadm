package metricsserver

import (
	"runtime"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

type MetricsServerConfig struct {
	Image string
	Arch  string
}

func NewMetricsServerConfig(cfg *kubeadmapi.ClusterConfiguration) addons.Configer {
	arch := runtime.GOARCH
	repo := cfg.ImageRepository
	config := MetricsServerConfig{
		Image: images.GetGenericImage(repo, constants.MetricsServer, constants.MetricsServerVersion),
		Arch:  arch,
	}
	return config
}

func (c MetricsServerConfig) Name() string {
	return "metrics-server"
}

func (c MetricsServerConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(MetricsServerTemplate, c)
}
