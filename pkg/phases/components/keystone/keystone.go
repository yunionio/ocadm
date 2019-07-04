package keystone

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"yunion.io/x/jsonutils"
	"yunion.io/x/ocadm/pkg/util/onecloud"
	identityapi "yunion.io/x/onecloud/pkg/apis/identity"
	"yunion.io/x/onecloud/pkg/keystone/options"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/components"
	configutil "yunion.io/x/ocadm/pkg/util/config"
)

var KeystoneComponent *components.Component

func init() {
	KeystoneComponent = &components.Component{
		Name:        constants.OnecloudKeystone,
		ServiceName: constants.ServiceNameKeystone,
		ServiceType: constants.ServiceTypeIdentity,
		CertConfig: &components.CertConfig{
			CertName: constants.KeystoneCertName,
		},
		ConfigDir:            constants.OnecloudKeystoneConfigDir,
		ConfigFileName:       constants.OnecloudKeystoneConfigFileName,
		ConfigurationFactory: GetKeystoneOptions,
		GetDBInfo:            GetDBInfo,
		GetServiceAccount:    nil,
		SetupFunc:            SetupKeystone,
		WaitRunningFunc: func(waiter onecloud.Waiter) error {
			return waiter.WaitForKeystone()
		},
		SysInitFunc: DoSysInit,
	}
}

func GetKeystoneOptions(_ *occonfig.RCAdminConfig, clusterCfg *apiv1.ClusterConfiguration, _ *apiv1.HostLocalInfo, certDir string) (interface{}, interface{}, error) {
	config := &apiv1.Keystone{}
	apiv1.SetDefaults_Keystone(config)
	configutil.SetServiceDBInfo(&config.ServiceDBOptions.DBInfo, &clusterCfg.MysqlConnection, constants.KeystoneDB, constants.KeystoneDBUser)

	opt := &options.Options
	if err := configutil.SetOptionsDefault(opt, identityapi.SERVICE_TYPE); err != nil {
		return nil, nil, err
	}
	configutil.SetDBOptions(&opt.DBOptions, config.ServiceDBOptions)

	configutil.EnableConfigTLS(&config.ServiceBaseOptions, certDir, constants.CACertName, constants.KeystoneCertName, constants.KeystoneKeyName)

	configutil.SetServiceBaseOptions(&opt.BaseOptions, config.ServiceBaseOptions)

	bootstrapAdminUserPasswd := clusterCfg.BootstrapPassword
	opt.BootstrapAdminUserPassword = bootstrapAdminUserPasswd
	opt.AdminPort = config.AdminPort

	return config, opt, nil
}

func GetDBInfo(obj interface{}) *apiv1.DBInfo {
	return &obj.(*apiv1.Keystone).ServiceDBOptions.DBInfo
}

func SetupKeystone(_ *mcclient.ClientSession, configObj interface{}, clusterCfg *apiv1.ClusterConfiguration, localCfg *apiv1.HostLocalInfo) error {
	config := configObj.(*apiv1.Keystone)
	certDir := clusterCfg.OnecloudCertificatesDir
	rcAdminConfig := occonfig.NewRCAdminConfig(
		configutil.GetAuthURL(*config, localCfg.ManagementNetInterface.IPAddress()),
		clusterCfg.Region,
		clusterCfg.BootstrapPassword,
		path.Join(certDir, constants.ClimcCertName),
		path.Join(certDir, constants.ClimcKeyName),
	)
	if err := occonfig.WriteRCAdminConfigFile(rcAdminConfig); err != nil {
		return errors.Wrap(err, "write rc_admin config file")
	}
	return nil
}

func DoSysInit(s *mcclient.ClientSession, cfg *apiv1.ClusterConfiguration, localCfg *apiv1.HostLocalInfo) error {
	if err := doPolicyRoleInit(s); err != nil {
		return errors.Wrap(err, "policy role init")
	}
	if _, err := doCreateRegion(s, cfg.Region); err != nil {
		return errors.Wrap(err, "create region")
	}

	if err := doRegisterCloudMeta(s, cfg.Region); err != nil {
		return errors.Wrap(err, "register cloudmeta endpoint")
	}
	if err := doRegisterTracker(s, cfg.Region); err != nil {
		return errors.Wrap(err, "register tracker endpoint")
	}
	keystoneCfgObj, err := occonfig.GetConfigFileObject(occonfig.KeystoneConfigFilePath())
	if err != nil {
		return err
	}
	adminPort, err := keystoneCfgObj.Int("admin_port")
	if err != nil {
		return err
	}
	publicPort, err := keystoneCfgObj.Int("port")
	if err != nil {
		return err
	}
	if err := doRegisterIdentity(s, cfg.Region, localCfg.ManagementNetInterface.Address.String(), int(adminPort), int(publicPort), true); err != nil {
		return errors.Wrap(err, "register identity endpoint")
	}
	if err := makeDomainAdminPublic(s); err != nil {
		return errors.Wrap(err, "always share domainadmin")
	}
	return nil
}

func shouldDoPolicyRoleInit(s *mcclient.ClientSession) (bool, error) {
	ret, err := modules.Policies.List(s, nil)
	if err != nil {
		return false, errors.Wrap(err, "list policy")
	}
	return ret.Total == 0, nil
}

func doPolicyRoleInit(s *mcclient.ClientSession) error {
	doInit, err := shouldDoPolicyRoleInit(s)
	if err != nil {
		return errors.Wrap(err, "should do policy init")
	}
	if !doInit {
		return nil
	}
	fmt.Println("Init policy and role...")
	for policyType, content := range DefaultPolicies {
		if _, err := PolicyCreate(s, policyType, content, true); err != nil {
			return errors.Wrapf(err, "create policy %s", policyType)
		}
	}
	for role, desc := range DefaultRoles {
		if _, err := onecloud.EnsureRole(s, role, desc); err != nil {
			return errors.Wrapf(err, "create role %s", role)
		}
	}
	if err := RolesPublic(s, constants.PublicRoles); err != nil {
		return errors.Wrap(err, "public roles")
	}
	if err := PoliciesPublic(s, constants.PublicPolicies); err != nil {
		return errors.Wrap(err, "public policies")
	}
	return nil
}

func doCreateRegion(s *mcclient.ClientSession, region string) (jsonutils.JSONObject, error) {
	return onecloud.CreateRegion(s, region, "")
}

func doRegisterCloudMeta(s *mcclient.ClientSession, regionId string) error {
	return occonfig.RegisterServicePublicInternalEndpoint(s, regionId,
		constants.ServiceNameCloudmeta,
		constants.ServiceTypeCloudmeta,
		constants.ServiceURLCloudmeta)
}

func doRegisterTracker(s *mcclient.ClientSession, regionId string) error {
	return occonfig.RegisterServicePublicInternalEndpoint(
		s, regionId,
		constants.ServiceNameTorrentTracker,
		constants.ServiceTypeTorrentTracker,
		constants.ServiceURLTorrentTracker)
}

func doRegisterIdentity(
	s *mcclient.ClientSession,
	regionId string,
	keystoneAddress string,
	adminPort int,
	publicPort int,
	enableSSL bool,
) error {
	proto := "http"
	if enableSSL {
		proto = "https"
	}
	genUrl := func(port int) string {
		return fmt.Sprintf("%s://%s:%d/v3", proto, keystoneAddress, port)
	}
	publicUrl := genUrl(publicPort)
	adminUrl := genUrl(adminPort)
	if err := occonfig.RegisterServicePublicInternalEndpoint(
		s, regionId, constants.ServiceNameKeystone,
		constants.ServiceTypeIdentity, publicUrl); err != nil {
		return errors.Wrapf(err, "register keystone public endpoint %s", publicUrl)
	}
	if err := occonfig.RegisterServiceEndpointByInterfaces(
		s, regionId, constants.ServiceNameKeystone, constants.ServiceTypeIdentity,
		adminUrl, []string{constants.EndpointTypeAdmin}); err != nil {
		return errors.Wrapf(err, "register keystone admin endpoint %s", adminUrl)
	}
	return nil
}

func makeDomainAdminPublic(s *mcclient.ClientSession) error {
	if err := RolesPublic(s, []string{constants.RoleDomainAdmin}); err != nil {
		return err
	}
	if err := PoliciesPublic(s, []string{constants.PolicyTypeDomainAdmin}); err != nil {
		return err
	}
	return nil
}
