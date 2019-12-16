package baremetal

import (
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	occonstants "yunion.io/x/ocadm/pkg/apis/constants"

	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud-operator/pkg/apis/constants"
	"yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
	"yunion.io/x/onecloud-operator/pkg/client/clientset/versioned"
	"yunion.io/x/onecloud/pkg/baremetal/options"
)

type baremetalEnableData interface {
	ClientSet() (*clientset.Clientset, error)
	GetNodes() []string
	GetListenInterface() string
	VersionedClient() (versioned.Interface, error)
}

func NodesEnableBaremetalAgent() workflow.Phase {
	return workflow.Phase{
		Name:  "enable-baremetal-agent",
		Short: "Enable baremetal agent",
		Phases: []workflow.Phase{
			{
				Name:  "enable-baremetl-agent",
				Short: "Add enable baremetal label to node",
				Run:   batchEnableBaremeatlAgent,
			},
		},
	}
}

func NodesDisableBaremetalAgent() workflow.Phase {
	return workflow.Phase{
		Name:  "disable-baremetal-agent",
		Short: "Disable baremetal agent",
		Phases: []workflow.Phase{
			{
				Name:  "disable-baremetl-agent",
				Short: "Add disable baremetal label to node",
				Run:   batchDisableBaremeatlAgent,
			},
		},
	}
}

func getData(
	c workflow.RunData) ([]string, string, clientset.Interface, versioned.Interface, error) {
	data, ok := c.(baremetalEnableData)
	if !ok {
		return nil, "", nil, nil, errors.New("host enable phase invoked with an invalid data struct")
	}
	cli, err := data.ClientSet()
	if err != nil {
		return nil, "", nil, nil, err
	}
	versiondCli, err := data.VersionedClient()
	if err != nil {
		return nil, "", nil, nil, err
	}

	return data.GetNodes(), data.GetListenInterface(), cli, versiondCli, nil
}

func batchEnableBaremeatlAgent(c workflow.RunData) error {
	nodes, listenInterface, cli, versionedCli, err := getData(c)
	if err != nil {
		return err
	}

	// update baremetal config map
	if len(listenInterface) > 0 {
		ret, err := versionedCli.OnecloudV1alpha1().
			OnecloudClusters(occonstants.OnecloudNamespace).List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "get cluster")
		}
		if len(ret.Items) == 0 {
			return errors.New("Cluster dosen't create ??")
		}
		clusterName := ret.Items[0].Name
		configmapName := fmt.Sprintf("%s-%s", clusterName, v1alpha1.BaremetalAgentComponentType)
		cfgmap, err := cli.CoreV1().ConfigMaps(occonstants.OnecloudNamespace).
			Get(configmapName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "fetch config maps")
		}
		var config = &options.Options
		configStr, ok := cfgmap.Data["config"]
		if !ok {
			config.ListenInterface = listenInterface
		} else {
			jsonobj, err := jsonutils.ParseYAML(configStr)
			if err != nil {
				return errors.Wrap(err, "parse baremetal config")
			}
			err = jsonobj.Unmarshal(config)
			if err != nil {
				return errors.Wrap(err, "unmarshal baremetal config")
			}
		}
		config.ListenInterface = listenInterface
		cfgmap.Data["config"] = jsonutils.Marshal(config).YAMLString()
		_, err = cli.CoreV1().ConfigMaps(occonstants.OnecloudNamespace).Update(cfgmap)
		if err != nil {
			return errors.Wrap(err, "update baremeatl configmap")
		}
	}

	for i := 0; i < len(nodes); i++ {
		klog.Infof("Enable baremetal for node %s", nodes[i])
		node, err := cli.CoreV1().Nodes().Get(nodes[i], metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Node %s enable baremetal failed on get: %s", nodes[i], err)
			continue
		}
		if node.Labels == nil {
			node.Labels = make(map[string]string)
		}
		node.Labels[constants.OnecloudEanbleBaremetalLabelKey] = "enable"
		_, err = cli.CoreV1().Nodes().Update(node)
		if err != nil {
			klog.Errorf("Node %s enable baremetal failed on update: %s", nodes[i], err)
			continue
		}
	}
	klog.Infof("Enable baremetal agent phase finished ...")
	return nil
}

func batchDisableBaremeatlAgent(c workflow.RunData) error {
	nodes, _, cli, _, err := getData(c)
	if err != nil {
		return err
	}
	for i := 0; i < len(nodes); i++ {
		klog.Infof("Disable baremetal for node %s", nodes[i])
		node, err := cli.CoreV1().Nodes().Get(nodes[i], metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Node %s disable baremetal failed on get: %s", nodes[i], err)
			continue
		}
		if node.Labels == nil {
			node.Labels = make(map[string]string)
		}
		node.Labels[constants.OnecloudEanbleBaremetalLabelKey] = "disable"
		_, err = cli.CoreV1().Nodes().Update(node)
		if err != nil {
			klog.Errorf("Node %s disable baremetal failed on update: %s", nodes[i], err)
			continue
		}
	}
	klog.Infof("Disable baremetal agent phase finished ...")
	return nil
}
