package keystone

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"path"
	"yunion.io/x/jsonutils"
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/util/onecloud"

	"github.com/pkg/errors"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/util/mysql"
	identityapi "yunion.io/x/onecloud/pkg/apis/identity"
	"yunion.io/x/onecloud/pkg/keystone/options"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/structarg"
)

func InitDBUser(conn *mysql.Connection, info apis.DBInfo) error {
	dbName := info.Database
	dbExists, err := conn.IsDatabaseExists(dbName)
	if err != nil {
		return errors.Wrap(err, "check db exists")
	}
	if dbExists {
		return errors.Errorf("database %q already exists", dbName)
	}
	if err := conn.CreateDatabase(dbName); err != nil {
		return errors.Wrapf(err, "create database %q", dbName)
	}
	user := info.Username
	password := info.Password
	if err := conn.CreateUser(user, password, dbName); err != nil {
		return errors.Wrapf(err, "create user %q for database %q", user, dbName)
	}
	return nil
}

func GetKeystoneOptions(config apis.Keystone, certDir string) (*options.SKeystoneOptions, error) {
	opt := &options.Options
	parser, err := structarg.NewArgumentParser(opt, identityapi.SERVICE_TYPE, "", "")
	if err != nil {
		return nil, err
	}
	parser.SetDefault()
	opt.ApplicationID = identityapi.SERVICE_TYPE

	bootstrapAdminUserPasswd := config.BootstrapAdminUserPassword
	dbInfo := config.DBInfo
	opt.SqlConnection = dbInfo.ToSQLConnection()
	opt.BootstrapAdminUserPassword = bootstrapAdminUserPasswd
	opt.Port = config.ServiceBaseOptions.Port
	opt.AdminPort = config.AdminPort
	opt.EnableSsl = config.EnableSSL
	opt.SslCaCerts = path.Join(certDir, constants.CACertName)
	opt.SslCertfile = path.Join(certDir, constants.KeystoneCertName)
	opt.SslKeyfile = path.Join(certDir, constants.KeystoneKeyName)
	return opt, nil
}

func SetupKeystone(
	rootDBConn *mysql.Connection,
	config apis.Keystone,
	region string,
	localAddress string,
	certDir string,
) error {
	dbInfo := config.DBInfo
	err := InitDBUser(rootDBConn, dbInfo)
	if err != nil {
		return err
	}
	opt, err := GetKeystoneOptions(config, certDir)
	if err != nil {
		return err
	}
	if err := occonfig.WriteKeystoneConfigFile(*opt); err != nil {
		return errors.Wrap(err, "write keystone config file")
	}
	rcAdminConfig := occonfig.NewRCAdminConfig(
		fmt.Sprintf("https://%s:%d/v3", localAddress, config.ServiceBaseOptions.Port),
		region,
		config.BootstrapAdminUserPassword,
		path.Join(certDir, constants.ClimcCertName),
		path.Join(certDir, constants.ClimcKeyName),
	)
	if err := occonfig.WriteRCAdminConfigFile(rcAdminConfig); err != nil {
		return errors.Wrap(err, "write rc_admin config file")
	}
	return nil
}

func DoSysInit(s *mcclient.ClientSession, cfg *apis.ClusterConfiguration, addr string) error {
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
	adminPort := cfg.Keystone.AdminPort
	publicPort := cfg.Keystone.ServiceBaseOptions.Port
	if err := doRegisterIdentity(s, cfg.Region, addr, adminPort, publicPort, true); err != nil {
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
	if err := RolesPublic(s, PublicRoles); err != nil {
		return errors.Wrap(err, "public roles")
	}
	if err := PoliciesPublic(s, PublicPolicies); err != nil {
		return errors.Wrap(err, "public policies")
	}
	return nil
}

func doCreateRegion(s *mcclient.ClientSession, region string) (jsonutils.JSONObject, error) {
	obj, err := modules.Regions.Get(s, region, nil)
	if err == nil {
		// region already exists
		return obj, nil
	}
	if !onecloud.IsNotFoundError(err) {
		return nil, err
	}
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(region), "id")
	return modules.Regions.Create(s, params)
}

func doRegisterServiceEndpointByInterfaces(
	s *mcclient.ClientSession,
	regionId string,
	serviceName string,
	serviceType string,
	endpointUrl string,
	interfaces []string,
) error {
	svc, err := onecloud.EnsureService(s, serviceName, serviceType)
	if err != nil {
		return err
	}
	svcId, err := svc.GetString("id")
	if err != nil {
		return err
	}
	errgrp := &errgroup.Group{}
	for _, inf := range interfaces {
		tmpInf := inf
		errgrp.Go(func() error {
			_, err = onecloud.EnsureEndpoint(s, svcId, regionId, tmpInf, endpointUrl)
			if err != nil {
				return err
			}
			return nil
		})
	}
	return errgrp.Wait()
}

func doRegisterServicePublicInternalEndpoint(
	s *mcclient.ClientSession,
	regionId string,
	serviceName string,
	serviceType string,
	endpointUrl string,
) error {
	return doRegisterServiceEndpointByInterfaces(s, regionId, serviceName, serviceType,
		endpointUrl, []string{constants.EndpointTypeInternal, constants.EndpointTypePublic})
}

func doRegisterCloudMeta(s *mcclient.ClientSession, regionId string) error {
	return doRegisterServicePublicInternalEndpoint(s, regionId,
		constants.ServiceNameCloudmeta,
		constants.ServiceTypeCloudmeta,
		constants.ServiceURLCloudmeta)
}

func doRegisterTracker(s *mcclient.ClientSession, regionId string) error {
	return doRegisterServicePublicInternalEndpoint(
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
	if err := doRegisterServicePublicInternalEndpoint(
		s, regionId, constants.ServiceNameKeystone,
		constants.ServiceTypeIdentity, publicUrl); err != nil {
		return errors.Wrapf(err, "register keystone public endpoint %s", publicUrl)
	}
	if err := doRegisterServiceEndpointByInterfaces(
		s, regionId, constants.ServiceNameKeystone, constants.ServiceTypeIdentity,
		adminUrl, []string{constants.EndpointTypeAdmin}); err != nil {
		return errors.Wrapf(err, "register keystone admin endpoint %s", adminUrl)
	}
	return nil
}

func makeDomainAdminPublic(s *mcclient.ClientSession) error {
	if err := RolesPublic(s, []string{RoleDomainAdmin}); err != nil {
		return err
	}
	if err := PoliciesPublic(s, []string{PolicyTypeDomainAdmin}); err != nil {
		return err
	}
	return nil
}
