package longhorn

import (
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

type LonghornConfig struct {
	DataPath                   string
	OverProvisioningPercentage int
	ReplicaCount               int

	LonghornStorageClass         string
	LonghornManagerImage         string
	LonghornEngineImage          string
	LonghornInstanceManagerImage string
	LonghornUiImage              string
}

func NewLonghornConfig(repo, dataPath string, overProvisioningPercentage, replicaCount int) addons.Configer {
	return &LonghornConfig{
		DataPath:                     dataPath,
		OverProvisioningPercentage:   overProvisioningPercentage,
		ReplicaCount:                 replicaCount,
		LonghornManagerImage:         images.GetGenericImage(repo, constants.LonghornManager, constants.DefaultLonghornVersion),
		LonghornEngineImage:          images.GetGenericImage(repo, constants.LonghornEngine, constants.DefaultLonghornVersion),
		LonghornInstanceManagerImage: images.GetGenericImage(repo, constants.LonghornInstanceManager, constants.DefaultLonghornVersion),
		LonghornUiImage:              images.GetGenericImage(repo, constants.LonghornUi, constants.DefaultLonghornVersion),
		LonghornStorageClass:         constants.LonghornStorageClass,
	}
}

func (c LonghornConfig) Name() string {
	return "longhorn"
}

func (c LonghornConfig) GenerateYAML() (string, error) {
	return addons.CompileTemplateFromMap(LonghornTemplate, c)
}
