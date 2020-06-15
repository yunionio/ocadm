package cmd

import (
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/phases/cluster"
	"yunion.io/x/ocadm/pkg/util/kubectl"

	"yunion.io/x/onecloud-operator/pkg/client/clientset/versioned"
)

type kubectlCli struct {
	kubectlClient  *kubectl.Client
	kubeconfigPath string
	client         versioned.Interface
}

func (d *kubectlCli) KubectlClient() (*kubectl.Client, error) {
	if d.kubectlClient != nil {
		return d.kubectlClient, nil
	}
	cli, err := kubectl.NewClientFormKubeconfigFile(d.KubeConfigPath())
	if err != nil {
		return nil, err
	}
	d.kubectlClient = cli
	return d.kubectlClient, nil
}

// KubeConfigPath returns the path to the kubeconfig file to use for connecting to Kubernetes
func (d *kubectlCli) KubeConfigPath() string {
	return d.kubeconfigPath
}

func (d *kubectlCli) VersionedClient() (versioned.Interface, error) {
	if d.client != nil {
		return d.client, nil
	}

	var tlsBootstrapCfg *clientcmdapi.Config
	var err error

	kubeConfigFile := constants.GetAdminKubeConfigPath()
	if _, err := os.Stat(kubeConfigFile); err == nil {
		// use the admin.conf as tlsBootstrapCfg, that is the kubeconfig file used for reading the ocadm-config during dicovery
		klog.V(1).Infof("[preflight] found %s. Use it for skipping discovery", kubeConfigFile)
		tlsBootstrapCfg, err = clientcmd.LoadFromFile(kubeConfigFile)
		if err != nil {
			return nil, errors.Wrapf(err, "Error loading %s", kubeConfigFile)
		}
	} else {
		return nil, err
	}

	if tlsBootstrapCfg == nil {
		return nil, errors.Errorf("Not found valid %s, please run this command at controlplane", kubeConfigFile)
	}
	cli, err := cluster.NewClusterClient(tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}
	d.client = cli
	return cli, nil
}
