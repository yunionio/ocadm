package region

import (
	"fmt"
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/onecloud/pkg/compute/options"
	"yunion.io/x/onecloud/pkg/mcclient"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/mysql"
)

func GetRegionOptions(config apis.RegionServer, certDir string) (*options.ComputeOptions, error) {
	opt := &options.Options
	if err := configutil.SetOptionsDefault(opt, constants.ServiceTypeCompute); err != nil {
		return nil, err
	}
	configutil.SetDBOptions(&opt.DBOptions, config.ServiceDBOptions)

	configutil.EnableConfigTLS(&config.ServiceBaseOptions, certDir, constants.CACertName, constants.RegionCertName, constants.RegionKeyName)

	configutil.SetServiceCommonOptions(&opt.CommonOptions, config.ServiceCommonOptions)

	opt.AutoSyncTable = true

	return opt, nil
}

func SetupRegionServer(
	s *mcclient.ClientSession,
	rootDBConn *mysql.Connection,
	config apis.RegionServer,
	certDir string,
	address string,
) error {
	if err := configutil.InitDBUser(rootDBConn, config.DBInfo); err != nil {
		return err
	}
	opt, err := GetRegionOptions(config, certDir)
	if err != nil {
		return err
	}
	if err := occonfig.WriteRegionConfigFile(*opt); err != nil {
		return err
	}

	if err := occonfig.InitServiceAccount(s, config.AdminUser, config.AdminPassword); err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s:%d", address, config.PortV2)
	if err := occonfig.RegisterServicePublicInternalEndpoint(s, config.Region, constants.ServiceNameRegionV2, constants.ServiceTypeComputeV2, url); err != nil {
		return err
	}
	return nil
}
