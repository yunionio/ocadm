package component

import (
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
)

const (
	MeterCloudConfigTemplate = `
banner.charset=UTF-8
banner.location=classpath:config/banner.txt

spring.main.banner-mode=console

mybatis.type-aliases-package=com.yunion.apps
mybatis.mapper-locations=classpath:mappings/*/*.xml


debug=false

spring.output.ansi.enabled=DETECT

# LOGGING
logging.level.org.springframework=INFO
logging.level.com.yunion.apps=INFO
logging.level.com.yunionyun.mcp.mcclient=INFO
logging.level.org.apache.http.wire=INFO

# EMBEDDED SERVER CONFIGURATION (ServerProperties)
server.port=9094
server.session-timeout=180
server.context-path=/
server.tomcat.basedir=.
server.tomcat.uri-encoding=UTF-8
server.tomcat.accesslog.dir=access_logs
server.tomcat.accesslog.enabled=true
server.tomcat.accesslog.file-date-format=.yyyy-MM-dd
server.tomcat.accesslog.prefix=access_log
server.tomcat.accesslog.suffix=.log
server.tomcat.accesslog.rotate=true

# JACKSON (JacksonProperties)
spring.jackson.serialization.write_dates_as_timestamps=false

spring.datasource.url=jdbc:mysql://{{.DBHost}}:{{.DBPort}}/{{.DB}}?useUnicode=true&characterEncoding=utf8&zeroDateTimeBehavior=convertToNull&useSSL=false&createDatabaseIfNotExist=true
spring.datasource.username={{.DBUser}}
spring.datasource.password={{.DBPassowrd}}

spring.datasource.driver-class-name=com.mysql.jdbc.Driver

# Keystone Authentication & Authorization
yunionyun.auth.url={{.AuthURL}}
yunionyun.auth.domain={{.AuthDomain}}
yunionyun.auth.username={{.AuthUsername}}
yunionyun.auth.password={{.AuthPassword}}
yunionyun.auth.project={{.AuthProject}}
yunionyun.auth.region={{.Region}}
yunionyun.auth.session-region={{.Region}}
yunionyun.auth.cache-size=500
yunionyun.auth.timeout=1000
yunionyun.auth.debug=true
yunionyun.auth.insecure=true
yunionyun.auth.refresh-interval=900000

download.awsfilepath=/opt/yunion-meter/awsdatafile/
`
)

var MeterCloudComponent IComponent = NewMeterCloud()

type MeterCloud struct {
	*BaseComponent
}

func NewMeterCloud() *MeterCloud {
	m := new(MeterCloud)
	m.BaseComponent = NewBaseComponent(onecloud.ComponentType("meter-cloud"), m)
	return m
}

func (m MeterCloud) NewService(oc *onecloud.OnecloudCluster) *corev1.Service {
	return NewNodePortService(m.GetComponentType(), oc, MeterCloudPort)
}

func (m MeterCloud) NewDeployment(oc *onecloud.OnecloudCluster) (*apps.Deployment, error) {
	cf := func(volMounts []corev1.VolumeMount) []corev1.Container {
		volMounts = SetJavaConfigVolumeMounts(volMounts)
		return []corev1.Container{
			{
				Name: "meter-cloud",
				//Image: GetImage(oc, m.GetComponentType(), ""),
				Image: GetJavaAppImage(oc, ""),
				Env: []corev1.EnvVar{
					{
						Name:  JAVA_APP_JAR,
						Value: "meter-cloud.jar",
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

func (m MeterCloud) NewConfigMap(oc *onecloud.OnecloudCluster, cCfg *OnecloudComponentsConfig) (*corev1.ConfigMap, error) {
	cfg := cCfg.MeterConfig
	config := NewJavaDBConfig(oc, cfg)
	return NewConfigMapByTemplate(m.GetComponentType(), oc, MeterCloudConfigTemplate, config)
}

func (m MeterCloud) NewCloudUser(cfg *OnecloudComponentsConfig) *onecloud.CloudUser {
	return &cfg.MeterConfig.CloudUser
}

func (m MeterCloud) NewDBConfig(cfg *OnecloudComponentsConfig) *onecloud.DBConfig {
	return &cfg.MeterConfig.DB
}

func (m MeterCloud) NewCloudEndpoint() *CloudEndpoint {
	return NewHTTPCloudEndpoint(ServiceNameMeterCloud, ServiceTypeMeterCloud, MeterCloudPort, "api")
}
