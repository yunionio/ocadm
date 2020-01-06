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

func GetOnecloudImage(image string, cfg *v1.ClusterConfiguration, kubeadmCfg *kubeadmapi.ClusterConfiguration) string {
	repoPrefix := kubeadmCfg.ImageRepository
	onecloudImageTag := cfg.OnecloudVersion
	return GetGenericImage(repoPrefix, image, onecloudImageTag)
}

// GetAllImages returns a list of container images expects to use on a control plane node
func GetAllImages(cfg *v1.ClusterConfiguration, kubeadmCfg *kubeadmapi.ClusterConfiguration, operatorVersion string) []string {
	imgs := images.GetControlPlaneImages(kubeadmCfg)
	//for _, component := range []string{
	//constants.OnecloudOperator,
	//} {
	//imgs = append(imgs, GetOnecloudImage(component, cfg, kubeadmCfg))
	//}
	repoPrefix := kubeadmCfg.ImageRepository
	for img, version := range map[string]string{
		constants.CalicoKubeControllers:       constants.DefaultCalicoVersion,
		constants.CalicoNode:                  constants.DefaultCalicoVersion,
		constants.CalicoCNI:                   constants.DefaultCalicoVersion,
		constants.RancherLocalPathProvisioner: constants.DefaultLocalProvisionerVersion,
		constants.IngressControllerTraefik:    constants.DefaultTraefikVersion,
		constants.OnecloudOperator:            operatorVersion,
	} {
		imgs = append(imgs, GetGenericImage(repoPrefix, img, version))
	}
	return imgs
}
