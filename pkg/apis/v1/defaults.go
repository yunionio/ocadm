package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/util/passwd"
)

const (
	DefaultKubernetesVersion           = "v1.14.3"
	DefaultOnecloudVersion             = "latest"
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
	DefaultImageRepository = "registry.hub.docker.com/yunion"
	// DefaultImageRepository = "registry.cn-beijing.aliyuncs.com/yunionio"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_InitConfiguration assigns default values for the InitConfiguration
func SetDefaults_InitConfiguration(obj *InitConfiguration) {
	SetDefaults_ClusterConfiguration(&obj.ClusterConfiguration)
	setDefaults_kubeadmInitConfiguration(&obj.InitConfiguration)
	obj.InitConfiguration.ImageRepository = obj.ClusterConfiguration.ImageRepository
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
	if obj.ImageRepository == "" {
		obj.ImageRepository = DefaultImageRepository
	}
	if obj.OnecloudCertificatesDir == "" {
		obj.OnecloudCertificatesDir = DefaultOnecloudCertificatesDir
	}
	if obj.Region == "" {
		obj.Region = DefaultOnecloudRegion
	}
	if obj.BootstrapPassword == "" {
		obj.BootstrapPassword = passwd.GeneratePassword()
	}
}

// SetDefaults_JoinConfiguration assigns default values to a regular node
func SetDefaults_JoinConfiguration(obj *JoinConfiguration) {
	setDefaults_kubeadmJoinConfiguration(&obj.JoinConfiguration)
}

func setDefaults_kubeadmJoinConfiguration(obj *kubeadmapi.JoinConfiguration) {
	defaultversionedcfg := &kubeadmapiv1beta1.JoinConfiguration{}
	kubeadmscheme.Scheme.Convert(obj, defaultversionedcfg, nil)
	kubeadmscheme.Scheme.Default(defaultversionedcfg)
	kubeadmscheme.Scheme.Convert(defaultversionedcfg, obj, nil)
}

func setDefaults_kubeadmInitConfiguration(obj *kubeadmapi.InitConfiguration) {
	defaultversionedcfg := &kubeadmapiv1beta1.InitConfiguration{}
	kubeadmscheme.Scheme.Convert(obj, defaultversionedcfg, nil)
	kubeadmscheme.Scheme.Default(defaultversionedcfg)

	// Takes passed flags into account; the defaulting is executed once again enforcing assignment of
	// static default values to cfg only for values not provided with flags
	kubeadmscheme.Scheme.Convert(defaultversionedcfg, obj, nil)
}

func SetDefaults_ServiceBaseOptions(obj *ServiceBaseOptions, listenPort int) {
	if obj.Port == 0 {
		obj.Port = listenPort
	}
	if obj.Address == "" {
		obj.Address = "0.0.0.0"
	}
	obj.DebugClient = false
	if obj.LogLevel == "" {
		obj.LogLevel = "info"
	}
	if obj.TempPath == "" {
		obj.TempPath = constants.OnecloudOptTmpDir
	}
	if obj.RequestWorkerCount == 0 {
		obj.RequestWorkerCount = 4
	}
	if len(obj.NotifyAdminUsers) == 0 {
		obj.NotifyAdminUsers = []string{constants.SysAdminUsername}
	}
	obj.EnableSSL = true
	obj.EnableRBAC = true
	obj.RBACDebug = false
	obj.RBACPolicySyncPeriodSeconds = 300 // 5 mins
	if obj.CalculateQuotaUsageIntervalSeconds == 0 {
		obj.CalculateQuotaUsageIntervalSeconds = 300 // 5 mins
	}
}

func SetDefaults_DBInfo(obj *DBInfo, defaultDB, defaultUser string) {
	if obj.Database == "" {
		obj.Database = defaultDB
	}
	if obj.Username == "" {
		obj.Username = defaultUser
	}
	if obj.Password == "" {
		obj.Password = passwd.GeneratePassword()
	}
}

func SetDefaults_ServiceDBOptions(obj *ServiceDBOptions, defaultDB, defaultUser string) {
	SetDefaults_DBInfo(&obj.DBInfo, defaultDB, defaultUser)
	obj.ExitAfterDBInit = false
	obj.GlobalVirtualResourceNamespace = false
	obj.DebugSqlchemy = false
}

func SetDefaults_Keystone(obj *Keystone) {
	SetDefaults_ServiceDBOptions(&obj.ServiceDBOptions, constants.KeystoneDB, constants.KeystoneDBUser)

	SetDefaults_ServiceBaseOptions(&obj.ServiceBaseOptions, constants.KeystonePublicPort)

	if obj.AdminPort == 0 {
		obj.AdminPort = constants.KeystoneAdminPort
	}
	if obj.FernetKeyRepository == "" {
		obj.FernetKeyRepository = DefaultKeystoneFernetKeyRepository
	}
	obj.SetupCredentialKeys = false
	if obj.AutoSyncIntervalSeconds == 0 {
		obj.AutoSyncIntervalSeconds = 30
	}
	if obj.DefaultSyncIntervalSeoncds == 0 {
		obj.DefaultSyncIntervalSeoncds = 900
	}
	if obj.FetchProjectResourceCountIntervalSeconds == 0 {
		obj.FetchProjectResourceCountIntervalSeconds = 900
	}
}

// SetDefaults_ServiceBaseOptions
func SetDefaults_ServiceCommonOptions(obj *ServiceCommonOptions, region, project, username string, listenPort int) {
	SetDefaults_ServiceBaseOptions(&obj.ServiceBaseOptions, listenPort)
	if obj.AuthTokenCacheSize == 0 {
		obj.AuthTokenCacheSize = 2048
	}
	if obj.Region == "" {
		obj.Region = region
	}
	if obj.AdminProject == "" {
		obj.AdminProject = project
	}
	if obj.AdminUser == "" {
		obj.AdminUser = username
	}
	if obj.AdminPassword == "" {
		obj.AdminPassword = passwd.GeneratePassword()
	}
}

func SetDefaults_RegionServer(obj *RegionServer, region string) {
	SetDefaults_ServiceDBOptions(&obj.ServiceDBOptions, constants.RegionDB, constants.RegionDBUser)
	SetDefaults_ServiceCommonOptions(&obj.ServiceCommonOptions, region, constants.SysAdminProject, constants.RegionAdminUser, constants.RegionPort)
	if obj.PortV2 == 0 {
		obj.PortV2 = obj.Port
	}
	if obj.SchedulerPort == 0 {
		obj.SchedulerPort = constants.SchedulerPort
	}
}

func SetDefaults_BaremetalAgent(obj *BaremetalAgent, region string) {
	SetDefaults_ServiceCommonOptions(&obj.ServiceCommonOptions, region, constants.SysAdminProject, constants.BaremetalAdminUser, constants.BaremetalPort)
}

func SetDefaults_HostLocalInfo(obj *HostLocalInfo) {
	if obj.Zone == "" {
		obj.Zone = DefaultOnecloudZone
	}
	if obj.ManagementNetInterface.Wire == "" {
		obj.ManagementNetInterface.Wire = DefaultOnecloudAdminWire
	}
}

func SetDefaults_Glance(obj *Glance, region string) {
	SetDefaults_ServiceDBOptions(&obj.ServiceDBOptions, constants.GlanceDB, constants.GlanceDBUser)
	SetDefaults_ServiceCommonOptions(&obj.ServiceCommonOptions, region, constants.GlanceAdminProject, constants.GlanceAdminUser, constants.GlanceAPIPort)
	if obj.FilesystemStoreDatadir == "" {
		obj.FilesystemStoreDatadir = constants.OnecloudGlanceFileStoreDir
	}
	if obj.TorrentStoreDir == "" {
		obj.TorrentStoreDir = constants.OnecloudGlanceTorrentStoreDir
	}
}
