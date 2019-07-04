package components

import (
	"io"
	clientset "k8s.io/client-go/kubernetes"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"yunion.io/x/onecloud/pkg/mcclient"

	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/util/mysql"
)

type ComponentData interface {
	Cfg() *kubeadmapi.InitConfiguration
	Client() (clientset.Interface, error)
	OutputWriter() io.Writer
	ManifestDir() string
	RootDBConnection() (*mysql.Connection, error)
	OnecloudAdminConfigPath() string
	OnecloudCfg() *apiv1.InitConfiguration
	OnecloudClientSession() (*mcclient.ClientSession, error)
	OnecloudCertificateWriteDir() string
	OnecloudCertificateDir() string
}

type ServiceAccount struct {
	AdminUser     string
	AdminPassword string
}

type CertConfig struct {
	CertName string
}
