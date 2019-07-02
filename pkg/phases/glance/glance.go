package glance

import (
	"fmt"

	"github.com/pkg/errors"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	configutil "yunion.io/x/ocadm/pkg/util/config"
	"yunion.io/x/ocadm/pkg/util/mysql"

	"yunion.io/x/onecloud/pkg/image/options"
	"yunion.io/x/onecloud/pkg/mcclient"
)

func GetGlanceOptions(config apis.Glance, certDir string) (*options.SImageOptions, error) {
	opt := &options.Options
	if err := configutil.SetOptionsDefault(opt, constants.ServiceTypeGlance); err != nil {
		return nil, err
	}

	configutil.SetDBOptions(&opt.DBOptions, config.ServiceDBOptions)

	configutil.EnableConfigTLS(&config.ServiceBaseOptions, certDir, constants.CACertName, constants.GlanceCertName, constants.GlanceKeyName)

	configutil.SetServiceCommonOptions(&opt.CommonOptions, config.ServiceCommonOptions)

	opt.AutoSyncTable = true
	opt.FilesystemStoreDatadir = config.FilesystemStoreDatadir
	opt.TorrentStoreDir = config.TorrentStoreDir
	opt.EnableTorrentService = config.EnableTorrentService

	return opt, nil
}

func SetupGlanceServer(
	s *mcclient.ClientSession,
	rootDBConn *mysql.Connection,
	config apis.Glance,
	certDir string,
	address string,
) error {
	if err := configutil.InitDBUser(rootDBConn, config.ServiceDBOptions.DBInfo); err != nil {
		return err
	}
	opt, err := GetGlanceOptions(config, certDir)
	if err != nil {
		return err
	}
	if err := occonfig.WriteGlanceConfigFile(*opt); err != nil {
		return errors.Wrap(err, "write glance config file")
	}

	if err := occonfig.InitServiceAccount(s, config.AdminUser, config.AdminPassword); err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s:%d", address, config.Port)
	if err := occonfig.RegisterServicePublicInternalEndpoint(s, config.Region, constants.ServiceNameGlance, constants.ServiceTypeGlance, url); err != nil {
		return errors.Wrap(err, "register service")
	}
	return nil
}
