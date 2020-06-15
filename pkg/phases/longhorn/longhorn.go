package longhorn

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	"yunion.io/x/ocadm/pkg/apis/constants"

	"yunion.io/x/ocadm/pkg/phases/addons"
	"yunion.io/x/ocadm/pkg/phases/addons/longhorn"
	"yunion.io/x/ocadm/pkg/util/kubectl"
)

func InstallLonghornPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "install-longhorn",
		Short: "install longhorn",
		Phases: []workflow.Phase{
			{
				Name:  "install-longhorn",
				Short: "install longhorn",
				Run:   runLonghornAddon,
			},
		},
	}
}

type LonghornData interface {
	KubectlClient() (*kubectl.Client, error)
	GetImageRepository() string
	LonghornConfig() *LonghornConfig
	ClientSet() (*clientset.Clientset, error)
	GetNodes() []string
}

type LonghornConfig struct {
	DataPath                    string
	OverProviosioningPercentage int
	ReplicaCount                int
}

func runLonghornAddon(c workflow.RunData) error {
	data, ok := c.(LonghornData)
	if !ok {
		return errors.New("addon phase invoked with an invalid data struct")
	}
	kubectlCli, err := data.KubectlClient()
	if err != nil {
		return err
	}
	loghornConfig := data.LonghornConfig()
	cli, err := data.ClientSet()
	if err != nil {
		return err
	}

	loghornConfig.ReplicaCount, err = lableLonghornNodes(cli, data.GetNodes())
	if err != nil {
		return err
	}
	configer := longhorn.NewLonghornConfig(
		data.GetImageRepository(), loghornConfig.DataPath,
		loghornConfig.OverProviosioningPercentage, loghornConfig.ReplicaCount,
	)
	return addons.KubectlApplyAddon(configer, kubectlCli, false)
}

func lableLonghornNodes(cli *clientset.Clientset, nodes []string) (int, error) {
	var replicaCount int
	if len(nodes) == 0 {
		nodelist, err := cli.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return 0, err
		}
		if len(nodelist.Items) < 3 {
			replicaCount = 1
		} else {
			replicaCount = 3
		}
		for i := 0; i < len(nodelist.Items); i++ {
			if nodelist.Items[i].Labels == nil {
				nodelist.Items[i].Labels = make(map[string]string)
			}
			nodelist.Items[i].Labels[constants.LonghornCreateDiskLable] = "true"
			_, err = cli.CoreV1().Nodes().Update(&nodelist.Items[i])
			if err != nil {
				klog.Errorf("Node %s enable default disk failed on update: %s", nodes[i], err)
				continue
			}
			klog.Infof("Enable default disk for node %s", nodelist.Items[i].Name)
		}
	} else {
		if len(nodes) < 3 {
			replicaCount = 1
		} else {
			replicaCount = 3
		}
		for i := 0; i < len(nodes); i++ {
			node, err := cli.CoreV1().Nodes().Get(nodes[i], metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Node %s enable default disk failed on get: %s", nodes[i], err)
				continue
			}
			if node.Labels == nil {
				node.Labels = make(map[string]string)
			}
			node.Labels[constants.LonghornCreateDiskLable] = "true"
			_, err = cli.CoreV1().Nodes().Update(node)
			if err != nil {
				klog.Errorf("Node %s enable default disk failed on update: %s", nodes[i], err)
				continue
			}
			klog.Infof("Enable default disk for node %s", nodes[i])
		}
	}
	return replicaCount, nil
}
