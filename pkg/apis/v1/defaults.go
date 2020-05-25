package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
	kubeproxyconfig "k8s.io/kubernetes/pkg/proxy/apis/config"
	"yunion.io/x/ocadm/pkg/apis/constants"
)

const (
	DefaultKubernetesVersion           = "v1.15.8"
	DefaultPodSubnetCIDR               = "10.40.0.0/16"
	DefaultOnecloudRegion              = "region0"
	DefaultOnecloudZone                = "zone0"
	DefaultOnecloudAdminWire           = "badm"
	DefaultOnecloudMasterWire          = "bcast0"
	DefaultOnecloudAdminNetwork        = "adm0"
	DefaultOnecloudHostNetwork         = "inf0"
	DefaultOnecloudInterface           = "eth0"
	DefaultVPCId                       = "default"
	DefaultMysqlUser                   = "root"
	DefaultMysqlAddress                = "127.0.0.1"
	DefaultMysqlPort                   = 3306
	DefaultKeystoneFernetKeyRepository = "/etc/yunion/keystone/fernet-keys"

	// DefaultOnecloudCertificatesDir defines default onecloud certificate directory
	DefaultOnecloudCertificatesDir = "/etc/yunion/pki"

	// DefaultImageRepository defines dfault image registry
	// DefaultImageRepository = "registry.hub.docker.com/yunion"
	DefaultImageRepository = "registry.cn-beijing.aliyuncs.com/yunionio"

	// default etcd version
	DefaultEtcdVersion = "3.4.6"
)

var (
	DefaultOperatorVersion = constants.DefaultOperatorVersion
	DefaultOnecloudVersion = constants.DefaultOnecloudVersion
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_InitConfiguration assigns default values for the InitConfiguration
func SetDefaults_InitConfiguration(obj *InitConfiguration) {
	SetDefaults_ClusterConfiguration(&obj.ClusterConfiguration)
	setDefaults_kubeadmInitConfiguration(&obj.InitConfiguration)
	SetDefaults_HostLocalInfo(&obj.HostLocalInfo)
}

func SetDefaults_MysqlConnection(obj *MysqlConnection) {
	if obj.Username == "" {
		obj.Username = DefaultMysqlUser
	}
	if obj.Server == "" {
		obj.Server = DefaultMysqlAddress
	}
	if obj.Port == 0 {
		obj.Port = DefaultMysqlPort
	}
}

// SetDefaults_ClusterConfiguration assigns default values for the ClusterConfiguration
func SetDefaults_ClusterConfiguration(obj *ClusterConfiguration) {
	SetDefaults_MysqlConnection(&obj.MysqlConnection)
	if obj.OnecloudVersion == "" {
		obj.OnecloudVersion = DefaultOnecloudVersion
	}
	if obj.Region == "" {
		obj.Region = DefaultOnecloudRegion
	}
}

// SetDefaults_JoinConfiguration assigns default values to a regular node
func SetDefaults_JoinConfiguration(obj *JoinConfiguration) {
	setDefaults_kubeadmJoinConfiguration(&obj.JoinConfiguration)
}

func setDefaults_kubeadmJoinConfiguration(obj *kubeadmapi.JoinConfiguration) {
	defaultversionedcfg := &kubeadmapiv1beta2.JoinConfiguration{}
	kubeadmscheme.Scheme.Convert(obj, defaultversionedcfg, nil)
	kubeadmscheme.Scheme.Default(defaultversionedcfg)
	kubeadmscheme.Scheme.Convert(defaultversionedcfg, obj, nil)
}

func setDefaults_kubeadmInitConfiguration(obj *kubeadmapi.InitConfiguration) {
	defaultversionedcfg := &kubeadmapiv1beta2.InitConfiguration{}
	dvClustercfg := &kubeadmapiv1beta2.ClusterConfiguration{}
	clusterConfig := &obj.ClusterConfiguration
	kubeadmscheme.Scheme.Convert(obj, defaultversionedcfg, nil)
	kubeadmscheme.Scheme.Default(defaultversionedcfg)

	kubeadmscheme.Scheme.Convert(clusterConfig, dvClustercfg, nil)
	kubeadmscheme.Scheme.Default(dvClustercfg)

	// Takes passed flags into account; the defaulting is executed once again enforcing assignment of
	// static default values to cfg only for values not provided with flags
	kubeadmscheme.Scheme.Convert(defaultversionedcfg, obj, nil)
	kubeadmscheme.Scheme.Convert(dvClustercfg, &obj.ClusterConfiguration, nil)

	obj.KubernetesVersion = DefaultKubernetesVersion
	obj.Networking.PodSubnet = DefaultPodSubnetCIDR

	// adjust extra args
	if obj.APIServer.ExtraArgs == nil {
		obj.APIServer.ExtraArgs = make(map[string]string)
	}
	obj.APIServer.ExtraArgs["default-not-ready-toleration-seconds"] = "10"
	obj.APIServer.ExtraArgs["default-unreachable-toleration-seconds"] = "10"
	// obj.APIServer.ExtraArgs["service-node-port-range"] = "5000-35357"
	// obj.APIServer.ExtraArgs["service-node-port-range"] = "30000-32767"
	if obj.ControllerManager.ExtraArgs == nil {
		obj.ControllerManager.ExtraArgs = make(map[string]string)
	}
	obj.ControllerManager.ExtraArgs["node-monitor-grace-period"] = "16s"
	obj.ControllerManager.ExtraArgs["node-monitor-period"] = "2s"

	if obj.ImageRepository == "" {
		obj.ImageRepository = DefaultImageRepository
	}
	if obj.ComponentConfigs.KubeProxy == nil {
		obj.ComponentConfigs.KubeProxy = &kubeproxyconfig.KubeProxyConfiguration{
			Mode: kubeproxyconfig.ProxyModeIPVS,
		}
	}
	if obj.Etcd.Local != nil {
		obj.Etcd.Local.ImageTag = "3.4.6"
	}
}

func SetDefaults_HostLocalInfo(obj *HostLocalInfo) {
	if obj.Zone == "" {
		obj.Zone = DefaultOnecloudZone
	}
	if obj.ManagementNetInterface.Wire == "" {
		obj.ManagementNetInterface.Wire = DefaultOnecloudAdminWire
	}
}
