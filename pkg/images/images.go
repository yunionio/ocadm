package images

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/apis/v1"
)

func GetOnecloudImage(image string, cfg *v1.ClusterConfiguration) string {
	repoPrefix := cfg.ImageRepository
	onecloudImageTag := cfg.OnecloudVersion
	return images.GetGenericImage(repoPrefix, image, onecloudImageTag)
}

// GetAllImages returns a list of container images expects to use on a control plane node
func GetAllImages(cfg *v1.ClusterConfiguration, kubeadmCfg *kubeadmapi.ClusterConfiguration) []string {
	imgs := images.GetAllImages(kubeadmCfg)
	for _, component := range []string{
		constants.OnecloudKeystone,
		constants.OnecloudRegion,
		constants.OnecloudScheduler,
		constants.OnecloudGlance,
		constants.OnecloudBaremetal,
	} {
		imgs = append(imgs, GetOnecloudImage(component, cfg))
	}
	return imgs
}
