package component

import (
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
)

const (
	MeterAlertTemplate = `
# ----------------------------------------
# CORE PROPERTIES
# ----------------------------------------
debug=false
trace=false

# LOGGING
logging.level.com.yunion.meteralert=INFO

# OUTPUT
spring.output.ansi.enabled=ALWAYS

# ----------------------------------------
# WEB PROPERTIES
# ----------------------------------------

# EMBEDDED SERVER CONFIGURATION (ServerProperties)
server.port={{.Port}}
server.session-timeout=180
server.context-path=/
server.tomcat.basedir=.
server.tomcat.uri-encoding=UTF-8
server.tomcat.accesslog.enabled=true
server.tomcat.accesslog.dir=access_logs
server.tomcat.accesslog.file-date-format=.yyyy-MM-dd
server.tomcat.accesslog.prefix=access_log
server.tomcat.accesslog.suffix=.log
server.tomcat.accesslog.rotate=true

# ----------------------------------------
# DATA PROPERTIES
# ----------------------------------------

# DATASOURCE (DataSourceAutoConfiguration & DataSourceProperties)
spring.datasource.url=jdbc:mysql://{{.DBHost}}:{{.DBPort}}/{{.DB}}?useUnicode=true&characterEncoding=utf8&zeroDateTimeBehavior=convertToNull&useSSL=false&createDatabaseIfNotExist=true
spring.datasource.username={{.DBUser}}
spring.datasource.password={{.DBPassowrd}}
spring.datasource.driver-class-name=com.mysql.jdbc.Driver
spring.datasource.initialize=false
spring.datasource.continue-on-error=false
spring.datasource.sql-script-encoding=utf-8
spring.datasource.tomcat.default-auto-commit=true
spring.datasource.tomcat.initial-size=10
spring.datasource.tomcat.max-active=25
spring.datasource.tomcat.max-wait=30000
spring.datasource.tomcat.test-on-borrow=true
spring.datasource.tomcat.test-while-idle=true
spring.datasource.tomcat.validation-query=SELECT 1
spring.datasource.tomcat.validation-query-timeout=3
spring.datasource.tomcat.time-between-eviction-runs-millis=10000
spring.datasource.tomcat.min-evictable-idle-time-millis=120000
spring.datasource.tomcat.remove-abandoned=true
spring.datasource.tomcat.remove-abandoned-timeout=300
spring.liquibase.change-log=classpath:sql/master.xml



# ----------------------------------------
# Custom PROPERTIES
# ----------------------------------------

# OneCloud Authentication
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

# Scheduled Task
yunion.rc.async-job.initial-delay=2000
yunion.rc.async-job.fixed-rate=300000
yunion.rc.async-job.fixed-thread-pool=10

# InfluxDB backend
yunion.rc.influxdb.database=meter_db
yunion.rc.influxdb.retention-policy=autogen

# Alert
yunion.rc.alert.disable-link-url=http://office.yunion.io
yunion.rc.alert.statechangeonly-threshold=8h

# Prepaid Alert
yunion.rc.alert.prepaid.enabled=true
yunion.rc.alert.prepaid.cron=0 0 8 * * *
`
)

var MeterAlertComponent IComponent = NewMeterAlert()

type MeterAlert struct {
	*BaseComponent
}

func NewMeterAlert() *MeterAlert {
	m := new(MeterAlert)
	m.BaseComponent = NewBaseComponent(onecloud.ComponentType("meteralert"), m)
	return m
}

func (m MeterAlert) NewService(oc *onecloud.OnecloudCluster) *corev1.Service {
	return NewNodePortService(m.GetComponentType(), oc, MeterAlertPort)
}

func (c MeterAlert) NewDeployment(oc *onecloud.OnecloudCluster) (*apps.Deployment, error) {
	cf := func(volMounts []corev1.VolumeMount) []corev1.Container {
		volMounts = SetJavaConfigVolumeMounts(volMounts)
		return []corev1.Container{
			{
				Name: "meteralert",
				//Image: GetImage(oc, c.GetComponentType(), ""),
				Image: GetJavaAppImage(oc, ""),
				Env: []corev1.EnvVar{
					{
						Name:  JAVA_APP_JAR,
						Value: "meteralert.jar",
					},
				},
				VolumeMounts: volMounts,
			},
		}
	}
	cType := c.GetComponentType()
	deploy, err := NewDefaultDeployment(cType, oc, NewVolumeHelper(oc, cType), cf)
	if err != nil {
		return nil, err
	}
	podSpec := &deploy.Spec.Template.Spec
	podSpec.Volumes = SetJavaConfigVolumes(podSpec.Volumes)
	return deploy, nil
}

func (c MeterAlert) NewConfigMap(oc *onecloud.OnecloudCluster, cCfg *OnecloudComponentsConfig) (*corev1.ConfigMap, error) {
	cfg := cCfg.MeterAlertConfig
	config := NewJavaDBConfig(oc, cfg)
	return NewConfigMapByTemplate(c.GetComponentType(), oc, MeterAlertTemplate, config)
}

func (c MeterAlert) NewCloudUser(cfg *OnecloudComponentsConfig) *onecloud.CloudUser {
	return &cfg.MeterAlertConfig.CloudUser
}

func (c MeterAlert) NewDBConfig(cfg *OnecloudComponentsConfig) *onecloud.DBConfig {
	return &cfg.MeterAlertConfig.DB
}

func (c MeterAlert) NewCloudEndpoint() *CloudEndpoint {
	return NewHTTPCloudEndpoint(ServiceNameMeterAlert, ServiceTypeMeterAlert, MeterAlertPort, "api")
}
