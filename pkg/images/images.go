package images

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/apis/v1"
)

var (
	GetGenericImage = images.GetGenericImage
)

func GetOnecloudImage(image string, cfg *v1.ClusterConfiguration) string {
	repoPrefix := cfg.ImageRepository
	onecloudImageTag := cfg.OnecloudVersion
	return GetGenericImage(repoPrefix, image, onecloudImageTag)
}

// GetAllImages returns a list of container images expects to use on a control plane node
func GetAllImages(cfg *v1.ClusterConfiguration, kubeadmCfg *kubeadmapi.ClusterConfiguration) []string {
	imgs := images.GetControlPlaneImages(kubeadmCfg)
	for _, component := range []string{
		constants.OnecloudOperator,
	} {
		imgs = append(imgs, GetOnecloudImage(component, cfg))
	}
	repoPrefix := cfg.ImageRepository
	imgs = append(imgs, GetGenericImage(repoPrefix, constants.RancherLocalPathProvisioner, constants.DefaultLocalProvisionerVersion))
	imgs = append(imgs, GetGenericImage(repoPrefix, constants.IngressControllerTraefik, constants.DefaultTraefikVersin))
	return imgs
}
