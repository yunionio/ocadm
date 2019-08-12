package cluster

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
	"yunion.io/x/onecloud-operator/pkg/client/clientset/versioned"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	configutil "yunion.io/x/ocadm/pkg/util/config"
)

const (
	DefaultClusterName = "default"
)

type clusterData struct {
	cfg    *apiv1.InitConfiguration
	client versioned.Interface
}

func newClusterData(cmd *cobra.Command, args []string) (*clusterData, error) {
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

	cli, err := NewClusterClient(tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}
	initCfg, err := FetchInitConfiguration(tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}

	return &clusterData{
		cfg:    initCfg,
		client: cli,
	}, nil
}

func NewCmdCreate(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Run this command to create onecloud cluster",
		Run: func(cmd *cobra.Command, args []string) {
			data, err := newClusterData(cmd, args)
			kubeadmutil.CheckErr(err)

			oc, err := CreateCluster(data)
			kubeadmutil.CheckErr(err)

			fmt.Fprintf(out, "Cluster %s created\n", oc.GetName())
		},
		Args: cobra.NoArgs,
	}
	return cmd
}

func NewCmdConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rcadmin",
		Short: "Get climc rc admin auth config",
		Run: func(cmd *cobra.Command, args []string) {
			data, err := newClusterData(cmd, args)
			kubeadmutil.CheckErr(err)

			ret, err := GetClusterRCAdmin(data)
			kubeadmutil.CheckErr(err)

			fmt.Printf("%s\n", ret)
		},
		Args: cobra.NoArgs,
	}
	return cmd
}

func CreateCluster(data *clusterData) (*v1alpha1.OnecloudCluster, error) {
	cli := data.client
	cfg := data.cfg
	ret, err := cli.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(ret.Items) != 0 {
		return nil, errors.Errorf("Cluster already create")
	}
	return cli.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).Create(newCluster(cfg))
}

func newCluster(cfg *apiv1.InitConfiguration) *v1alpha1.OnecloudCluster {
	lbEndpoint := cfg.ControlPlaneEndpoint
	if lbEndpoint != "" {
		lbEndpoint = strings.Split(lbEndpoint, ":")[0]
	}
	if lbEndpoint == "" {
		lbEndpoint = cfg.ManagementNetInterface.IPAddress()
	}
	return &v1alpha1.OnecloudCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.OnecloudNamespace,
			Name:      DefaultClusterName,
		},
		Spec: v1alpha1.OnecloudClusterSpec{
			Mysql: v1alpha1.Mysql{
				Host:     cfg.MysqlConnection.Server,
				Port:     int32(cfg.MysqlConnection.Port),
				Username: cfg.MysqlConnection.Username,
				Password: cfg.MysqlConnection.Password,
			},
			LoadBalancerEndpoint: lbEndpoint,
			ImageRepository:      cfg.ImageRepository,
		},
	}
}

func GetClusterRCAdmin(data *clusterData) (string, error) {
	cli := data.client
	cluster, err := cli.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).Get(DefaultClusterName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	authURL := fmt.Sprintf("https://%s:%d/v3", cluster.Spec.LoadBalancerEndpoint, constants.KeystoneAdminPort)
	passwd := cluster.Spec.Keystone.BootstrapPassword
	return fmt.Sprintf(
		`export OS_AUTH_URL=%s
export OS_USERNAME=sysadmin
export OS_PASSWORD=%s
export OS_PROJECT_NAME=system
export YUNION_INSECURE=true
export OS_REGION_NAME=%s
export OS_ENDPOINT_TYPE=publicURL`, authURL, passwd, cluster.Spec.Region), nil
}

func NewClusterClient(config *clientcmdapi.Config) (*versioned.Clientset, error) {
	clientConfig, err := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create API client configuration from kubeconfig")
	}
	return versioned.NewForConfig(clientConfig)
}

func FetchInitConfiguration(tlsBootstrapCfg *clientcmdapi.Config) (*apiv1.InitConfiguration, error) {
	// creates a client to access the cluster using the bootstrap token identity
	tlsClient, err := kubeconfigutil.ToClientSet(tlsBootstrapCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to access the cluster")
	}
	initconfiguration, err := configutil.FetchInitConfigurationFromCluster(tlsClient, ioutil.Discard, "preflight", true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch the ocadm-config ConfigMap")
	}
	return initconfiguration, nil
}
