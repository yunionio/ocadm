package components

import (
	"fmt"

	"yunion.io/x/ocadm/pkg/apis/constants"

	v1 "k8s.io/api/core/v1"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	staticpodutil "k8s.io/kubernetes/cmd/kubeadm/app/util/staticpod"

	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/occonfig"
)

func getHostPathVolumesForTheControlPlane(cfg *apiv1.ClusterConfiguration) controlPlaneHostPathMounts {
	hostPathDirectoryOrCreate := v1.HostPathDirectoryOrCreate
	hostPathFileOrCreate := v1.HostPathFileOrCreate
	mounts := newControlPlaneHostPathMounts()

	// HostPath volumes for the keystone
	mounts.NewHostPathMount(
		constants.OnecloudKeystone,
		constants.OnecloudEtcKeystoneVolumeName,
		constants.OnecloudKeystoneConfigDir,
		constants.OnecloudKeystoneConfigDir,
		false,
		&hostPathDirectoryOrCreate,
	)
	mounts.NewHostPathMount(
		constants.OnecloudKeystone,
		constants.OnecloudOptTmpVolumeName,
		constants.OnecloudOptTmpDir,
		constants.OnecloudOptTmpDir,
		false,
		&hostPathDirectoryOrCreate,
	)
	// Read-only mount for the keystone config file
	keystoneConfigFile := occonfig.KeystoneConfigFilePath()
	mounts.NewHostPathMount(
		constants.OnecloudKeystone,
		constants.OnecloudConfigVolumeName,
		keystoneConfigFile,
		keystoneConfigFile,
		true,
		&hostPathFileOrCreate,
	)
	// Read-only mount for the keystone certs
	mounts.NewHostPathMount(
		constants.OnecloudKeystone,
		constants.OnecloudPKICertsVolumeName,
		cfg.OnecloudCertificatesDir,
		cfg.OnecloudCertificatesDir,
		true,
		&hostPathDirectoryOrCreate,
	)

	// Read-only mount for the keystone config file
	regionConfigFile := occonfig.RegionConfigFilePath()
	mounts.NewHostPathMount(
		constants.OnecloudRegion,
		constants.OnecloudConfigVolumeName,
		regionConfigFile,
		regionConfigFile,
		true,
		&hostPathFileOrCreate,
	)
	// Read-only mount for the region certs
	mounts.NewHostPathMount(
		constants.OnecloudRegion,
		constants.OnecloudPKICertsVolumeName,
		cfg.OnecloudCertificatesDir,
		cfg.OnecloudCertificatesDir,
		true,
		&hostPathDirectoryOrCreate,
	)

	// Read-only mount for the scheduler config file
	mounts.NewHostPathMount(
		constants.OnecloudScheduler,
		constants.OnecloudConfigVolumeName,
		regionConfigFile,
		regionConfigFile,
		true,
		&hostPathFileOrCreate,
	)
	// Read-only mount for the scheduler certs
	mounts.NewHostPathMount(
		constants.OnecloudScheduler,
		constants.OnecloudPKICertsVolumeName,
		cfg.OnecloudCertificatesDir,
		cfg.OnecloudCertificatesDir,
		true,
		&hostPathDirectoryOrCreate,
	)

	// Read-only mount for the glance config file
	glanceConfigFile := occonfig.GlanceConfigFilePath()
	mounts.NewHostPathMount(
		constants.OnecloudGlance,
		constants.OnecloudConfigVolumeName,
		glanceConfigFile,
		glanceConfigFile,
		true,
		&hostPathFileOrCreate,
	)

	// Read-only mount for the glance certs
	mounts.NewHostPathMount(
		constants.OnecloudGlance,
		constants.OnecloudPKICertsVolumeName,
		cfg.OnecloudCertificatesDir,
		cfg.OnecloudCertificatesDir,
		true,
		&hostPathDirectoryOrCreate,
	)

	// Read-Write mount for glance images
	mounts.NewHostPathMount(
		constants.OnecloudGlance,
		constants.OnecloudGlanceImageVolumeName,
		constants.OnecloudGlanceFileStoreDir,
		constants.OnecloudGlanceFileStoreDir,
		false,
		&hostPathDirectoryOrCreate,
	)

	// Read-only mount for the glance probe image
	mounts.NewHostPathMount(
		constants.OnecloudGlance,
		constants.OnecloudQemuBinaryVolumeName,
		constants.OnecloudQemuPath,
		constants.OnecloudQemuPath,
		true,
		&hostPathDirectoryOrCreate,
	)

	// Read-only mount for the glance probe image
	mounts.NewHostPathMount(
		constants.OnecloudGlance,
		constants.OnecloudKernelVolumeName,
		constants.OnecloudKernelPath,
		constants.OnecloudKernelPath,
		true,
		&hostPathDirectoryOrCreate,
	)

	// Read-only mount for the baremetal config file
	baremetalConfigFile := occonfig.BaremetalConfigFilePath()
	mounts.NewHostPathMount(
		constants.OnecloudBaremetal,
		constants.OnecloudConfigVolumeName,
		baremetalConfigFile,
		baremetalConfigFile,
		true,
		&hostPathFileOrCreate,
	)
	// Read-only mount for the region certs
	mounts.NewHostPathMount(
		constants.OnecloudBaremetal,
		constants.OnecloudPKICertsVolumeName,
		cfg.OnecloudCertificatesDir,
		cfg.OnecloudCertificatesDir,
		true,
		&hostPathDirectoryOrCreate,
	)
	// Read-only mount for the baremetal tftp root
	mounts.NewHostPathMount(
		constants.OnecloudBaremetal,
		constants.OnecloudBaremetalTFTPVolumeName,
		constants.OnecloudBaremetalTFTPRoot,
		constants.OnecloudBaremetalTFTPRoot,
		true,
		&hostPathDirectoryOrCreate,
	)
	// Read-write for baremetal desc save path
	mounts.NewHostPathMount(
		constants.OnecloudBaremetal,
		constants.OnecloudBaremetalsVolumeName,
		constants.OnecloudBaremetalsPath,
		constants.OnecloudBaremetalsPath,
		false,
		&hostPathDirectoryOrCreate,
	)

	return mounts
}

// controlPlaneHostPathMounts is a helper struct for handling all the control plane's hostPath mounts in an easy way
type controlPlaneHostPathMounts struct {
	// volumes is a nested map that forces a unique volumes. The outer map's
	// keys are a string that should specify the target component to add the
	// volume to. The values (inner map) of the outer map are maps with string
	// keys and v1.Volume values. The inner map's key should specify the volume
	// name.
	volumes map[string]map[string]v1.Volume
	// volumeMounts is a nested map that forces a unique volume mounts. The
	// outer map's keys are a string that should specify the target component
	// to add the volume mount to. The values (inner map) of the outer map are
	// maps with string keys and v1.VolumeMount values. The inner map's key
	// should specify the volume mount name.
	volumeMounts map[string]map[string]v1.VolumeMount
}

func newControlPlaneHostPathMounts() controlPlaneHostPathMounts {
	return controlPlaneHostPathMounts{
		volumes:      map[string]map[string]v1.Volume{},
		volumeMounts: map[string]map[string]v1.VolumeMount{},
	}
}

func (c *controlPlaneHostPathMounts) NewHostPathMount(component, mountName, hostPath, containerPath string, readOnly bool, hostPathType *v1.HostPathType) {
	vol := staticpodutil.NewVolume(mountName, hostPath, hostPathType)
	c.addComponentVolume(component, vol)
	volMount := staticpodutil.NewVolumeMount(mountName, containerPath, readOnly)
	c.addComponentVolumeMount(component, volMount)
}

func (c *controlPlaneHostPathMounts) AddHostPathMounts(component string, vols []v1.Volume, volMounts []v1.VolumeMount) {
	for _, v := range vols {
		c.addComponentVolume(component, v)
	}
	for _, v := range volMounts {
		c.addComponentVolumeMount(component, v)
	}
}

// AddExtraHostPathMounts adds host path mounts and overwrites the default
// paths in the case that a user specifies the same volume/volume mount name.
func (c *controlPlaneHostPathMounts) AddExtraHostPathMounts(component string, extraVols []kubeadmapi.HostPathMount) {
	for _, extraVol := range extraVols {
		fmt.Printf("[controlplane] Adding extra host path mount %q to %q\n", extraVol.Name, component)
		hostPathType := extraVol.PathType
		c.NewHostPathMount(component, extraVol.Name, extraVol.HostPath, extraVol.MountPath, extraVol.ReadOnly, &hostPathType)
	}
}

func (c *controlPlaneHostPathMounts) GetVolumes(component string) map[string]v1.Volume {
	return c.volumes[component]
}

func (c *controlPlaneHostPathMounts) GetVolumeMounts(component string) map[string]v1.VolumeMount {
	return c.volumeMounts[component]
}

func (c *controlPlaneHostPathMounts) addComponentVolume(component string, vol v1.Volume) {
	if _, ok := c.volumes[component]; !ok {
		c.volumes[component] = map[string]v1.Volume{}
	}
	c.volumes[component][vol.Name] = vol
}

func (c *controlPlaneHostPathMounts) addComponentVolumeMount(component string, volMount v1.VolumeMount) {
	if _, ok := c.volumeMounts[component]; !ok {
		c.volumeMounts[component] = map[string]v1.VolumeMount{}
	}
	c.volumeMounts[component][volMount.Name] = volMount
}
