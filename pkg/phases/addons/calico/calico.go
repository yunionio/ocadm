package calico

import (
	"fmt"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
	"yunion.io/x/ocadm/pkg/util/kubectl"
)

const (
	DefaultVersion = constants.DefaultCalicoVersion
)

type CNICalicoConfig struct {
	ControllerImage string
	NodeImage       string
	CNIImage        string
	ClusterCIDR     string
}

func (c CNICalicoConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(CNICalicoTemplate, c)
}

// EnsureCalicoAddon creates the calico cni
func EnsureCalicoAddon(cfg *kubeadmapi.ClusterConfiguration, kubectlCli *kubectl.Client) error {
	repo := cfg.ImageRepository
	config := CNICalicoConfig{
		ControllerImage: images.GetGenericImage(repo, constants.CalicoKubeControllers, DefaultVersion),
		NodeImage:       images.GetGenericImage(repo, constants.CalicoNode, DefaultVersion),
		CNIImage:        images.GetGenericImage(repo, constants.CalicoCNI, DefaultVersion),
		ClusterCIDR:     cfg.Networking.PodSubnet,
	}
	manifest, err := config.GenerateYAML()
	if err != nil {
		return err
	}
	if err := kubectlCli.Apply(manifest); err != nil {
		return err
	}
	fmt.Println("[oc-addons] Applied essential addon: calico")
	return nil
}
