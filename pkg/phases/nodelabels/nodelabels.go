package nodelabels

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

type nodesSetLablesData interface {
	ClientSet() (*clientset.Clientset, error)
	GetNodes() []string
	GetLabels() map[string]string
}

func NodesSetLabels() workflow.Phase {
	return workflow.Phase{
		Name:  "nodes-set-labels",
		Short: "Nodes set labels",
		Phases: []workflow.Phase{
			{
				Name:  "nodes-set-labels",
				Short: "Nodes set lables",
				Run:   nodesSetLabels,
			},
		},
	}
}

func getData(c workflow.RunData,
) ([]string, clientset.Interface, map[string]string, error) {
	data, ok := c.(nodesSetLablesData)
	if !ok {
		return nil, nil, nil, errors.New("nodes set labels phase invoked with an invalid data struct")
	}
	cli, err := data.ClientSet()
	if err != nil {
		return nil, nil, nil, err
	}
	return data.GetNodes(), cli, data.GetLabels(), nil
}

func nodesSetLabels(c workflow.RunData) error {
	nodes, cli, labels, err := getData(c)
	if err != nil {
		return err
	}
	for i := 0; i < len(nodes); i++ {
		klog.Infof("Set labels for node %s", nodes[i])
		node, err := cli.CoreV1().Nodes().Get(nodes[i], metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Node %s set labels failed on get: %s", nodes[i], err)
			continue
		}
		if node.Labels == nil {
			node.Labels = make(map[string]string)
		}
		for k, v := range labels {
			node.Labels[k] = v
		}
		_, err = cli.CoreV1().Nodes().Update(node)
		if err != nil {
			klog.Errorf("Node %s set labels failed on update: %s", nodes[i], err)
			continue
		}
	}
	klog.Info("Set nodes lables phase finished ...")
	return nil
}
