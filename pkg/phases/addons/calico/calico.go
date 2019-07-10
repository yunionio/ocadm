package calico

import (
	"fmt"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
	"yunion.io/x/ocadm/pkg/util/kubectl"
)

const (
	DefaultVersion = "v3.7.2"
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
		ControllerImage: images.GetGenericImage(repo, "calico-kube-controllers", DefaultVersion),
		NodeImage:       images.GetGenericImage(repo, "calico-node", DefaultVersion),
		CNIImage:        images.GetGenericImage(repo, "calico-cni", DefaultVersion),
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
