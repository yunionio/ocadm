package component

import (
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
)

const (
	ItsmTemplate = `
debug=false
trace=false

# BANNER
banner.charset=UTF-8

# LOGGING
logging.level.com.yunion=INFO

# OUTPUT
spring.output.ansi.enabled=ALWAYS
spring.security.basic.enabled=false

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
datasource.primary.jdbc-url=jdbc:mysql://{{.DBHost}}:{{.DBPort}}/{{.DB}}?useUnicode=true&characterEncoding=utf8&zeroDateTimeBehavior=convertToNull&useSSL=false&createDatabaseIfNotExist=true
datasource.primary.username={{.DBUser}}
datasource.primary.password={{.DBPassowrd}}
datasource.primary.driver-class-name=com.mysql.cj.jdbc.Driver
datasource.primary.initialize=false
datasource.primary.continue-on-error=false
datasource.primary.sql-script-encoding=utf-8
datasource.primary.schema=classpath:sql/schema.sql
datasource.primary.initialization-mode=always

datasource.secondary.jdbc-url=jdbc:mysql://{{.DBHost}}:{{.DBPort}}/{{.DB2nd}}?useUnicode=true&characterEncoding=utf8&zeroDateTimeBehavior=convertToNull&useSSL=false&createDatabaseIfNotExist=true&useTimezone=true&serverTimezone=UTC
datasource.secondary.username={{.DBUser}}
datasource.secondary.password={{.DBPassowrd}}
datasource.secondary.driver-class-name=com.mysql.cj.jdbc.Driver
datasource.secondary.initialize=false
datasource.secondary.continue-on-error=false
datasource.secondary.sql-script-encoding=utf-8

# ----------------------------------------
# PROCESS SERVICE PROPERTIES
# ----------------------------------------

# Workflow
camunda.bpm.history-level=FULL
camunda.bpm.admin-user.id=demo
camunda.bpm.admin-user.password=demo
camunda.bpm.admin-user.firstName=Yunion
camunda.bpm.admin-user.lastName=ITSM
camunda.bpm.admin-user.email=ningyu@yunion.cn
camunda.bpm.filter.create=All tasks
camunda.bpm.database.type=mysql


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

# email notify.
yunion.rc.email.link.parameter.name=mail-call-back-address
yunion.rc.email.link.parameter.key=address
yunion.rc.email.link.encryption.key={{.EncryptionKey}}

#
non_default_domain_projects=false`
)

var ItsmComponent IComponent = NewItsm()

type Itsm struct {
	*BaseComponent
}

func NewItsm() *Itsm {
	m := new(Itsm)
	m.BaseComponent = NewBaseComponent(onecloud.ComponentType("itsm"), m)
	return m
}

func (m Itsm) NewDeployment(oc *onecloud.OnecloudCluster) (*apps.Deployment, error) {
	cf := func(volMounts []corev1.VolumeMount) []corev1.Container {
		volMounts = SetJavaConfigVolumeMounts(volMounts)
		return []corev1.Container{
			{
				Name: "itsm",
				//Image: GetImage(oc, m.GetComponentType(), ""),
				Image: GetJavaAppImage(oc, ""),
				Env: []corev1.EnvVar{
					{
						Name:  JAVA_APP_JAR,
						Value: "itsm.jar",
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

type ItsmConfigOption struct {
	JavaDBConfig
	DB2nd         string
	EncryptionKey string
}

func (m Itsm) NewService(oc *onecloud.OnecloudCluster) *corev1.Service {
	return NewNodePortService(m.GetComponentType(), oc, ItsmPort)
}

func (m Itsm) NewConfigMap(oc *onecloud.OnecloudCluster, cCfg *OnecloudComponentsConfig) (*corev1.ConfigMap, error) {
	cfg := cCfg.ItsmConfig
	config := ItsmConfigOption{
		JavaDBConfig:  *NewJavaDBConfig(oc, cfg.ServiceDBCommonOptions),
		DB2nd:         cfg.SecondDatabase,
		EncryptionKey: cfg.EncryptionKey,
	}
	return NewConfigMapByTemplate(m.GetComponentType(), oc, ItsmTemplate, config)
}

func (m Itsm) NewCloudUser(cfg *OnecloudComponentsConfig) *onecloud.CloudUser {
	return &cfg.ItsmConfig.CloudUser
}

func (m Itsm) NewDBConfig(cfg *OnecloudComponentsConfig) *onecloud.DBConfig {
	return &cfg.ItsmConfig.DB
}

func (m Itsm) NewDBConfig2(cfg *OnecloudComponentsConfig) *onecloud.DBConfig {
	tmp := cfg.ItsmConfig.DB
	tmp.Database = cfg.ItsmConfig.SecondDatabase
	return &tmp
}

func (m Itsm) NewCloudEndpoint() *CloudEndpoint {
	return NewHTTPCloudEndpoint(ServiceNameItsm, ServiceTypeItsm, ItsmPort, "")
}
