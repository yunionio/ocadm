package webconsole

import (
	"fmt"
	"github.com/pkg/errors"

	"yunion.io/x/onecloud/pkg/webconsole/options"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/components"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/onecloud"
	"yunion.io/x/onecloud/pkg/mcclient"
)

var (
	WebconsoleComponent *components.Component
)

func init() {
	WebconsoleComponent = &components.Component{
		Name:                 constants.OnecloudWebconsole,
		ServiceName:          constants.ServiceNameWebconsole,
		ServiceType:          constants.ServiceTypeWebconsole,
		CertConfig:           &components.CertConfig{constants.WebconsoleCertName},
		ConfigDir:            constants.OnecloudConfigDir,
		ConfigFileName:       constants.OnecloudWebconsoleConfigFileName,
		ConfigurationFactory: GetOptions,
		GetServiceAccount:    GetServiceAccount,
		SetupFunc:            SetupWebconsoleServer,
		PreUninstallFunc: func(s *mcclient.ClientSession, clusterCfg *apiv1.ClusterConfiguration, hostLocalCfg *apiv1.HostLocalInfo) error {
			return onecloud.DeleteServiceEndpoints(s, constants.ServiceNameWebconsole)
		},
	}
}

func GetOptions(authConfig *occonfig.RCAdminConfig, clusterCfg *apiv1.ClusterConfiguration, localCfg *apiv1.HostLocalInfo, certDir string) (interface{}, interface{}, error) {
	config := &apiv1.Webconsole{}
	apiv1.SetDefaults_Webconsole(config, authConfig.Region)
	configutil.SetServiceAuthInfo(&config.ServiceCommonOptions, authConfig.Region, authConfig.AuthUrl)

	opt := &options.Options
	if err := configutil.SetOptionsDefault(opt, constants.ServiceTypeWebconsole); err != nil {
		return nil, nil, err
	}
	configutil.EnableConfigTLS(&config.ServiceBaseOptions, certDir, constants.CACertName, constants.WebconsoleCertName, constants.WebconsoleKeyName)
	configutil.SetServiceCommonOptions(&opt.CommonOptions, config.ServiceCommonOptions)

	opt.IpmitoolPath = "/usr/sbin/ipmitool"
	opt.ApiServer = getLocalAPIServer(localCfg.ManagementNetInterface.IPAddress())
	opt.EnableAutoLogin = true
	return config, opt, nil
}

func GetServiceAccount(opt interface{}) *components.ServiceAccount {
	config := opt.(*apiv1.Webconsole)
	return &components.ServiceAccount{
		AdminUser:     config.ServiceCommonOptions.AdminUser,
		AdminPassword: config.ServiceCommonOptions.AdminPassword,
	}
}

func getLocalAPIServer(address string) string {
	return fmt.Sprintf("https://%s:%d", address, constants.WebconsolePort)
}

func SetupWebconsoleServer(
	s *mcclient.ClientSession,
	cfgObj interface{},
	_ *apiv1.ClusterConfiguration,
	localCfg *apiv1.HostLocalInfo,
) error {
	config := cfgObj.(*apiv1.Webconsole)
	url := getLocalAPIServer(localCfg.ManagementNetInterface.IPAddress())
	if err := occonfig.RegisterServicePublicInternalEndpoint(s, config.Region, constants.ServiceNameWebconsole, constants.ServiceTypeWebconsole, url); err != nil {
		return errors.Wrapf(err, "register service")
	}
	return nil
}
