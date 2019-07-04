package baremetal

import (
	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/components"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/onecloud/pkg/baremetal/options"
)

var BaremetalComponent *components.Component

func init() {
	BaremetalComponent = &components.Component{
		Name:                 constants.OnecloudBaremetal,
		CertConfig:           &components.CertConfig{constants.BaremetalCertName},
		ConfigDir:            constants.OnecloudConfigDir,
		ConfigFileName:       constants.OnecloudBaremetalConfigFileName,
		ConfigurationFactory: GetBaremetalOptions,
		GetDBInfo:            nil,
		GetServiceAccount:    GetServiceAccount,
		WaitRunningFunc:      nil,
		SysInitFunc:          nil,
	}
}

func GetServiceAccount(configOpt interface{}) *components.ServiceAccount {
	config := configOpt.(*apiv1.BaremetalAgent)
	return &components.ServiceAccount{
		AdminUser:     config.ServiceCommonOptions.AdminUser,
		AdminPassword: config.ServiceCommonOptions.AdminPassword,
	}
}

func GetBaremetalOptions(authConfig *occonfig.RCAdminConfig, _ *apiv1.ClusterConfiguration, localCfg *apiv1.HostLocalInfo, certDir string) (interface{}, interface{}, error) {
	config := &apiv1.BaremetalAgent{}
	apiv1.SetDefaults_BaremetalAgent(config, authConfig.Region)
	configutil.SetServiceAuthInfo(&config.ServiceCommonOptions, authConfig.Region, authConfig.AuthUrl)

	opt := &options.Options
	if err := configutil.SetOptionsDefault(opt, "baremetal"); err != nil {
		return nil, nil, err
	}
	configutil.EnableConfigTLS(&config.ServiceBaseOptions, certDir, constants.CACertName, constants.BaremetalCertName, constants.BaremetalKeyName)
	configutil.SetServiceCommonOptions(&opt.CommonOptions, config.ServiceCommonOptions)

	opt.ListenInterface = localCfg.ManagementNetInterface.Interface
	opt.AutoRegisterBaremetal = config.AutoRegisterBaremetal
	opt.LinuxDefaultRootUser = config.LinuxDefaultRootUser
	opt.DefaultIpmiPassword = config.DefaultIPMIPassword
	opt.EnableTftpHttpDownload = config.EnableTFTPHTTPDownload
	opt.TftpRoot = constants.OnecloudBaremetalTFTPRoot
	opt.BaremetalsPath = constants.OnecloudBaremetalsPath
	return config, opt, nil
}
