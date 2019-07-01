package controlplane

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/klog"
	"sort"

	v1 "k8s.io/api/core/v1"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/occonfig"
	staticpodutil "yunion.io/x/ocadm/pkg/util/staticpod"
)

func GetSaticPodSpecs(cfg *apiv1.ClusterConfiguration) map[string]v1.Pod {
	// Get the required hostpath mounts
	mounts := getHostPathVolumesForTheControlPlane(cfg)

	// Prepare static pod specs
	staticPodSpecs := map[string]v1.Pod{
		// keystone pod
		constants.OnecloudKeystone: staticpodutil.ComponentPodWithInit(
			&v1.Container{
				Name:            fmt.Sprintf("%s-init", constants.OnecloudKeystone),
				Image:           images.GetOnecloudImage(constants.OnecloudKeystone, cfg),
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         []string{"/opt/yunion/bin/keystone"},
				Args:            getKeystoneInitArgs(cfg.Keystone),
				VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudKeystone)),
				Resources:       staticpodutil.ComponentResources("250m"),
			},
			&v1.Container{
				Name:            constants.OnecloudKeystone,
				Image:           images.GetOnecloudImage(constants.OnecloudKeystone, cfg),
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         []string{"/opt/yunion/bin/keystone"},
				Args:            getKeystoneArgs(cfg.Keystone),
				VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudKeystone)),
				Resources:       staticpodutil.ComponentResources("250m"),
			},
			mounts.GetVolumes(constants.OnecloudKeystone),
		),

		// region pod
		constants.OnecloudRegion: staticpodutil.ComponentPodWithInit(
			nil,
			&v1.Container{
				Name:            constants.OnecloudRegion,
				Image:           images.GetOnecloudImage(constants.OnecloudRegion, cfg),
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         []string{"/opt/yunion/bin/region"},
				Args:            getRegionArgs(cfg.RegionServer),
				VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudRegion)),
				Resources:       staticpodutil.ComponentResources("250m"),
			},
			mounts.GetVolumes(constants.OnecloudRegion),
		),

		// scheduler pod
		constants.OnecloudScheduler: staticpodutil.ComponentPodWithInit(
			nil,
			&v1.Container{
				Name:            constants.OnecloudScheduler,
				Image:           images.GetOnecloudImage(constants.OnecloudScheduler, cfg),
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         []string{"/opt/yunion/bin/scheduler"},
				Args:            getRegionArgs(cfg.RegionServer),
				VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudRegion)),
				Resources:       staticpodutil.ComponentResources("1024m"),
			},
			mounts.GetVolumes(constants.OnecloudScheduler),
		),
	}

	return staticPodSpecs
}

// CreateStaticPodFiles creates all the onecloud requested static pod files.
func CreateStaticPodFiles(manifestDir string, cfg *apiv1.ClusterConfiguration, componentNames ...string) error {
	specs := GetSaticPodSpecs(cfg)

	// creates required static pod specs
	for _, componentName := range componentNames {
		// retrives the StaticPodSpec for given component
		spec, exists := specs[componentName]
		if !exists {
			return errors.Errorf("couldn't retrive StaticPodSpec for %q", componentName)
		}

		// writes the StaticPodSpec to disk
		if err := staticpodutil.WriteStaticPodToDisk(componentName, manifestDir, spec); err != nil {
			return errors.Wrapf(err, "failed to create static pod manifest file for %q", componentName)
		}

		klog.V(1).Infof("[control-plane] wrote static Pod manifest for component %q to %q\n", componentName, constants.GetStaticPodFilepath(componentName, manifestDir))
	}

	return nil
}

func BuildArgumentListFromMap(base map[string]string, override map[string]string) []string {
	var command []string
	var keys []string

	argsMap := make(map[string]string)

	for k, v := range base {
		argsMap[k] = v
	}

	for k, v := range override {
		argsMap[k] = v
	}

	for k := range argsMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		val := argsMap[k]
		cmd := fmt.Sprintf("--%s", k)
		command = append(command, cmd)
		if val != "" {
			//cmd = fmt.Sprintf("%s %s", cmd, val)
			command = append(command, val)
		}
		//command = append(command, cmd)
	}

	return command
}

func getKeystoneArgs(cfg apiv1.Keystone) []string {
	defaultArgs := map[string]string{
		"config": occonfig.KeystoneConfigFilePath(),
	}

	return BuildArgumentListFromMap(defaultArgs, nil)
}

func getKeystoneInitArgs(cfg apiv1.Keystone) []string {
	defaultArgs := map[string]string{
		"config":             occonfig.KeystoneConfigFilePath(),
		"auto-sync-table":    "",
		"exit-after-db-init": "",
	}
	if cfg.BootstrapAdminUserPassword != "" {
		defaultArgs["bootstrap-admin-user-password"] = cfg.BootstrapAdminUserPassword
	}

	return BuildArgumentListFromMap(defaultArgs, nil)
}

func getRegionArgs(cfg apiv1.RegionServer) []string {
	defaultArgs := map[string]string{
		"config": occonfig.RegionConfigFilePath(),
	}

	return BuildArgumentListFromMap(defaultArgs, nil)
}
