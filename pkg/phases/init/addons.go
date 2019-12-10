package init

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	"k8s.io/kubernetes/pkg/util/normalizer"

	"yunion.io/x/onecloud-operator/pkg/apis/constants"

	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	calicoaddon "yunion.io/x/ocadm/pkg/phases/addons/calico"
	ocaddon "yunion.io/x/ocadm/pkg/phases/addons/onecloudoperator"
	"yunion.io/x/ocadm/pkg/util/kubectl"
)

var (
	CalicoCNIAddonLongDesc = normalizer.LongDesc(`
	Installs the calico cni addon components via the API server.
	`)
)

func NodeEnableHostAgent() workflow.Phase {
	return workflow.Phase{
		Name:  "enable-host-agent",
		Short: "Enable host agent",
		Phases: []workflow.Phase{
			{
				Name:  "enable-host-agent",
				Short: "Add enable host label to node",
				Run:   enableHostAgent,
			},
		},
	}
}

func enableHostAgent(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("addon phase invoked with an invalid data struct")
	}
	if !data.EnabledHostAgent() {
		return nil
	}

	cfg, cli, _, err := getInitData(c)
	if err != nil {
		return err
	}
	klog.Infof("[enable-host] Enable host for node %s", cfg.NodeRegistration.Name)
	node, err := cli.CoreV1().Nodes().Get(cfg.NodeRegistration.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "get node")
	}
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	node.Labels[constants.OnecloudEnableHostLabelKey] = "enable"
	_, err = cli.CoreV1().Nodes().Update(node)
	if err != nil {
		return errors.Wrap(err, "update node")
	}
	return nil
}

// NewAddonPhase returns the addon Cobra command
func NewOCAddonPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "oc-addon",
		Short: "Installs onecloud required addons to kubernetes cluster",
		Phases: []workflow.Phase{
			{
				Name:           "all",
				Short:          "Installs all the addons",
				InheritFlags:   getAddonPhaseFlags("all"),
				RunAllSiblings: true,
			},
			{
				Name:  "calico",
				Short: "Install the calico cni addon to a Kubernetes cluster",
				Long:  CalicoCNIAddonLongDesc,
				Run:   runCalicoAddon,
			},
			{
				Name:  "onecloud-operator",
				Short: "Install the onecloud operator addon",
				Run:   runOCOperatorAddon,
			},
		},
	}
}

func getInitData(c workflow.RunData) (*apiv1.InitConfiguration, clientset.Interface, *kubectl.Client, error) {
	data, ok := c.(InitData)
	if !ok {
		return nil, nil, nil, errors.New("addon phase invoked with an invalid data struct")
	}
	cfg := data.OnecloudCfg()
	client, err := data.Client()
	if err != nil {
		return nil, nil, nil, err
	}
	ctlCli, err := data.KubectlClient()
	if err != nil {
		return nil, nil, nil, err
	}
	return cfg, client, ctlCli, err
}

func runCalicoAddon(c workflow.RunData) error {
	cfg, _, kubectlCli, err := getInitData(c)
	if err != nil {
		return err
	}
	return calicoaddon.EnsureCalicoAddon(&cfg.InitConfiguration.ClusterConfiguration, kubectlCli)
}

func runOCOperatorAddon(c workflow.RunData) error {
	cfg, _, kubectlCli, err := getInitData(c)
	if err != nil {
		return err
	}
	for _, f := range []func(*kubeadmapi.ClusterConfiguration, *kubectl.Client) error{
		ocaddon.EnsureOnecloudOperatorAddon,
		ocaddon.EnsureLocalPathProvisionerAddon,
		ocaddon.EnsureIngressTraefikAddon,
	} {
		if err := f(&cfg.InitConfiguration.ClusterConfiguration, kubectlCli); err != nil {
			return err
		}
	}
	return nil
}

func getAddonPhaseFlags(name string) []string {
	flags := []string{
		options.CfgPath,
		options.KubeconfigPath,
		options.KubernetesVersion,
		options.ImageRepository,
	}
	if name == "all" || name == "calico" {
		flags = append(flags,
			options.NetworkingPodSubnet,
		)
	}
	return flags
}
