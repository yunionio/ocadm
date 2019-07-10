package constants

import (
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

const (
	SysAdminUsername = "sysadmin"
	SysAdminProject  = "system"
	DefaultDomain    = "Default"

	KeystoneDB         = "keystone"
	KeystoneDBUser     = "keystone"
	KeystonePublicPort = 5000
	KeystoneAdminPort  = 35357

	GlanceDB           = "glance"
	GlanceDBUser       = "glance"
	GlanceAdminUser    = "glance"
	GlanceAdminProject = SysAdminProject
	GlanceRegistryPort = 9191
	GlanceAPIPort      = 9292
	ServiceNameGlance  = "glance"
	ServiceTypeGlance  = "image"

	RegionAdminUser    = "regionadmin"
	RegionAdminProject = SysAdminProject
	RegionPort         = 8889
	SchedulerPort      = 8897
	RegionDB           = "yunioncloud"
	RegionDBUser       = "yunioncloud"

	BaremetalAdminUser    = "baremetal"
	BaremetalAdminProject = SysAdminProject
	BaremetalPort         = 8879

	WebconsoleAdminUser    = "webconsole"
	WebconsoleAdminProject = SysAdminProject
	WebconsolePort         = 8899

	InfluxdbPort = 8086

	AnsibleServerAdminUser    = "ansibleadmin"
	AnsibleServerAdminProject = SysAdminProject
	AnsibleServerPort         = 8890
	AnsibleServerDB           = "yunionansible"

	OnecloudKeystone       = "keystone"
	OnecloudRegion         = "region"
	OnecloudScheduler      = "scheduler"
	OnecloudGlance         = "glance"
	OnecloudBaremetal      = "baremetal"
	OnecloudBaremetalAgent = "baremetal-agent" // used by image
	OnecloudWebconsole     = "webconsole"
	OnecloudInfluxdb       = "influxdb"
	OnecloudAPIGateway     = "yunionapi"

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
)

const (
	OnecloudConfigVolumeName        = "config"
	OnecloudEtcKeystoneVolumeName   = "etc-yunion-keystone"
	OnecloudOptVolumeName           = "opt-yunion"
	OnecloudOptTmpVolumeName        = "opt-yunion-tmp"
	OnecloudPKICertsVolumeName      = "pki-certs"
	OnecloudGlanceImageVolumeName   = "glance-images"
	OnecloudGlanceImageTorrentName  = "glance-torrents"
	OnecloudQemuBinaryVolumeName    = "qemu"
	OnecloudKernelVolumeName        = "kernel"
	OnecloudBaremetalTFTPVolumeName = "tftp"
	OnecloudBaremetalsVolumeName    = "bms"

	OnecloudConfigDir              = "/etc/yunion"
	OnecloudKeystoneConfigDir      = "/etc/yunion/keystone"
	OnecloudConfigFileSuffix       = ".yaml"
	OnecloudKeystoneConfigFileName = "keystone.conf"
	OnecloudOptDir                 = "/opt/yunion"
	OnecloudOptTmpDir              = "/opt/yunion/tmp"

	OnecloudRegionConfigFileName = "region.conf"
	OnecloudAdminConfigFileName  = "rc_admin"

	OnecloudGlanceConfigDir       = "/etc/yunion/glance"
	OnecloudGlanceConfigFileName  = "glance-api.conf"
	OnecloudGlanceFileStoreDir    = "/opt/cloud/workspace/data/glance/images"
	OnecloudGlanceTorrentStoreDir = "/opt/cloud/workspace/data/glance/torrents"

	OnecloudQemuPath   = "/usr/local/qemu-2.12.1"
	OnecloudKernelPath = "/lib/modules"

	OnecloudBaremetalConfigFileName = "baremetal.conf"
	OnecloudBaremetalsPath          = "/opt/cloud/workspace/baremetals"
	OnecloudBaremetalTFTPRoot       = "/opt/cloud/yunion/baremetal"

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

	// KeystoneCertAndKeyBaseName defines keystone server certificate and key base name
	KeystoneCertAndKeyBaseName = "keystone"
	// KeystoneCertName defines keystone server certificate name
	KeystoneCertName = "keystone.crt"
	// KeystoneKeyName defines keysotne server key name
	KeystoneKeyName = "keystone.key"

	ClimcClientCertAndKeyBaseName = "climc"
	ClimcCertName                 = "climc.crt"
	ClimcKeyName                  = "climc.key"

	RegionCertAndKeyBaseName = "region"
	RegionCertName           = "region.crt"
	RegionKeyName            = "region.key"

	GlanceCertAndKeyBaseName = "glance"
	GlanceCertName           = "glance.crt"
	GlanceKeyName            = "glance.key"

	BaremetalCertName = "baremetal.crt"
	BaremetalKeyName  = "baremetal.key"

	WebconsoleCertName = "webconsole.crt"
	WebconsoleKeyName  = "webconsole.key"

	InfluxdbCertName = "influxdb.crt"
	InfluxdbKeyName  = "influxdb.key"

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
