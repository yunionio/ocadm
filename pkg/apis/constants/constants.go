package constants

import (
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"yunion.io/x/pkg/util/version"
)

var (
	DefaultOperatorVersion = version.Get().GitVersion
	DefaultOnecloudVersion = DefaultOperatorVersion
)

const (
	SysAdminUsername = "sysadmin"
	SysAdminProject  = "system"
	DefaultDomain    = "Default"

	OnecloudOperator               = "onecloud-operator"
	RancherLocalPathProvisioner    = "local-path-provisioner"
	DefaultLocalProvisionerVersion = "v0.0.11"
	IngressControllerTraefik       = "traefik"
	DefaultTraefikVersion          = "v1.7"
	CalicoKubeControllers          = "calico-kube-controllers"
	CalicoNode                     = "calico-node"
	CalicoCNI                      = "calico-cni"
	DefaultCalicoVersion           = "v3.12.1"
	Loki                           = "loki"
	DefaultLokiVersion             = "v1.2.0"
	Promtail                       = "promtail"
	DefaultPromtailVersion         = DefaultLokiVersion
	Grafana                        = "grafana"
	DefaultGrafanaVersion          = "6.5.2"
	DefaultKeepalivedVersionTag    = "v2.0.23"
	// mirror of kiwigrid/k8s-sidecar:0.1.20
	K8sSidecar               = "k8s-sidecar"
	DefaultK8sSidecarVersion = "0.1.20"
	Busybox                  = "busybox"
	BusyboxVersion           = "1.28.0-glibc"
	MetricsServer            = "metrics-server-amd64"
	MetricsServerVersion     = "v0.3.6"

	EndpointTypeInternal = "internal"
	EndpointTypePublic   = "public"
	EndpointTypeAdmin    = "admin"
	EndpointTypeConsole  = "console"

	// define service constants
	ServiceNameKeystone = "keystone"
	ServiceTypeIdentity = "identity"

	ServiceNameRegion    = "region"
	ServiceNameRegionV2  = "region2"
	ServiceTypeCompute   = "compute"
	ServiceTypeComputeV2 = "compute_v2"

	ServiceNameScheduler = "scheduler"
	ServiceTypeScheduler = "scheduler"

	ServiceNameWebconsole = "webconsole"
	ServiceTypeWebconsole = "webconsole"

	ServiceNameInfluxdb = "influxdb"
	ServiceTypeInfluxdb = "influxdb"

	ServiceURLCloudmeta  = "https://meta.yunion.cn"
	ServiceNameCloudmeta = "cloudmeta"
	ServiceTypeCloudmeta = "cloudmeta"

	ServiceURLTorrentTracker  = "https://tracker.yunion.cn"
	ServiceNameTorrentTracker = "torrent-tracker"
	ServiceTypeTorrentTracker = "torrent-tracker"

	NetworkTypeBaremetal = "baremetal"
	NetworkTypeServer    = "server"

	// longhorn
	LonghornStorageClass                      = "longhorn"
	DefaultLonghornVersion                    = "v1.0.0"
	LonghornManager                           = "longhorn-manager"
	LonghornEngine                            = "longhorn-engine"
	LonghornUi                                = "longhorn-ui"
	LonghornDefaultDataPath                   = "/opt/longhorn"
	LonghornDefaultOverProvisioningPercentage = 100
	LonghornDefaultReplicaCount               = 3
	// longhorn-instance-manager image name must be no more than 63 characters
	// https://github.com/longhorn/longhorn/issues/1106
	// registry.cn-beijing.aliyuncs.com/yunionio/longhorn-instance-manager is too long
	LonghornInstanceManager = "longhorn-im"
	LonghornCreateDiskLable = "node.longhorn.io/create-default-disk"
)

const (
	OnecloudNamespace = "onecloud"

	OnecloudConfigDir              = "/etc/yunion"
	OnecloudKeystoneConfigDir      = "/etc/yunion/keystone"
	OnecloudConfigFileSuffix       = ".yaml"
	OnecloudKeystoneConfigFileName = "keystone.conf"

	OnecloudRegionConfigFileName = "region.conf"
	OnecloudAdminConfigFileName  = "rc_admin"

	OnecloudGlanceConfigDir      = "/etc/yunion/glance"
	OnecloudGlanceConfigFileName = "glance-api.conf"

	OnecloudBaremetalConfigFileName = "baremetal.conf"

	OnecloudWebconsoleConfigFileName = "webconsole.conf"

	OnecloudInfluxdbConfigFileName = "influxdb.conf"

	// OnecloudAdminConfigConfigMap specifies in what ConfigMap in the kube-system namespace the `ocadm init` configuration should be stored
	OnecloudAdminConfigConfigMap = "ocadm-config"

	// ClusterConfigurationConfigMapKey specifies in what ConfigMap key the cluster configuration should be stored
	ClusterConfigurationConfigMapKey = "ClusterConfiguration"

	// ClusterAdminAuthConfigMapKey specifies keystone admin auth info
	ClusterAdminAuthConfigMapKey = "AdminAuthConfiguration"

	// ClusterConfigurationKind is the string kind value for the ClusterConfiguration struct
	ClusterConfigurationKind = "ClusterConfiguration"

	// InitConfigurationKind is the string kind value for the InitConfiguration struct
	InitConfigurationKind = "InitConfiguration"

	// JoinConfigurationKind is the string kind value for the JoinConfiguration struct
	JoinConfigurationKind = "JoinConfiguration"
)

const (
	// CACertAndKeyBaseName defines certificate authority base name
	CACertAndKeyBaseName = kubeadmconstants.CACertAndKeyBaseName
	// CACertName defines certificate name
	CACertName = kubeadmconstants.CACertName
	// CAKeyName defines certificate name
	CAKeyName = kubeadmconstants.CAKeyName

	AdminKubeConfigFileName = kubeadmconstants.AdminKubeConfigFileName

	ClimcClientCertAndKeyBaseName = "climc"
	ClimcCertName                 = "climc.crt"
	ClimcKeyName                  = "climc.key"

	OcadmCertsSecret = "ocadm-certs"
)

const (
	RoleAdmin        = "admin"
	RoleFA           = "fa"
	RoleSA           = "sa"
	RoleProjectOwner = "project_owner"
	RoleMember       = "member"
	RoleDomainAdmin  = "domainadmin"

	PolicyTypeDomainAdmin  = "domainadmin"
	PolicyTypeMember       = "member"
	PolicyTypeProjectFA    = "projectfa"
	PolicyTypeProjectOwner = "projectowner"
	PolicyTypeProjectSA    = "projectsa"
	PolicyTypeSysAdmin     = "sysadmin"
	PolicyTypeSysFA        = "sysfa"
	PolicyTypeSysSA        = "syssa"
)

var (
	PublicRoles = []string{
		RoleFA,
		RoleSA,
		RoleProjectOwner,
		RoleMember,
		RoleDomainAdmin,
	}
	PublicPolicies = []string{
		PolicyTypeDomainAdmin, PolicyTypeProjectOwner,
		PolicyTypeProjectSA, PolicyTypeProjectFA,
		PolicyTypeMember,
	}
)

const (
	YAMLDocumentSeparator    = kubeadmconstants.YAMLDocumentSeparator
	APICallRetryInterval     = kubeadmconstants.APICallRetryInterval
	DiscoveryRetryInterval   = kubeadmconstants.DiscoveryRetryInterval
	DefaultCertTokenDuration = kubeadmconstants.DefaultCertTokenDuration
)

var (
	GetStaticPodFilepath     = kubeadmconstants.GetStaticPodFilepath
	GetStaticPodDirectory    = kubeadmconstants.GetStaticPodDirectory
	GetAdminKubeConfigPath   = kubeadmconstants.GetAdminKubeConfigPath
	GetKubeletKubeConfigPath = kubeadmconstants.GetKubeletKubeConfigPath
)
