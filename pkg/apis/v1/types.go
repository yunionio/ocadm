package v1

import (
	"fmt"
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InitConfiguration struct {
	metav1.TypeMeta

	// InitConfiguration holds the kubeadm init configuration
	kubeadmapi.InitConfiguration `json:"-"`

	// ClusterConfiguration holds the cluster-wide information, and embeds that struct (which can be (un)marshalled separately as well)
	ClusterConfiguration `json:"-"`

	// HostLocalInfo holds the local node info
	HostLocalInfo `json:"-"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JoinConfiguration contains elements describing a particular node.
type JoinConfiguration struct {
	metav1.TypeMeta

	kubeadmapi.JoinConfiguration `json:"-"`
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

	// ImageRepository sets the container registry to pull images from.
	// If empty, `k8s.gcr.io` will be used by default; in case of kubernetes version is a CI build (kubernetes version starts with `ci/` or `ci-cross/`)
	// `gcr.io/kubernetes-ci-images` will be used as a default for control plane components and for kube-proxy, while `k8s.gcr.io`
	// will be used for all the other images.
	ImageRepository string
}

type HostLocalInfo struct {
	// Zone is the first default zone
	Zone string

	// ManagementNetInterface is teh management services network interface
	ManagementNetInterface NetInterface
}

type NetInterface struct {
	// Wire is the first default management wire
	Wire string

	// Address is the services endpoint address
	Address net.IP

	// MaskLen is the address mask length
	MaskLen int

	// Interface is the listen network interface
	Interface string

	// Gateway is the interface default gateway
	Gateway net.IP
}

func (iface NetInterface) IPAddress() string {
	return iface.Address.String()
}

type MysqlConnection struct {
	Server   string
	Port     int
	Username string
	Password string
}

type DBInfo struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func (info DBInfo) ToSQLConnection() string {
	return fmt.Sprintf("mysql+pymysql://%s:%s@%s:%d/%s?charset=utf8", info.Username, info.Password, info.Host, info.Port, info.Database)
}
