package v1

import (
	"fmt"

	//v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//kubeletconfig "k8s.io/kubernetes/pkg/kubelet/apis/config"
	//kubeproxyconfig "k8s.io/kubernetes/pkg/proxy/apis/config"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InitConfiguration struct {
	metav1.TypeMeta

	// KubeadmInitConfiguration holds the kubeadm init configuration
	kubeadmapi.InitConfiguration `json:"-"`

	// ClusterConfiguration holds the cluster-wide information, and embeds that struct (which can be (un)marshalled separately as well)
	ClusterConfiguration `json:"-"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterConfiguration contains cluster-wide configuration for a onecloud cluster
type ClusterConfiguration struct {
	metav1.TypeMeta

	// MysqlConnection specifies mysql admin connection info.
	MysqlConnection MysqlConnection

	// OnecloudVersion is the target version of the control plane.
	OnecloudVersion string

	// Region specify keystone auth region
	Region string

	// Zone is the first default zone
	Zone string

	// Keystone holds configuration for keystone.
	Keystone Keystone

	// RegionServer holds configuration for controller region service.
	RegionServer RegionServer

	// ImageRepository sets the container registry to pull images from.
	// If empty, `k8s.gcr.io` will be used by default; in case of kubernetes version is a CI build (kubernetes version starts with `ci/` or `ci-cross/`)
	// `gcr.io/kubernetes-ci-images` will be used as a default for control plane components and for kube-proxy, while `k8s.gcr.io`
	// will be used for all the other images.
	ImageRepository string

	// OnecloudCertificatesDir specifies where to store or look for all required certificates.
	OnecloudCertificatesDir string
}

type MysqlConnection struct {
	Server   string
	Port     int
	Username string
	Password string
}

type ServiceBaseOptions struct {
	Region string

	Port    int
	Address string

	DebugClient     bool
	LogLevel        string
	LogVerboseLevel string
	LogFilePrefix   string

	CorsHosts []string
	TempPath  string

	ApplicationID      string
	RequestWorkerCount int

	EnableSSL   bool
	SSLCAFile   string
	SSLCertFile string
	SSLKeyFile  string

	NotifyAdminUsers  []string
	NotifyAdminGroups []string

	EnableRBAC                  bool
	RBACDebug                   bool
	RBACPolicySyncPeriodSeconds int

	IsSlaveNode bool

	CalculateQuotaUsageIntervalSeconds int
}

type ServiceCommonOptions struct {
	ServiceBaseOptions

	AuthURL            string
	AdminUser          string
	AdminDomain        string
	AdminPassword      string
	AdminProject       string
	AdminProjectDomain string
	AuthTokenCacheSize uint32
}

type DBInfo struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type ServiceDBOptions struct {
	DBInfo

	AutoSyncTable   bool
	ExitAfterDBInit bool

	GlobalVirtualResourceNamespace bool
	DebugSqlchemy                  bool

	QueryOffsetOptimization bool
}

func (info DBInfo) ToSQLConnection() string {
	return fmt.Sprintf("mysql+pymysql://%s:%s@%s:%d/%s?charset=utf8", info.Username, info.Password, info.Host, info.Port, info.Database)
}

type Keystone struct {
	ServiceBaseOptions
	ServiceDBOptions

	// listening port for admin API(deprecated), default: 35357
	AdminPort int

	// token expiration seconds, default: 86400
	TokenExpirationSeconds int

	// fernet key repo directory, default: /etc/yunion/keystone/fernet-keys
	FernetKeyRepository string

	// setup standalone fernet keys for credentials, default: false
	SetupCredentialKeys bool

	// bootstreap sysadmin user password
	BootstrapAdminUserPassword string

	// frequency to check auto sync tasks, default: 30
	AutoSyncIntervalSeconds int

	// frequency to do auto sync tasks, default: 900
	DefaultSyncIntervalSeoncds int

	// frequency tp fetch project resource counts, default: 900
	FetchProjectResourceCountIntervalSeconds int
}

type RegionServer struct {
	ServiceCommonOptions
	ServiceDBOptions

	// Address of DNS server
	DNSServer string
	// Domain suffix for virtual servers
	DNSDomain string
	// Upstream DNS resolvers
	DNSResolvers []string

	// Listening port for region V2
	PortV2 int
	// The port that the scheduler's http service runs on
	SchedulerPort int

	// Count memory for running guests only when do scheduling. Ignore memory allocation for non-running guests
	IgnoreNonRunningGuests bool

	// Baremetal online register package
	BaremetalPreparePackageUrl string
	// Kvm baremetal convert option
	ConvertHypervisorDefaultTemplate string
}
