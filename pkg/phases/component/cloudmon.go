package component

import (
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
)

const (
	CloudMonConfigTemplate = `
debug=false
trace=false
logging.level.com.yunion.cloudmon=INFO

yunion.rc.auth.url={{.AuthURL}}
yunion.rc.auth.domain={{.AuthDomain}}
yunion.rc.auth.username={{.AuthUsername}}
yunion.rc.auth.password={{.AuthPassword}}
yunion.rc.auth.project={{.AuthProject}}
yunion.rc.auth.region={{.Region}}
yunion.rc.auth.cache-size=500
yunion.rc.auth.timeout=1000
yunion.rc.auth.debug=true
yunion.rc.auth.insecure=true
yunion.rc.auth.refresh-interval=300000

yunion.rc.async-job.initial-delay=2000
yunion.rc.async-job.fixed-rate=300000
yunion.rc.async-job.fixed-thread-pool=10

yunion.rc.influxdb.database=telegraf
yunion.rc.influxdb.policy=30day_only
yunion.rc.influxdb.measurement=instance

yunion.rc.metrics.ins.providers=Aliyun,Azure,Aws,Qcloud,VMWare,Huawei,Openstack,Ucloud,ZStack
yunion.rc.metrics.eip.providers=Aliyun,Qcloud
`
)

var CloudMonComponent IComponent = NewCloudMon()

type CloudMon struct {
	*BaseComponent
}

func NewCloudMon() *CloudMon {
	m := new(CloudMon)
	m.BaseComponent = NewBaseComponent(onecloud.ComponentType("cloudmon"), m)
	return m
}

func (m CloudMon) NewDeployment(oc *onecloud.OnecloudCluster) (*apps.Deployment, error) {
	cf := func(volMounts []corev1.VolumeMount) []corev1.Container {
		volMounts = SetJavaConfigVolumeMounts(volMounts)
		return []corev1.Container{
			{
				Name:  "cloudmon",
				Image: GetImage(oc, m.GetComponentType(), ""),
				Env: []corev1.EnvVar{
					{
						Name:  JAVA_APP_JAR,
						Value: "cloudmon.jar",
					},
				},
				VolumeMounts: volMounts,
			},
		}
	}
	cType := m.GetComponentType()
	deploy, err := NewDefaultDeployment(cType, oc, NewVolumeHelper(oc, cType), cf)
	if err != nil {
		return nil, err
	}
	podSpec := &deploy.Spec.Template.Spec
	podSpec.Volumes = SetJavaConfigVolumes(podSpec.Volumes)
	return deploy, nil
}

func (m CloudMon) NewConfigMap(oc *onecloud.OnecloudCluster, cCfg *OnecloudComponentsConfig) (*corev1.ConfigMap, error) {
	cfg := cCfg.CloudmonConfig
	config := NewJavaBaseConfig(oc, cfg.Port, cfg.Username, cfg.Password)
	return NewConfigMapByTemplate(m.GetComponentType(), oc, CloudMonConfigTemplate, config)
}

func (m CloudMon) NewCloudUser(cfg *OnecloudComponentsConfig) *onecloud.CloudUser {
	return &cfg.CloudmonConfig.CloudUser
}
