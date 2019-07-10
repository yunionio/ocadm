package components

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/images"
	"yunion.io/x/ocadm/pkg/occonfig"
	staticpodutil "yunion.io/x/ocadm/pkg/util/staticpod"
)

var (
	True = true
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
				Args:            getKeystoneInitArgs(cfg.BootstrapPassword),
				VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudKeystone)),
				Resources:       staticpodutil.ComponentResources("250m"),
			},
			&v1.Container{
				Name:            constants.OnecloudKeystone,
				Image:           images.GetOnecloudImage(constants.OnecloudKeystone, cfg),
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         []string{"/opt/yunion/bin/keystone"},
				Args:            getKeystoneArgs(),
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
				Args:            getRegionArgs(),
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
				Args:            getRegionArgs(),
				VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudRegion)),
				Resources:       staticpodutil.ComponentResources("1024m"),
			},
			mounts.GetVolumes(constants.OnecloudScheduler),
		),

		// glance pod
		constants.OnecloudGlance: staticpodutil.ComponentPodWithInit(
			nil,
			&v1.Container{
				Name:            constants.OnecloudGlance,
				Image:           images.GetOnecloudImage(constants.OnecloudGlance, cfg),
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         []string{"/opt/yunion/bin/glance"},
				Args:            getGlanceArgs(),
				VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudGlance)),
				Resources:       staticpodutil.ComponentResources("1024m"),
			},
			mounts.GetVolumes(constants.OnecloudGlance),
		),

		// baremetal agent pod
		constants.OnecloudBaremetal: staticpodutil.ComponentPodWithHostIPC(
			&v1.Container{
				Name:         constants.OnecloudBaremetal,
				Image:        images.GetOnecloudImage(constants.OnecloudBaremetalAgent, cfg),
				Command:      []string{"/opt/yunion/bin/baremetal-agent"},
				Args:         []string{"--config", occonfig.BaremetalConfigFilePath()},
				VolumeMounts: staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudBaremetal)),
				SecurityContext: &v1.SecurityContext{
					Privileged: &True,
				},
			},
			mounts.GetVolumes(constants.OnecloudBaremetal),
		),

		// webconsole pod
		constants.OnecloudWebconsole: staticpodutil.ComponentPodWithInit(
			nil,
			&v1.Container{
				Name:            constants.OnecloudWebconsole,
				Image:           images.GetOnecloudImage(constants.OnecloudWebconsole, cfg),
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         []string{"/opt/yunion/bin/webconsole"},
				Args:            []string{"--config", occonfig.WebconsoleConfigFilePath()},
				VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(constants.OnecloudWebconsole)),
				Resources:       staticpodutil.ComponentResources("250m"),
			},
			mounts.GetVolumes(constants.OnecloudWebconsole),
		),

		// influxdb pod
		constants.OnecloudInfluxdb: staticpodutil.ComponentPodWithInit(
			nil,
			&v1.Container{
				Name:            constants.OnecloudInfluxdb,
				Image:           images.GetGenericImage(cfg.ImageRepository, constants.OnecloudInfluxdb, "1.7.7"),
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         []string{"influxd", "-config", "/etc/influxdb/influxdb.conf"},
			},
			mounts.GetVolumes(constants.OnecloudInfluxdb),
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

func getKeystoneArgs() []string {
	defaultArgs := map[string]string{
		"config": occonfig.KeystoneConfigFilePath(),
	}

	return BuildArgumentListFromMap(defaultArgs, nil)
}

func getKeystoneInitArgs(bootstrapPassword string) []string {
	defaultArgs := map[string]string{
		"config":             occonfig.KeystoneConfigFilePath(),
		"auto-sync-table":    "",
		"exit-after-db-init": "",
	}
	defaultArgs["bootstrap-admin-user-password"] = bootstrapPassword

	return BuildArgumentListFromMap(defaultArgs, nil)
}

func getRegionArgs() []string {
	defaultArgs := map[string]string{
		"config": occonfig.RegionConfigFilePath(),
	}

	return BuildArgumentListFromMap(defaultArgs, nil)
}

func getGlanceArgs() []string {
	defaultArgs := map[string]string{
		"config": occonfig.GlanceConfigFilePath(),
	}
	return BuildArgumentListFromMap(defaultArgs, nil)
}
