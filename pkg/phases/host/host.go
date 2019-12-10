package host

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/onecloud-operator/pkg/apis/constants"
)

type hostEnableData interface {
	ClientSet() (*clientset.Clientset, error)
	GetNodes() []string
}

func NodesEnableHostAgent() workflow.Phase {
	return workflow.Phase{
		Name:  "enable-host-agent",
		Short: "Enable host agent",
		Phases: []workflow.Phase{
			{
				Name:  "enable-host-agent",
				Short: "Add enable host label to node",
				Run:   batchEnableHostAgent,
			},
		},
	}
}

func getData(c workflow.RunData) ([]string, clientset.Interface, error) {
	data, ok := c.(hostEnableData)
	if !ok {
		return nil, nil, errors.New("host enable phase invoked with an invalid data struct")
	}
	cli, err := data.ClientSet()
	if err != nil {
		return nil, nil, err
	}
	return data.GetNodes(), cli, nil
}

func batchEnableHostAgent(c workflow.RunData) error {
	nodes, cli, err := getData(c)
	if err != nil {
		return err
	}
	for i := 0; i < len(nodes); i++ {
		klog.Infof("Enable host for node %s", nodes[i])
		node, err := cli.CoreV1().Nodes().Get(nodes[i], metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Node %s enable host failed on get: %s", nodes[i], err)
			continue
		}
		if node.Labels == nil {
			node.Labels = make(map[string]string)
		}
		node.Labels[constants.OnecloudEnableHostLabelKey] = "enable"
		_, err = cli.CoreV1().Nodes().Update(node)
		if err != nil {
			klog.Errorf("Node %s enable host failed on update: %s", nodes[i], err)
			continue
		}
	}
	return nil
}
