package init

import (
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/onecloud/pkg/mcclient"

	"yunion.io/x/ocadm/pkg/apis/constants"
	v1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/region"
	"yunion.io/x/ocadm/pkg/util/mysql"
	"yunion.io/x/ocadm/pkg/util/onecloud"
)

// NewRegionPhase creates a ocadm workflow phase that implements handing of region
func NewRegionPhase() workflow.Phase {
	servicePhase := &ServiceBasePhase{
		Name: constants.OnecloudRegion,
		Type: constants.ServiceTypeComputeV2,
		InheritFlags: []string{
			options.CfgPath,
		},
		SetupUseSession: true,
		SetupFunc: func(sqlConn *mysql.Connection, s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, localCfg *v1.HostLocalInfo, certDir string) error {
			return region.SetupRegionServer(s, sqlConn, clusterCfg.RegionServer, certDir, localCfg)
		},
		WaitRunningFunc: func(waiter onecloud.Waiter) error {
			return waiter.WaitForRegion()
		},
		SysInitFunc: func(s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, localCfg *v1.HostLocalInfo) error {
			return region.DoSysInit(s, clusterCfg, localCfg)
		},
	}
	return servicePhase.ToPhase()
}

func NewSchedulerPhase() workflow.Phase {
	p := &ServiceBasePhase{
		Name: constants.OnecloudScheduler,
		Type: constants.ServiceTypeScheduler,
		InheritFlags: []string{
			options.CfgPath,
		},
		SetupUseSession: true,
		SetupFunc: func(sqlConn *mysql.Connection, s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, hostLocalCfg *v1.HostLocalInfo, certDir string) error {
			return nil
		},
		WaitRunningFunc: func(waiter onecloud.Waiter) error {
			// ServiceBasePhase will create scheduler pods
			return nil
		},
		SysInitFunc: func(s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, hostLocalCfg *v1.HostLocalInfo) error {
			return nil
		},
	}
	return p.ToPhase()
}
