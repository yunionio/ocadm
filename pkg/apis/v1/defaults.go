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
	DefaultOnecloudRegion              = "region1"
	DefaultOnecloudZone                = "zone1"
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
	if obj.Zone == "" {
		obj.Zone = DefaultOnecloudZone
	}
	if obj.ImageRepository == "" {
		obj.ImageRepository = DefaultImageRepository
	}
	if obj.OnecloudCertificatesDir == "" {
		obj.OnecloudCertificatesDir = DefaultOnecloudCertificatesDir
	}
	SetDefaults_Keystone(&obj.Keystone)
}

func setDefaults_kubeadmInitConfiguration(obj *kubeadmapi.InitConfiguration) {
	defaultversionedcfg := &kubeadmapiv1beta1.InitConfiguration{}
	kubeadmscheme.Scheme.Convert(obj, defaultversionedcfg, nil)
	kubeadmscheme.Scheme.Default(defaultversionedcfg)

	// Takes passed flags into account; the defaulting is executed once again enforcing assignment of
	// static default values to cfg only for values not provided with flags
	kubeadmscheme.Scheme.Convert(defaultversionedcfg, obj, nil)
}

func SetDefaults_ServiceBaseOptions(obj *ServiceBaseOptions) {
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

	SetDefaults_ServiceBaseOptions(&obj.ServiceBaseOptions)

	if obj.ServiceBaseOptions.Port == 0 {
		obj.ServiceBaseOptions.Port = constants.KeystonePublicPort
	}
	if obj.AdminPort == 0 {
		obj.AdminPort = constants.KeystoneAdminPort
	}
	if obj.FernetKeyRepository == "" {
		obj.FernetKeyRepository = DefaultKeystoneFernetKeyRepository
	}
	obj.SetupCredentialKeys = false
	if obj.BootstrapAdminUserPassword == "" {
		obj.BootstrapAdminUserPassword = passwd.GeneratePassword()
	}
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
func SetDefaults_ServiceCommonOptions(obj *ServiceCommonOptions) {
	SetDefaults_ServiceBaseOptions(&obj.ServiceBaseOptions)
	if obj.AuthTokenCacheSize == 0 {
		obj.AuthTokenCacheSize = 2048
	}
}
