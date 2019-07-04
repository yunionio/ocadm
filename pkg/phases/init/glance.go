package init

/*func NewGlancePhase() workflow.Phase {
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
}*/
