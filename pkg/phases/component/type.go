package component

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"yunion.io/x/ocadm/pkg/apis/constants"
	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
	"yunion.io/x/onecloud-operator/pkg/controller"
	"yunion.io/x/onecloud-operator/pkg/util/passwd"
)

const (
	OnecloudComponentsConfigKey = "OnecloudComponentsConfig"

	MeterAdminUser   = "meter"
	MeterPort        = 9090
	MeterDB          = "meter"
	MeterDBUser      = "meter"
	ServiceNameMeter = "meter"
	ServiceTypeMeter = "meter"

	MeterAlertAdminUser   = "meteralert"
	MeterAlertPort        = 9494
	MeterAlertDB          = "meteralert"
	MeterAlertDBUser      = "meteralert"
	ServiceNameMeterAlert = "meteralert"
	ServiceTypeMeterAlert = "meteralert"

	CloudmonAdminUser = "cloudmon"

	MeterServicePort      = 9091
	MeterTrafficPort      = 9093
	MeterCloudPort        = 9094
	ServiceNameMeterCloud = "meter-cloud"
	ServiceTypeMeterCloud = "meter-cloud"

	CloudWatcherAdminUser   = "cloudwatcher"
	CloudWatcherPort        = 8787
	CloudWatcherDB          = "cloudwatcher"
	CloudWatcherDBUser      = "cloudwatcher"
	ServiceNameCloudWatcher = "cloudwatcher"
	ServiceTypeCloudWatcher = "cloudwatcher"

	ItsmAdminUser   = "itsm"
	ItsmPort        = 9595
	ItsmDB          = "itsm"
	ItsmDBUser      = "itsm"
	ServiceNameItsm = "itsm"
	ServiceTypeItsm = "itsm"
)

type ItsmConfigOptions struct {
	onecloud.ServiceDBCommonOptions
	SecondDatabase string `json:"secondDatabase"`
	EncryptionKey  string `json:"encryptionKey"`
}

type OnecloudComponentsConfig struct {
	MeterConfig        onecloud.ServiceDBCommonOptions `json:"meter"`
	MeterAlertConfig   onecloud.ServiceDBCommonOptions `json:"meteralert"`
	CloudmonConfig     onecloud.ServiceCommonOptions   `json:"cloudmon"`
	CloudWatcherConfig onecloud.ServiceDBCommonOptions `json:"cloudwatcher"`
	ItsmConfig         ItsmConfigOptions               `json:"itsm"`
}

func NewOnecloudComponentsConfig(old *OnecloudComponentsConfig) (*OnecloudComponentsConfig, error) {
	newObj := new(OnecloudComponentsConfig)
	if old != nil {
		bs, err := yaml.Marshal(old)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(bs, newObj); err != nil {
			return nil, err
		}
	}
	return FillOnecloudComponentsConfigDefault(newObj), nil
}

func FillOnecloudComponentsConfigDefault(obj *OnecloudComponentsConfig) *OnecloudComponentsConfig {
	type userDBPort struct {
		user   string
		port   int
		db     string
		dbUser string
	}

	for opt, tmp := range map[*onecloud.ServiceDBCommonOptions]userDBPort{
		&obj.MeterConfig:                       {MeterAdminUser, MeterPort, MeterDB, MeterDBUser},
		&obj.MeterAlertConfig:                  {MeterAlertAdminUser, MeterAlertPort, MeterAlertDB, MeterAlertDBUser},
		&obj.CloudWatcherConfig:                {CloudWatcherAdminUser, CloudWatcherPort, CloudWatcherDB, CloudWatcherDBUser},
		&obj.ItsmConfig.ServiceDBCommonOptions: {ItsmAdminUser, ItsmPort, ItsmDB, ItsmDBUser},
	} {
		onecloud.SetDefaults_ServiceDBCommonOptions(opt, tmp.db, tmp.dbUser, tmp.user, tmp.port)
	}

	itsmConfig := &obj.ItsmConfig
	if itsmConfig.SecondDatabase == "" {
		itsmConfig.SecondDatabase = fmt.Sprintf("%s_engine", itsmConfig.DB.Database)
	}
	if itsmConfig.EncryptionKey == "" {
		itsmConfig.EncryptionKey = passwd.GeneratePassword()
	}

	type userPort struct {
		user string
		port int
	}
	for opt, up := range map[*onecloud.ServiceCommonOptions]userPort{
		&obj.CloudmonConfig: {CloudmonAdminUser, 0},
	} {
		onecloud.SetDefaults_ServiceCommonOptions(opt, up.user, up.port)
	}

	return obj
}

func ComponentsConfigMapName(oc *onecloud.OnecloudCluster) string {
	return fmt.Sprintf("%s-%s", oc.GetName(), "cluster-components-config")
}

func NewOnecloudComponentsConfigFromYaml(data string) (*OnecloudComponentsConfig, error) {
	cfg := &OnecloudComponentsConfig{}
	if err := yaml.Unmarshal([]byte(data), cfg); err != nil {
		return nil, errors.Wrap(err, "failed to decode onecloud cluster config data")
	}
	return cfg, nil
}

func NewOnecloudComponentsConfigFromConfigMap(cfgMap *corev1.ConfigMap) (*OnecloudComponentsConfig, error) {
	data, ok := cfgMap.Data[OnecloudComponentsConfigKey]
	if !ok {
		return nil, errors.Errorf("unexpected error when reading %s ConfigMap: %s key value pair missing", cfgMap.GetName(), OnecloudComponentsConfigKey)
	}
	cfg, err := NewOnecloudComponentsConfigFromYaml(data)
	if err != nil {
		return nil, err
	}
	return NewOnecloudComponentsConfig(cfg)
}

func (obj *OnecloudComponentsConfig) ToYaml() (string, error) {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *OnecloudComponentsConfig) ToConfigMap(oc *onecloud.OnecloudCluster) (*corev1.ConfigMap, error) {
	data, err := obj.ToYaml()
	if err != nil {
		return nil, err
	}
	cfgMapName := ComponentsConfigMapName(oc)
	return &corev1.ConfigMap{
		ObjectMeta: GetObjectMeta(oc, cfgMapName, nil),
		Data: map[string]string{
			OnecloudComponentsConfigKey: data,
		},
	}, nil
}

type JavaBaseConfig struct {
	Port         int
	AuthURL      string
	AuthDomain   string
	AuthUsername string
	AuthPassword string
	AuthProject  string
	Region       string
}

func NewJavaBaseConfig(oc *onecloud.OnecloudCluster, port int, user, passwd string) *JavaBaseConfig {
	return &JavaBaseConfig{
		AuthURL:      controller.GetAuthURL(oc),
		AuthProject:  constants.SysAdminProject,
		AuthDomain:   constants.DefaultDomain,
		AuthUsername: user,
		AuthPassword: passwd,
		Region:       oc.Spec.Region,
		Port:         port,
	}
}

type JavaDBConfig struct {
	JavaBaseConfig
	DBHost     string
	DBPort     int32
	DB         string
	DBUser     string
	DBPassowrd string
}

func NewJavaDBConfig(oc *onecloud.OnecloudCluster, cfg onecloud.ServiceDBCommonOptions) *JavaDBConfig {
	opt := NewJavaBaseConfig(oc, cfg.Port, cfg.CloudUser.Username, cfg.CloudUser.Password)
	dbCfg := &JavaDBConfig{
		JavaBaseConfig: *opt,
		DBHost:         oc.Spec.Mysql.Host,
		DBPort:         oc.Spec.Mysql.Port,
		DB:             cfg.DB.Database,
		DBUser:         cfg.DB.Username,
		DBPassowrd:     cfg.DB.Password,
	}
	return dbCfg
}

type CloudEndpoint struct {
	Proto       string
	Region      string
	ServiceName string
	ServiceType string
	Port        int
	Prefix      string
}

func NewProtoCloudEndpoint(proto, svcName, svcType string, port int, prefix string) *CloudEndpoint {
	return &CloudEndpoint{
		Proto:       proto,
		ServiceName: svcName,
		ServiceType: svcType,
		Port:        port,
		Prefix:      prefix,
	}
}

func NewHTTPCloudEndpoint(svcName, svcType string, port int, prefix string) *CloudEndpoint {
	return NewProtoCloudEndpoint("http", svcName, svcType, port, prefix)
}

func NewHTTPSCloudEndpoint(svcName, svcType string, port int, prefix string) *CloudEndpoint {
	return NewProtoCloudEndpoint("https", svcName, svcType, port, prefix)
}

func (ep CloudEndpoint) GetUrl(address string) string {
	url := fmt.Sprintf("%s://%s:%d", ep.Proto, address, ep.Port)
	if ep.Prefix != "" {
		url = fmt.Sprintf("%s/%s", url, ep.Prefix)
	}
	return url
}
