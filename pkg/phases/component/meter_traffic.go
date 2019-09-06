package component

import (
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
)

const (
	MeterTrafficConfigTemplate = `
# ----------------------------------------
# CORE PROPERTIES
# ----------------------------------------

# BANNER
banner.charset=UTF-8
banner.location=classpath:config/banner.txt

# APPLICATION SETTINGS (SpringApplication)
# Mode used to display the banner when the application runs.
#   console printed on System.out(console)
#   log     using the configured logger
#   off     not at all
spring.main.banner-mode=console

mybatis.type-aliases-package=com.yunion.apps
mybatis.mapper-locations=classpath:mappings/*/*.xml

# ----------------------------------------
# CORE PROPERTIES
# ----------------------------------------

# DEBUG
debug=false

# OUTPUT
#   NEVER\uff1adisable ANSI-colored output (default)
#   DETECT\uff1awill check whether the terminal supports ANSI, yes, then use the color output (recommend)
#   ALWAYS\uff1aalways use ANSI-colored format output, if the terminal does not support, there will be a lot of interference information (not recommended)
spring.output.ansi.enabled=DETECT

# LOGGING
logging.level.org.springframework=INFO
logging.level.com.yunion.apps=DEBUG
logging.level.com.yunionyun.mcp.mcclient=INFO

# EMBEDDED SERVER CONFIGURATION (ServerProperties)
server.port={{.Port}}
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

# DATASOURCE (DataSourceAutoConfiguration & DataSourceProperties)
spring.datasource.url=jdbc:mysql://{{.DBHost}}:{{.DBPort}}/{{.DB}}?useUnicode=true&characterEncoding=utf8&zeroDateTimeBehavior=convertToNull&useSSL=false&createDatabaseIfNotExist=true
spring.datasource.username={{.DBUser}}
spring.datasource.password={{.DBPassowrd}}
spring.datasource.driver-class-name=com.mysql.jdbc.Driver
liquibase.change-log=classpath:sql/master.xml

# ----------------------------------------
# INTEGRATION PROPERTIES
# ----------------------------------------

# Keystone Authentication & Authorization
yunionyun.auth.url={{.AuthURL}}
yunionyun.auth.domain={{.AuthDomain}}
yunionyun.auth.username={{.AuthUsername}}
yunionyun.auth.password={{.AuthPassword}}
yunionyun.auth.project={{.AuthProject}}
yunionyun.auth.cache-size=500
yunionyun.auth.timeout=1000
yunionyun.auth.debug=true
yunionyun.auth.insecure=true
yunionyun.auth.refresh-interval=900000
yunionyun.auth.session-region={{.Region}}
##synchronize event-log region with &&
yunionyun.auth.eventlog-region={{.Region}}

default.platform=KVM
yunionyun.service.type=influxdb
yunionyun.influxdb.name=telegraf
yunionyun.hourly.hourInterval=1h
yunionyun.hourly.minuteInterval=5m

# ### schedule timing setting
# 1-demo>>  17:10 every day
#jobs.cron_demoSchedule=0 10 17 * * ?
# 2-statDataProcess>>  17:10 every day
jobs.cron_statDataProcess=0 30 0 * * ?
# 3-synchronizeData config
jobs.synchronizeData.pagesize=2000
## 1000 * 60
jobs.synchronizeData.initialDelay=60000
## 1000 * 60 * 4
jobs.synchronizeData.fixedDelay=240000
# 4-etlProcessData config
## 1000 * 60 * 2
jobs.etlProcessData.initialDelay=120000
## 1000 * 60 * 4
jobs.etlProcessData.fixedDelay=240000

# >> hourly every day=(xx:01)
jobs.synchronize.netio.data=0 18 * * * ?

# >> 01:01 every day=(xx:01)
jobs.process.netio.result=0 1 1 * * ?

# 1 minutes after tomcat is running
jobs.usesumlog.initialDelay=60000
# 10 minutes interval
jobs.usesumlog.fixedDelay=3600000
# 80 minutes for each time
yunionyun.usesum.hourInterval=80m
# 6 hours the first time
yunionyun.usesum.firsttimeInterval=2h

# >> 01:01 every day=(xx:01)
jobs.fetch.aliyun.priceinfo=0 30 0 * * ?
`
)

var MeterTrafficComponent IComponent = NewMeterTraffic()

type MeterTraffic struct {
	*BaseComponent
}

func NewMeterTraffic() *MeterTraffic {
	m := new(MeterTraffic)
	m.BaseComponent = NewBaseComponent(onecloud.ComponentType("meter-traffic"), m)
	return m
}

func (m MeterTraffic) NewService(oc *onecloud.OnecloudCluster) *corev1.Service {
	return NewNodePortService(m.GetComponentType(), oc, MeterTrafficPort)
}

func (m MeterTraffic) NewDeployment(oc *onecloud.OnecloudCluster) (*apps.Deployment, error) {
	cf := func(volMounts []corev1.VolumeMount) []corev1.Container {
		volMounts = SetJavaConfigVolumeMounts(volMounts)
		return []corev1.Container{
			{
				Name: "meter-traffic",
				//Image: GetImage(oc, m.GetComponentType(), ""),
				Image: GetJavaAppImage(oc, ""),
				Env: []corev1.EnvVar{
					{
						Name:  JAVA_APP_JAR,
						Value: "meter-traffic.jar",
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

func (m MeterTraffic) NewConfigMap(oc *onecloud.OnecloudCluster, cCfg *OnecloudComponentsConfig) (*corev1.ConfigMap, error) {
	cfg := cCfg.MeterConfig
	config := NewJavaDBConfig(oc, cfg)
	return NewConfigMapByTemplate(m.GetComponentType(), oc, MeterTrafficConfigTemplate, config)
}

func (m MeterTraffic) NewCloudUser(cfg *OnecloudComponentsConfig) *onecloud.CloudUser {
	return &cfg.MeterConfig.CloudUser
}

func (m MeterTraffic) NewDBConfig(cfg *OnecloudComponentsConfig) *onecloud.DBConfig {
	return &cfg.MeterConfig.DB
}
