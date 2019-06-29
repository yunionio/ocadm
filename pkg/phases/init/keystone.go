package init

import (
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/onecloud/pkg/mcclient"

	"yunion.io/x/ocadm/pkg/apis/constants"
	v1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/keystone"
	"yunion.io/x/ocadm/pkg/util/mysql"
	"yunion.io/x/ocadm/pkg/util/onecloud"
)

// NewKeystonePhase creates a ocadm workflow phase that implements handing of keystone
func NewKeystonePhase() workflow.Phase {
	servicePhase := &ServiceBasePhase{
		Name: constants.OnecloudKeystone,
		Type: constants.ServiceTypeIdentity,
		InheritFlags: []string{
			options.CfgPath,
			options.MysqlAddress,
			options.MysqlPort,
			options.MysqlUser,
			options.MysqlPassword,
			options.Region,
		},
		SetupFunc: func(sqlConn *mysql.Connection, _ *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, localAddress string, certDir string) error {
			return keystone.SetupKeystone(sqlConn, &clusterCfg.Keystone, clusterCfg.Region, localAddress, certDir)
		},
		WaitRunningFunc: func(waiter onecloud.Waiter) error {
			return waiter.WaitForKeystone()
		},
		SysInitFunc: func(s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, localAddress string) error {
			return keystone.DoSysInit(s, clusterCfg, localAddress)
		},
	}

	return servicePhase.ToPhase()
}
