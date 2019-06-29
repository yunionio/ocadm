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
		SetupFunc: func(sqlConn *mysql.Connection, s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, localAddress string, certDir string) error {
			return region.SetupRegionServer(s, sqlConn, clusterCfg.RegionServer, certDir, localAddress)
		},
		WaitRunningFunc: func(waiter onecloud.Waiter) error {
			return waiter.WaitForRegion()
		},
		SysInitFunc: func(s *mcclient.ClientSession, clusterCfg *v1.ClusterConfiguration, localAddress string) error {
			return nil
		},
	}
	return servicePhase.ToPhase()
}
