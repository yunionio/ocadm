package init

import (
	initphases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/init"

	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/util/kubectl"
	"yunion.io/x/ocadm/pkg/util/mysql"
	"yunion.io/x/onecloud/pkg/mcclient"
)

// InitData is the interface to use for init phases.
// The "initData" type from "cmd/init.go" must satisfy this interface.
type InitData interface {
	initphases.InitData

	RootDBConnection() (*mysql.Connection, error)
	LocalAddress() string
	OnecloudAdminConfigPath() string
	OnecloudCfg() *apiv1.InitConfiguration
	OnecloudClientSession() (*mcclient.ClientSession, error)
	KubectlClient() (*kubectl.Client, error)
}
