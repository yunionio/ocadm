package webconsole

import (
	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/components"
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
		PreUninstallFunc: func(s *mcclient.ClientSession, clusterCfg *apiv1.ClusterConfiguration, hostLocalCfg *apiv1.HostLocalInfo) error {
			return onecloud.DeleteServiceEndpoints(s, constants.ServiceNameWebconsole)

		},
	}
}

func GetOptions(authConfig *occonfig.RCAdminConfig, clusterCfg *apiv1.ClusterConfiguration, _ *apiv1.HostLocalInfo, certDir string) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func GetServiceAccount(opt interface{}) *components.ServiceAccount {
	return nil
}
