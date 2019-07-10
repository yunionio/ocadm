package influxdb

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
	"yunion.io/x/ocadm/pkg/phases/components"
	"yunion.io/x/ocadm/pkg/util/onecloud"
	"yunion.io/x/onecloud/pkg/mcclient"
)

var (
	InfluxdbComponent *components.Component
)

func init() {
	InfluxdbComponent = &components.Component{
		Name:           constants.OnecloudInfluxdb,
		ServiceName:    constants.ServiceNameInfluxdb,
		ServiceType:    constants.ServiceTypeInfluxdb,
		CertConfig:     &components.CertConfig{constants.InfluxdbCertName},
		ConfigDir:      constants.OnecloudConfigDir,
		ConfigFileName: constants.OnecloudInfluxdbConfigFileName,
		UseSession:     true,
		SetupFunc:      SetupInfluxdb,
		PreUninstallFunc: func(s *mcclient.ClientSession, clusterCfg *apiv1.ClusterConfiguration, hostLocalCfg *apiv1.HostLocalInfo) error {
			return onecloud.DeleteServiceEndpoints(s, constants.ServiceNameInfluxdb)
		},
	}
}

func SetupInfluxdb(s *mcclient.ClientSession, _ interface{}, clusterCfg *apiv1.ClusterConfiguration, localCfg *apiv1.HostLocalInfo) error {
	config := Config{
		CertPath: path.Join(clusterCfg.OnecloudCertificatesDir, constants.InfluxdbCertName),
		KeyPath:  path.Join(clusterCfg.OnecloudCertificatesDir, constants.InfluxdbKeyName),
	}
	content, err := config.GetContent()
	if err != nil {
		return err
	}
	if err := occonfig.WriteOnecloudFile(path.Join(constants.OnecloudConfigDir, constants.OnecloudInfluxdbConfigFileName), content); err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s:%d", localCfg.ManagementNetInterface.IPAddress(), constants.InfluxdbPort)
	if err := occonfig.RegisterServicePublicInternalEndpoint(s, clusterCfg.Region, constants.ServiceNameInfluxdb, constants.ServiceTypeInfluxdb, url); err != nil {
		return errors.Wrapf(err, "register service")
	}
	return nil
}
