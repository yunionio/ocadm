package init

import (
	"io"

	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/util/mysql"
	"yunion.io/x/onecloud/pkg/mcclient"
)

// InitData is the interface to use for init phases.
// The "initData" type from "cmd/init.go" must satisfy this interface.
type InitData interface {
	RootDBConnection() (*mysql.Connection, error)
	OnecloudCfg() *v1.InitConfiguration
	LocalAddress() string
	Cfg() *kubeadmapi.InitConfiguration
	DryRun() bool
	IgnorePreflightErrors() sets.String
	ManifestDir() string
	OnecloudClientSession() (*mcclient.ClientSession, error)
	OnecloudCertificateWriteDir() string
	OnecloudCertificateDir() string
	Client() (clientset.Interface, error)
	OutputWriter() io.Writer
}
