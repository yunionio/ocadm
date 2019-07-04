package region

import (
	"fmt"

	"yunion.io/x/onecloud/pkg/compute/options"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/pkg/errors"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/components"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/onecloud"
)

var (
	RegionComponent    *components.Component
	SchedulerComponent *components.Component
)

func init() {
	RegionComponent = &components.Component{
		Name:                 constants.OnecloudRegion,
		ServiceName:          constants.ServiceNameRegion,
		ServiceType:          constants.ServiceTypeComputeV2,
		CertConfig:           &components.CertConfig{constants.RegionCertName},
		ConfigDir:            constants.OnecloudConfigDir,
		ConfigFileName:       constants.OnecloudRegionConfigFileName,
		ConfigurationFactory: GetRegionOptions,
		GetDBInfo:            GetDBInfo,
		GetServiceAccount:    GetServiceAccount,
		SetupFunc:            SetupRegionServer,
		WaitRunningFunc: func(waiter onecloud.Waiter) error {
			return waiter.WaitForRegion()
		},
		SysInitFunc: DoSysInit,
	}

	SchedulerComponent = &components.Component{
		Name:        constants.OnecloudScheduler,
		ServiceName: constants.ServiceNameScheduler,
		ServiceType: constants.ServiceTypeScheduler,
		SetupFunc:   SetupScheduler,
		UseSession:  true,
	}
}

func GetRegionOptions(authConfig *occonfig.RCAdminConfig, clusterCfg *apiv1.ClusterConfiguration, _ *apiv1.HostLocalInfo, certDir string) (interface{}, interface{}, error) {
	config := &apiv1.RegionServer{}
	apiv1.SetDefaults_RegionServer(config, clusterCfg.Region)
	configutil.SetServiceDBInfo(&config.ServiceDBOptions.DBInfo, &clusterCfg.MysqlConnection, constants.RegionDB, constants.RegionDBUser)
	configutil.SetServiceAuthInfo(&config.ServiceCommonOptions, authConfig.Region, authConfig.AuthUrl)

	opt := &options.Options
	if err := configutil.SetOptionsDefault(opt, constants.ServiceTypeCompute); err != nil {
		return nil, nil, err
	}
	configutil.SetDBOptions(&opt.DBOptions, config.ServiceDBOptions)

	configutil.EnableConfigTLS(&config.ServiceBaseOptions, certDir, constants.CACertName, constants.RegionCertName, constants.RegionKeyName)

	configutil.SetServiceCommonOptions(&opt.CommonOptions, config.ServiceCommonOptions)

	opt.PortV2 = config.PortV2
	opt.AutoSyncTable = true

	return config, &occonfig.RegionSchedulerOptions{
		ComputeOptions: *opt,
		SchedulerPort:  config.SchedulerPort,
	}, nil
}

func GetDBInfo(cfgObj interface{}) *apiv1.DBInfo {
	return &cfgObj.(*apiv1.RegionServer).ServiceDBOptions.DBInfo
}

func GetServiceAccount(opt interface{}) *components.ServiceAccount {
	config := opt.(*apiv1.RegionServer)
	return &components.ServiceAccount{
		AdminUser:     config.ServiceCommonOptions.AdminUser,
		AdminPassword: config.ServiceCommonOptions.AdminPassword,
	}
}

func SetupRegionServer(
	s *mcclient.ClientSession,
	cfgObj interface{},
	_ *apiv1.ClusterConfiguration,
	localCfg *apiv1.HostLocalInfo,
) error {
	config := cfgObj.(*apiv1.RegionServer)
	url := fmt.Sprintf("https://%s:%d", localCfg.ManagementNetInterface.IPAddress(), config.PortV2)
	if err := occonfig.RegisterServicePublicInternalEndpoint(s, config.Region, constants.ServiceNameRegionV2, constants.ServiceTypeComputeV2, url); err != nil {
		return err
	}
	return nil
}

func DoSysInit(s *mcclient.ClientSession, cfg *apiv1.ClusterConfiguration, localCfg *apiv1.HostLocalInfo) error {
	if err := ensureZone(s, localCfg.Zone); err != nil {
		return errors.Wrapf(err, "create zone %s", localCfg.Zone)
	}
	if err := ensureAdminNetwork(s, localCfg.Zone, localCfg.ManagementNetInterface); err != nil {
		return errors.Wrapf(err, "create admin network")
	}
	if err := ensureRegionZone(s, cfg.Region, localCfg.Zone); err != nil {
		return errors.Wrapf(err, "create region-zone %s-%s", cfg.Region, localCfg.Zone)
	}
	if err := initScheduleData(s); err != nil {
		return errors.Wrap(err, "init sched data")
	}
	// TODO: how to inject AWS instance type json
	return nil
}

func ensureZone(s *mcclient.ClientSession, name string) error {
	_, exists, err := onecloud.IsZoneExists(s, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := onecloud.CreateZone(s, name); err != nil {
		return err
	}
	return nil
}

func ensureWire(s *mcclient.ClientSession, zone, name string, bw int) error {
	_, exists, err := onecloud.IsWireExists(s, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := onecloud.CreateWire(s, zone, name, bw, apiv1.DefaultVPCId); err != nil {
		return err
	}
	return nil
}

func ensureAdminNetwork(s *mcclient.ClientSession, zone string, iface apiv1.NetInterface) error {
	if err := ensureWire(s, zone, iface.Wire, 1000); err != nil {
		return errors.Wrapf(err, "create wire %s", iface.Wire)
	}
	name := apiv1.DefaultOnecloudAdminNetwork
	startIP := iface.Address.String()
	endIP := iface.Address.String()
	gateway := iface.Gateway.String()
	maskLen := iface.MaskLen
	if _, err := onecloud.CreateNetwork(s, name, gateway, constants.NetworkTypeBaremetal, iface.Wire, maskLen, startIP, endIP); err != nil {
		return errors.Wrapf(err, "name %q, gateway %q, %s-%s, masklen %d", name, gateway, startIP, endIP, maskLen)
	}
	return nil
}

func ensureRegionZone(s *mcclient.ClientSession, region, zone string) error {
	_, err := onecloud.CreateRegion(s, region, zone)
	return err
}

func initScheduleData(s *mcclient.ClientSession) error {
	if err := registerSchedSameProjectCloudprovider(s); err != nil {
		return err
	}
	if err := registerSchedAzureClassicHost(s); err != nil {
		return err
	}
	return nil
}

func registerSchedSameProjectCloudprovider(s *mcclient.ClientSession) error {
	obj, err := onecloud.EnsureSchedtag(s, "same_project", "prefer", "Prefer hosts belongs to same project")
	if err != nil {
		return errors.Wrap(err, "create schedtag same_project")
	}
	id, _ := obj.GetString("id")
	if _, err := onecloud.EnsureDynamicSchedtag(s, "same_cloudprovider_project", id, "host.cloudprovider.tenant_id == server.owner_tenant_id"); err != nil {
		return err
	}
	return nil
}

func registerSchedAzureClassicHost(s *mcclient.ClientSession) error {
	obj, err := onecloud.EnsureSchedtag(s, "azure_classic", "exclude", "Do not use azure classic host to create VM")
	if err != nil {
		return errors.Wrap(err, "create schedtag azure_classic")
	}
	id, _ := obj.GetString("id")
	if _, err := onecloud.EnsureDynamicSchedtag(s, "avoid_azure_classic_host", id, `host.name.endswith("-classic") && host.host_type == "azure"`); err != nil {
		return err
	}
	return nil
}

func SetupScheduler(
	s *mcclient.ClientSession,
	_ interface{},
	clusterCfg *apiv1.ClusterConfiguration,
	localCfg *apiv1.HostLocalInfo,
) error {
	url := fmt.Sprintf("https://%s:%d", localCfg.ManagementNetInterface.IPAddress(), constants.SchedulerPort)
	if err := occonfig.RegisterServicePublicInternalEndpoint(s, clusterCfg.Region, constants.ServiceNameScheduler, constants.ServiceTypeScheduler, url); err != nil {
		return err
	}
	return nil
}
