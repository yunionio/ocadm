package cluster

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	operatorconstants "yunion.io/x/onecloud-operator/pkg/apis/constants"
	"yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
	"yunion.io/x/onecloud-operator/pkg/client/clientset/versioned"
	occonfig "yunion.io/x/onecloud-operator/pkg/manager/config"
	"yunion.io/x/onecloud-operator/pkg/util/image"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/apis/scheme"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/kube"
	ocutil "yunion.io/x/ocadm/pkg/util/onecloud"
)

const (
	DefaultClusterName  = "default"
	DefaultOperatorName = "onecloud-operator"
	// Annotations of CE autoupdate component
	AutoUpdateCurrentVersion = "autoupdate.onecloud.yunion.io/current-version"
)

type clusterData struct {
	cfg        *apiv1.InitConfiguration
	client     versioned.Interface
	k8sClient  kubernetes.Interface
	kubeClient *kube.Client
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
	k8sCli, initCfg, err := FetchInitConfiguration(tlsBootstrapCfg)
	if err != nil {
		return nil, err
	}
	kubeCli, err := kube.NewClientByFile(kubeConfigFile)
	if err != nil {
		return nil, err
	}
	return &clusterData{
		cfg:        initCfg,
		client:     cli,
		k8sClient:  k8sCli,
		kubeClient: kubeCli,
	}, nil
}

func (data *clusterData) GetDefaultCluster() (*v1alpha1.OnecloudCluster, error) {
	cli := data.client
	ret, err := cli.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).Get(DefaultClusterName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (data *clusterData) GetOperator() (*appv1.Deployment, error) {
	cli := data.k8sClient
	return cli.AppsV1().Deployments(constants.OnecloudNamespace).Get(DefaultOperatorName, metav1.GetOptions{})
}

type createOptions struct {
	useEE       bool
	version     string
	wait        bool
	useLonghorn bool

	// cluster upgrade from onecloud 2.x
	region        string
	zone          string
	upgradeFromV2 bool
}

func newCreateOptions() *createOptions {
	return &createOptions{}
}

func NewCmdCreate(out io.Writer) *cobra.Command {
	opt := newCreateOptions()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Run this command to create onecloud cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if opt.upgradeFromV2 {
				if len(opt.zone) == 0 {
					kubeadmutil.CheckErr(errors.New("missing onecloud cluster zone id "))
				}
				if len(opt.region) == 0 {
					kubeadmutil.CheckErr(errors.New("misssing onecloud cluster region id"))
				}
			}

			data, err := newClusterData(cmd, args)
			kubeadmutil.CheckErr(err)

			oc, err := CreateCluster(data, opt)
			kubeadmutil.CheckErr(err)

			fmt.Fprintf(out, "Cluster %s created\n", oc.GetName())
		},
		Args: cobra.NoArgs,
	}
	AddCreateOptions(cmd.Flags(), opt)
	return cmd
}

func AddCreateOptions(flagSet *flag.FlagSet, opt *createOptions) {
	flagSet.BoolVar(&opt.useEE, "use-ee", opt.useEE, "Use EE edition")
	flagSet.StringVar(&opt.version, "version", opt.version, "onecloud cluster version")
	flagSet.BoolVar(&opt.wait, "wait", opt.wait, "wait until workload created")
	flagSet.StringVar(&opt.region, "cluster-region-id", "", "For upgrade from v2, onecloud cluster region id, climc region-list get region ids")
	flagSet.StringVar(&opt.zone, "cluster-zone-id", "", "For upgrade from v2, onecloud cluster zone id, climc zone-list get zone ids")
	flagSet.BoolVar(&opt.upgradeFromV2, "upgrade-from-v2", opt.upgradeFromV2, "cluster upgrade from onecloud 2.x")
	flagSet.BoolVar(&opt.useLonghorn, "use-longhorn", opt.useLonghorn, "Use longhorn as glanc and influxdb storage class, but you should enable longhorn by `ocadm longhorn enable` at first.")
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

func CreateCluster(data *clusterData, opt *createOptions) (*v1alpha1.OnecloudCluster, error) {
	cli := data.client
	cfg := data.cfg
	ret, err := cli.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "list onecloud cluster")
	}
	if len(ret.Items) != 0 {
		oc := ret.Items[0]
		if opt.wait {
			if err := ocutil.WaitOnecloudDeploymentUpdated(data.client, oc.GetName(), oc.GetNamespace(), 30*time.Minute, nil); err != nil {
				return &oc, errors.Wrap(err, "wait onecloud cluster services running")
			}
		}
		return &oc, nil
	}
	var cluster *v1alpha1.OnecloudCluster
	if opt.upgradeFromV2 {
		cluster, err = newClusterConfig(data.k8sClient, cfg, opt)
		if err != nil {
			return nil, errors.Wrap(err, "take out cluster config")
		}
	} else {
		cluster = newCluster(cfg, opt)
		if opt.useLonghorn {
			cluster.Spec.Glance.StorageClassName = constants.LonghornStorageClass
			cluster.Spec.Influxdb.StorageClassName = constants.LonghornStorageClass
			cluster.Spec.Meter.StorageClassName = constants.LonghornStorageClass
			cluster.Spec.BaremetalAgent.StorageClassName = constants.LonghornStorageClass
			cluster.Spec.EsxiAgent.StorageClassName = constants.LonghornStorageClass
		}
	}
	oc, err := cli.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).Create(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "create cluster")
	}
	if opt.wait {
		if err := ocutil.WaitOnecloudDeploymentUpdated(data.client, oc.GetName(), oc.GetNamespace(), 30*time.Minute, nil); err != nil {
			return oc, errors.Wrap(err, "wait onecloud cluster services running")
		}
	}
	return oc, nil
}

func newCluster(cfg *apiv1.InitConfiguration, opt *createOptions) *v1alpha1.OnecloudCluster {
	lbEndpoint := cfg.ControlPlaneEndpoint
	if lbEndpoint != "" {
		lbEndpoint = strings.Split(lbEndpoint, ":")[0]
	}
	if lbEndpoint == "" {
		lbEndpoint = cfg.ManagementNetInterface.IPAddress()
	}
	oc := &v1alpha1.OnecloudCluster{
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
			Version:              cfg.OnecloudVersion,
			Region:               cfg.Region,
		},
	}
	if opt.version != "" {
		oc.Spec.Version = opt.version
	}
	if opt.useEE {
		ocutil.SetOCUseEE(oc)
	} else {
		ocutil.SetOCUseCE(oc)
	}
	return oc
}

func newCluster2(env map[string]string, cfg *apiv1.InitConfiguration, opt *createOptions) (*v1alpha1.OnecloudCluster, error) {
	mysqlPort, err := strconv.Atoi(env["MYSQL_PORT"])
	if err != nil {
		return nil, errors.Wrap(err, "parse mysql port")
	}
	oc := &v1alpha1.OnecloudCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OnecloudCluster",
			APIVersion: "onecloud.yunion.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.OnecloudNamespace,
			Name:      DefaultClusterName,
		},
		Spec: v1alpha1.OnecloudClusterSpec{
			Mysql: v1alpha1.Mysql{
				Host:     env["MYSQL_HOST"],
				Port:     int32(mysqlPort),
				Username: "root",
				Password: env["MYSQL_ROOT_PASSWORD"],
			},
			LoadBalancerEndpoint: env["MANAGEMENT_IP"],
			ImageRepository:      cfg.ImageRepository,
			Version:              cfg.OnecloudVersion,
			Region:               opt.region,
			Zone:                 opt.zone,
			Keystone: v1alpha1.KeystoneSpec{
				BootstrapPassword: env["SYSADMIN_PASSWORD"],
			},
			Glance: v1alpha1.StatefulDeploymentSpec{
				DeploymentSpec: v1alpha1.DeploymentSpec{
					NodeSelector: map[string]string{
						"onecloud.yunion.io/glance": "enable",
					},
				},
			},
			BaremetalAgent: v1alpha1.ZoneStatefulDeploymentSpec{
				StatefulDeploymentSpec: v1alpha1.StatefulDeploymentSpec{
					DeploymentSpec: v1alpha1.DeploymentSpec{
						NodeSelector: map[string]string{
							"onecloud.yunion.io/baremetal": "enable",
						},
					},
				},
			},
			EsxiAgent: v1alpha1.ZoneStatefulDeploymentSpec{
				StatefulDeploymentSpec: v1alpha1.StatefulDeploymentSpec{
					DeploymentSpec: v1alpha1.DeploymentSpec{
						NodeSelector: map[string]string{
							"onecloud.yunion.io/esxi": "enable",
						},
					},
				},
			},
		},
		Status: v1alpha1.OnecloudClusterStatus{
			RegionServer: v1alpha1.RegionStatus{
				RegionId:     opt.region,
				RegionZoneId: opt.region,
				ZoneId:       opt.zone,
			},
		},
	}

	if opt.version != "" {
		oc.Spec.Version = opt.version
	}
	if opt.useEE {
		ocutil.SetOCUseEE(oc)
	} else {
		ocutil.SetOCUseCE(oc)
	}
	return oc, nil
}

func newClusterConfig(cli kubernetes.Interface, cfg *apiv1.InitConfiguration, opt *createOptions) (*v1alpha1.OnecloudCluster, error) {
	env, err := godotenv.Read("/opt/cloud/workspace/globalrc", "/opt/yunionsetup/vars")
	if err != nil {
		return nil, errors.Wrap(err, "failed load onecloud config")
	}
	defaultClusterConfigmap := generateClusterConfigmap(env)
	cfgMap := new(corev1.ConfigMap)
	err = runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), []byte(defaultClusterConfigmap), cfgMap)
	if err != nil {
		return nil, errors.Wrap(err, "decode configmap")
	}
	_, err = cli.CoreV1().ConfigMaps(cfgMap.Namespace).Create(cfgMap)
	if err != nil {
		return nil, errors.Wrap(err, "create configmap")
	}
	return newCluster2(env, cfg, opt)
}

func generateClusterConfigmap(cfg map[string]string) string {
	return fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: default-cluster-config
  namespace: onecloud
data:
  OnecloudClusterConfig: |
    apiVersion: onecloud.yunion.io/v1alpha1
    kind: OnecloudClusterConfig
    apiGateway:
      username: %s
      password: %s
    glance:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    keystone:
      db:
        database: %s
        username: %s
        password: %s
    kubeserver:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    logger:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    notify:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    region:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    webconsole:
      username: %s
      password: %s
    yunionagent:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    yunionconf:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    ansibleserver:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    cloudevent:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    cloudnet:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    meter:
      db:
        database: %s
        username: %s
        password: %s
      username: %s
      password: %s
    baremetal:
      username: %s
      password: %s
    esxiagent:
      username: %s
      password: %s
    host:
      username: %s
      password: %s
    itsm:
      db:
        database: %s
        password: %s
        username: %s
      encryptionKey: %s
      password: %s
      username: %s
`,
		// yunionapi
		cfg["YUNIONAPI_ADMIN_USER"], cfg["YUNIONAPI_ADMIN_PASS"],
		// glance
		cfg["MYSQL_DB_GLANCE"], cfg["MYSQL_USER_GLANCE"], cfg["MYSQL_PASS_GLANCE"], cfg["GLANCE_ADMIN_USER"], cfg["GLANCE_ADMIN_PASS"],
		// keystone
		cfg["MYSQL_DB_KEYSTONE"], cfg["MYSQL_USER_KEYSTONE"], cfg["MYSQL_PASS_KEYSTONE"],
		// kube server
		cfg["MYSQL_DB_KUBE"], cfg["MYSQL_USER_KUBE"], cfg["MYSQL_PASS_KUBE"], cfg["YUNION_KUBE_SERVER_ADMIN_USER"], cfg["YUNION_KUBE_SERVER_ADMIN_PASS"],
		// logger
		cfg["MYSQL_DB_LOGGER"], cfg["MYSQL_USER_LOGGER"], cfg["MYSQL_PASS_LOGGER"], cfg["LOGGER_ADMIN_USER"], cfg["LOGGER_ADMIN_PASS"],
		// notify
		cfg["MYSQL_DB_NOTIFY"], cfg["MYSQL_USER_NOTIFY"], cfg["MYSQL_PASS_NOTIFY"], cfg["YUNION_NOTIFY_DOCKER_USER"], cfg["YUNION_NOTIFY_DOCKER_PSWD"],
		// region
		cfg["MYSQL_DB_REGION"], cfg["MYSQL_USER_REGION"], cfg["MYSQL_PASS_REGION"], cfg["REGION_ADMIN_USER"], cfg["REGION_ADMIN_PASS"],
		// webconsole
		cfg["YUNION_WEBCONSOLE_ADMIN_USER"], cfg["YUNION_WEBCONSOLE_ADMIN_PASS"],
		// yunionagent
		cfg["MYSQL_DB_YUNIONAGENT"], cfg["MYSQL_USER_YUNIONAGENT"], cfg["MYSQL_PASS_YUNIONAGENT"], cfg["YUNIONAGENT_ADMIN_USER"], cfg["YUNIONAGENT_ADMIN_PASS"],
		// yunionconf
		cfg["MYSQL_DB_YUNIONCONF"], cfg["MYSQL_USER_YUNIONCONF"], cfg["MYSQL_PASS_YUNIONCONF"], cfg["YUNIONCONF_ADMIN_USER"], cfg["YUNIONCONF_ADMIN_PASS"],
		// ansibleserver
		cfg["MYSQL_DB_ANSIBLESERVER"], cfg["MYSQL_USER_ANSIBLESERVER"], cfg["MYSQL_PASS_ANSIBLESERVER"], cfg["ANSIBLESERVER_ADMIN_USER"], cfg["ANSIBLESERVER_ADMIN_PASS"],
		// cloudevent
		cfg["MYSQL_DB_CLOUDEVENT"], cfg["MYSQL_USER_CLOUDEVENT"], cfg["MYSQL_PASS_CLOUDEVENT"], cfg["CLOUDEVENT_ADMIN_USER"], cfg["CLOUDEVENT_ADMIN_PASS"],
		// cloudnet
		cfg["MYSQL_DB_CLOUDNET"], cfg["MYSQL_USER_CLOUDNET"], cfg["MYSQL_PASS_CLOUDNET"], cfg["CLOUDNET_ADMIN_USER"], cfg["CLOUDNET_ADMIN_PASS"],
		// meter
		cfg["MYSQL_DB_METER"], cfg["MYSQL_USER_METER"], cfg["MYSQL_PASS_METER"], cfg["YUNION_METER_DOCKER_USER"], cfg["YUNION_METER_DOCKER_PSWD"],
		// baremetal
		cfg["BAREMETAL_ADMIN_USER"], cfg["BAREMETAL_ADMIN_PASS"],
		// esxiagent
		cfg["ESXIAGENT_ADMIN_USER"], cfg["ESXIAGENT_ADMIN_PASS"],
		// host
		cfg["HOST_ADMIN_USER"], cfg["HOST_ADMIN_PASS"],
		// itsm
		cfg["MYSQL_DB_ITSM_SECONDARY"], cfg["MYSQL_PASS_ITSM"], cfg["MYSQL_DB_ITSM_SECONDARY"],
		cfg["ITSM_ENCRYPTION_KEY"], cfg["YUNION_ITSM_DOCKER_PSWD"], cfg["YUNION_ITSM_DOCKER_USER"],
	)
}

func GetClusterRCAdmin(data *clusterData) (string, error) {
	cli := data.client
	cluster, err := cli.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).Get(DefaultClusterName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	cfg, err := occonfig.GetClusterConfigByClient(data.k8sClient, cluster)
	if err != nil {
		return "", err
	}
	keystonePort := cfg.Keystone.Port
	authURL := fmt.Sprintf("https://%s:%d/v3", cluster.Spec.LoadBalancerEndpoint, keystonePort)
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

func FetchInitConfiguration(tlsBootstrapCfg *clientcmdapi.Config) (*kubernetes.Clientset, *apiv1.InitConfiguration, error) {
	// creates a client to access the cluster using the bootstrap token identity
	tlsClient, err := kubeconfigutil.ToClientSet(tlsBootstrapCfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to access the cluster")
	}
	initconfiguration, err := configutil.FetchInitConfigurationFromCluster(tlsClient, ioutil.Discard, "preflight", true)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to fetch the ocadm-config ConfigMap")
	}
	return tlsClient, initconfiguration, nil
}

type updateOptions struct {
	version         string
	operatorVersion string
	imageRepository string
	wait            bool
	useEE           bool
	useCE           bool
}

func newUpdateOptions() *updateOptions {
	return &updateOptions{}
}

func NewCmdUpdate(out io.Writer) *cobra.Command {
	opt := newUpdateOptions()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Run this command to update onecloud cluster",
		Run: func(cmd *cobra.Command, args []string) {
			data, err := newClusterData(cmd, args)
			kubeadmutil.CheckErr(err)
			err = updateCluster(data, opt)
			kubeadmutil.CheckErr(err)
		},
		Args: cobra.NoArgs,
	}
	AddUpdateOptions(cmd.Flags(), opt)
	return cmd
}

func AddUpdateOptions(flagSet *flag.FlagSet, opt *updateOptions) {
	flagSet.StringVar(&opt.version, "version", opt.version, "onecloud cluster version")
	flagSet.StringVar(&opt.operatorVersion, options.OperatorVersion, opt.operatorVersion, "onecloud operator version")
	flagSet.StringVar(&opt.imageRepository, "image-repository", opt.imageRepository, "image registry repo")
	flagSet.BoolVar(&opt.wait, "wait", opt.wait, "wait until workload updated")
	flagSet.BoolVar(&opt.useEE, "use-ee", opt.useEE, "use enterprise edition onecloud")
	flagSet.BoolVar(&opt.useCE, "use-ce", opt.useCE, "use community edition onecloud")
}

func updateCluster(data *clusterData, opt *updateOptions) error {
	operator, err := data.GetOperator()
	if err != nil {
		return errors.Wrap(err, "get onecloud operator")
	}
	ref, err := getOperatorImage(operator)
	if err != nil {
		return errors.Wrap(err, "get operator image reference")
	}
	reg := ref.Repository
	imgName := ref.Image
	version := ref.Tag
	digest := ref.Digest
	if opt.operatorVersion != "" {
		if opt.operatorVersion != version {
			version = opt.operatorVersion
		}
	}
	if opt.imageRepository != "" {
		if opt.imageRepository != reg {
			reg = opt.imageRepository
		}
	}
	imgStr := fmt.Sprintf("%s/%s:%s", reg, imgName, version)
	if version == "" && digest != "" {
		imgStr = fmt.Sprintf("%s/%s@%s", reg, imgName, digest)
	}
	operator.Spec.Template.Spec.Containers[0].Image = imgStr
	if _, err := data.k8sClient.AppsV1().Deployments(constants.OnecloudNamespace).Update(operator); err != nil {
		return errors.Wrap(err, "update operator")
	}
	if opt.wait {
		rollout, err := data.kubeClient.Rollout()
		if err != nil {
			return errors.Wrap(err, "get rollout cmd")
		}
		if err := rollout.Status(0).
			SetNamespace(constants.OnecloudNamespace).
			RunDeployment(operator.GetName()); err != nil {
			return err
		}
	}

	oc, err := data.GetDefaultCluster()
	if err != nil {
		return errors.Wrap(err, "get default onecloud cluster")
	}
	updateOC := false
	if oc.Annotations == nil {
		oc.Annotations = make(map[string]string)
	}
	edition := oc.Annotations[operatorconstants.OnecloudEditionAnnotationKey]
	if opt.useEE && edition != operatorconstants.OnecloudEnterpriseEdition {
		oc.Annotations[operatorconstants.OnecloudEditionAnnotationKey] = operatorconstants.OnecloudEnterpriseEdition
		updateOC = true
	} else if opt.useCE && edition != operatorconstants.OnecloudCommunityEdition {
		oc.Annotations[operatorconstants.OnecloudEditionAnnotationKey] = operatorconstants.OnecloudCommunityEdition
		updateOC = true
	}
	if opt.version != "" {
		if opt.version != oc.Spec.Version {
			oc.Spec.Version = opt.version
		}
		updateOC = true
	}
	if opt.imageRepository != "" {
		if opt.imageRepository != oc.Spec.ImageRepository {
			oc.Spec.ImageRepository = opt.imageRepository
		}
		updateOC = true
	}
	// remove autoupdate related annotation
	{
		delete(oc.Annotations, AutoUpdateCurrentVersion)
		oc.Spec.AutoUpdate.Tag = ""
		oc.Spec.HostAgent.Tag = ""
	}
	if updateOC {
		if _, err := data.client.OnecloudV1alpha1().OnecloudClusters(constants.OnecloudNamespace).Update(oc); err != nil {
			return errors.Wrap(err, "update default onecloud cluster")
		}
		if opt.wait {
			if err := ocutil.WaitOnecloudDeploymentUpdated(data.client, oc.GetName(), oc.GetNamespace(), 30*time.Minute, nil); err != nil {
				return errors.Wrap(err, "wait onecloud cluster updated")
			}
			rollout, err := data.kubeClient.Rollout()
			if err != nil {
				return errors.Wrap(err, "get rollout cmd")
			}
			if err := rollout.Status(0).
				SetNamespace(constants.OnecloudNamespace).
				RunDeployment(fmt.Sprintf("%s-web", oc.GetName())); err != nil {
				return err
			}
		}
	}
	return nil
}

func getRepoImageName(img string) (string, string, string, error) {
	ret, err := image.ParseImageReference(img)
	if err != nil {
		return "", "", "", err
	}
	tag := ret.Tag
	if tag == "" {
		tag = ret.Digest
	}
	return ret.Repository, ret.Image, tag, nil
}

func getOperatorImage(operator *appv1.Deployment) (*image.ImageReference, error) {
	img := operator.Spec.Template.Spec.Containers[0].Image
	return image.ParseImageReference(img)
}
