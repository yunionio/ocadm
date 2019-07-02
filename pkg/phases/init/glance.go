package init

import (
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/ocadm/pkg/apis/constants"
	v1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/glance"
	"yunion.io/x/ocadm/pkg/util/mysql"
	"yunion.io/x/ocadm/pkg/util/onecloud"
	"yunion.io/x/onecloud/pkg/mcclient"
)

func NewGlancePhase() workflow.Phase {
	servicePhase := &ServiceBasePhase{
		Name: constants.ServiceNameGlance,
		Type: constants.ServiceTypeGlance,
		InheritFlags: []string{
			options.CfgPath,
		},
		SetupUseSession: true,
		SetupFunc: func(sqlConn *mysql.Connection, s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, localCfg *v1.HostLocalInfo, certDir string) error {
			return glance.SetupGlanceServer(s, sqlConn, clusterCfg.Glance,
				certDir, localCfg.ManagementNetInterface.IPAddress())
		},
		WaitRunningFunc: func(waiter onecloud.Waiter) error {
			return waiter.WaitForGlance()
		},
		SysInitFunc: func(s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, localCfg *v1.HostLocalInfo) error {
			return nil
		},
	}
	return servicePhase.ToPhase()
}
