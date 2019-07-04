package glance

import (
	"fmt"

	"github.com/pkg/errors"

	"yunion.io/x/onecloud/pkg/image/options"
	"yunion.io/x/onecloud/pkg/mcclient"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/components"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/onecloud"
)

var (
	GlanceComponent *components.Component
)

func init() {
	GlanceComponent = &components.Component{
		Name:                 constants.OnecloudGlance,
		ServiceName:          constants.ServiceNameGlance,
		ServiceType:          constants.ServiceTypeGlance,
		CertConfig:           &components.CertConfig{constants.GlanceCertName},
		ConfigDir:            constants.OnecloudGlanceConfigDir,
		ConfigFileName:       constants.OnecloudGlanceConfigFileName,
		ConfigurationFactory: GetGlanceOptions,
		GetDBInfo:            GetDBInfo,
		GetServiceAccount:    GetServiceAccount,
		SetupFunc:            SetupGlanceServer,
		WaitRunningFunc: func(waiter onecloud.Waiter) error {
			return waiter.WaitForGlance()
		},
		PreUninstallFunc: func(s *mcclient.ClientSession, clusterCfg *apiv1.ClusterConfiguration, hostLocalCfg *apiv1.HostLocalInfo) error {
			return onecloud.DeleteServiceEndpoints(s, constants.ServiceNameGlance)
		},
	}
}

func GetGlanceOptions(authConfig *occonfig.RCAdminConfig, clusterCfg *apiv1.ClusterConfiguration, _ *apiv1.HostLocalInfo, certDir string) (interface{}, interface{}, error) {
	config := &apiv1.Glance{}
	apiv1.SetDefaults_Glance(config, authConfig.Region)
	configutil.SetServiceDBInfo(&config.ServiceDBOptions.DBInfo, &clusterCfg.MysqlConnection, constants.GlanceDB, constants.GlanceDBUser)
	configutil.SetServiceAuthInfo(&config.ServiceCommonOptions, authConfig.Region, authConfig.AuthUrl)

	opt := &options.Options
	if err := configutil.SetOptionsDefault(opt, constants.ServiceTypeGlance); err != nil {
		return nil, nil, err
	}

	configutil.SetDBOptions(&opt.DBOptions, config.ServiceDBOptions)

	configutil.EnableConfigTLS(&config.ServiceBaseOptions, certDir, constants.CACertName, constants.GlanceCertName, constants.GlanceKeyName)

	configutil.SetServiceCommonOptions(&opt.CommonOptions, config.ServiceCommonOptions)

	opt.AutoSyncTable = true
	opt.FilesystemStoreDatadir = config.FilesystemStoreDatadir
	opt.TorrentStoreDir = config.TorrentStoreDir
	opt.EnableTorrentService = config.EnableTorrentService

	return config, opt, nil
}

func GetDBInfo(cfg interface{}) *apiv1.DBInfo {
	return &cfg.(*apiv1.Glance).ServiceDBOptions.DBInfo
}

func GetServiceAccount(opt interface{}) *components.ServiceAccount {
	config := opt.(*apiv1.Glance)
	return &components.ServiceAccount{
		AdminUser:     config.ServiceCommonOptions.AdminUser,
		AdminPassword: config.ServiceCommonOptions.AdminPassword,
	}
}

func SetupGlanceServer(
	s *mcclient.ClientSession,
	cfgObj interface{},
	_ *apiv1.ClusterConfiguration,
	localCfg *apiv1.HostLocalInfo,
) error {
	config := cfgObj.(*apiv1.Glance)
	url := fmt.Sprintf("https://%s:%d", localCfg.ManagementNetInterface.IPAddress(), config.Port)
	if err := occonfig.RegisterServicePublicInternalEndpoint(s, config.Region, constants.ServiceNameGlance, constants.ServiceTypeGlance, url); err != nil {
		return errors.Wrap(err, "register service")
	}
	return nil
}
