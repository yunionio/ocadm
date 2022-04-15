package calico

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

const (
	DefaultVersion = constants.DefaultCalicoVersion
)

type CNICalicoConfig struct {
	ControllerImage       string
	NodeImage             string
	CNIImage              string
	ClusterCIDR           string
	IPAutodetectionMethod string
	FelixChaininsertmode  string
	IPV4PoolBlockSize     int
}

func NewCalicoConfig(cfg *kubeadmapi.ClusterConfiguration, IPAutodetectionMethod, FelixChaininsertmode string, ipv4PoolBlockSize int) addons.Configer {
	if len(FelixChaininsertmode) == 0 {
		FelixChaininsertmode = constants.DefaultCalicoFelixChaininsertmode
	}
	if ipv4PoolBlockSize < 0 {
		ipv4PoolBlockSize = 26
	}
	repo := cfg.ImageRepository
	config := &CNICalicoConfig{
		ControllerImage:       images.GetGenericImage(repo, constants.CalicoKubeControllers, DefaultVersion),
		NodeImage:             images.GetGenericImage(repo, constants.CalicoNode, DefaultVersion),
		CNIImage:              images.GetGenericImage(repo, constants.CalicoCNI, DefaultVersion),
		ClusterCIDR:           cfg.Networking.PodSubnet,
		IPAutodetectionMethod: IPAutodetectionMethod,
		FelixChaininsertmode:  FelixChaininsertmode,
		IPV4PoolBlockSize:     ipv4PoolBlockSize,
	}
	return config
}

func (c CNICalicoConfig) Name() string {
	return "calico"
}

func (c CNICalicoConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(CNICalicoTemplate, c)
}
