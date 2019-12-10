package join

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/onecloud-operator/pkg/apis/constants"
)

// JoinData is the interface to use for join phases.
// The "joinData" type from "cmd/join.go" must satisfy this interface.
type JoinData interface {
	phases.JoinData
	OnecloudInitCfg() (*apiv1.InitConfiguration, error)
	OnecloudJoinCfg() *apiv1.JoinConfiguration
	EnabledHostAgent() bool
}

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
	data, ok := c.(JoinData)
	if !ok {
		return errors.New("enable host phase invoked with an invalid data struct")
	}
	if !data.EnabledHostAgent() {
		return nil
	}

	cfg, cli, err := getJoinData(c)
	if err != nil {
		return err
	}
	klog.Infof("Enable host for node %s", cfg.NodeRegistration.Name)
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

func getJoinData(c workflow.RunData) (*apiv1.JoinConfiguration, clientset.Interface, error) {
	data, ok := c.(JoinData)
	if !ok {
		return nil, nil, errors.New("join phase invoked with an invalid data struct")
	}
	cfg := data.OnecloudJoinCfg()
	client, err := data.ClientSet()
	if err != nil {
		return nil, nil, err
	}

	return cfg, client, err
}
