package occonfig

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"golang.org/x/sync/errgroup"
	"yunion.io/x/ocadm/pkg/util/onecloud"
	"yunion.io/x/onecloud/pkg/mcclient"

	"github.com/pkg/errors"

	"yunion.io/x/jsonutils"
	baremetaloptions "yunion.io/x/onecloud/pkg/baremetal/options"
	regionoptions "yunion.io/x/onecloud/pkg/compute/options"
	imageoptions "yunion.io/x/onecloud/pkg/image/options"
	keystoneoptions "yunion.io/x/onecloud/pkg/keystone/options"

	"yunion.io/x/ocadm/pkg/apis/constants"
)

const (
	RCAdminConfigTemplate = `export OS_AUTH_URL={{.AuthUrl}}
export OS_PROJECT_NAME={{.ProjectName}}
export OS_USERNAME={{.Username}}
export OS_PASSWORD={{.Password}}
export OS_DOMAIN_NAME={{.DomainName}}
export YUNION_INSECURE={{.Insecure}}
export OS_REGION_NAME={{.Region}}
export YUNION_CERT_FILE={{.CertFile}}
export YUNION_KEY_FILE={{.KeyFile}}
`
)

type RCAdminConfig struct {
	AuthUrl       string
	Region        string
	Username      string
	Password      string
	DomainName    string
	ProjectName   string
	ProjectDomain string
	Insecure      bool
	Debug         bool
	Timeout       int
	CertFile      string
	KeyFile       string
}

func NewRCAdminConfig(authURL string, region string, passwd string, certFile string, keyFile string) *RCAdminConfig {
	return &RCAdminConfig{
		AuthUrl:       authURL,
		Region:        region,
		Username:      constants.SysAdminUsername,
		Password:      passwd,
		DomainName:    constants.DefaultDomain,
		ProjectName:   constants.SysAdminProject,
		Insecure:      true,
		Debug:         false,
		Timeout:       600,
		ProjectDomain: "", // TODO: should we use this?
		CertFile:      certFile,
		KeyFile:       keyFile,
	}
}

func generateTemplate(kind string, tlp string, data interface{}) (string, error) {
	t, err := template.New(kind).Parse(tlp)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %s", kind)
	}
	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return "", errors.Wrapf(err, "failed to generate %s template", kind)
	}
	return out.String(), nil
}

func (c RCAdminConfig) RCAdminContent() (string, error) {
	return generateTemplate(constants.OnecloudAdminConfigFileName, RCAdminConfigTemplate, c)
}

func WriteRCAdminConfigFile(opt *RCAdminConfig) error {
	configFile := path.Join(constants.OnecloudConfigDir, constants.OnecloudAdminConfigFileName)
	content, err := opt.RCAdminContent()
	if err != nil {
		return err
	}
	if err := writeOnecloudFile(configFile, content); err != nil {
		return err
	}
	if err := WriteOnecloudConfigFile(constants.OnecloudConfigDir, constants.OnecloudAdminConfigFileName, opt); err != nil {
		return err
	}
	return nil
}

func WriteKeystoneConfigFile(opt keystoneoptions.SKeystoneOptions) error {
	return WriteOnecloudConfigFile(
		constants.OnecloudKeystoneConfigDir,
		constants.OnecloudKeystoneConfigFileName,
		opt,
	)
}

type RegionSchedulerOptions struct {
	regionoptions.ComputeOptions
	SchedulerPort int
}

func WriteRegionConfigFile(opt RegionSchedulerOptions) error {
	return WriteOnecloudConfigFile(
		constants.OnecloudConfigDir,
		constants.OnecloudRegionConfigFileName,
		opt,
	)
}

func WriteGlanceConfigFile(opt imageoptions.SImageOptions) error {
	return WriteOnecloudConfigFile(
		constants.OnecloudGlanceConfigDir,
		constants.OnecloudGlanceConfigFileName,
		opt,
	)
}

func WriteBaremetalConfigFile(opt baremetaloptions.BaremetalOptions) error {
	return WriteOnecloudConfigFile(
		constants.OnecloudConfigDir,
		constants.OnecloudBaremetalConfigFileName,
		opt,
	)
}

func AdminConfigFilePath() string {
	return YAMLConfigFilePath(
		constants.OnecloudConfigDir,
		constants.OnecloudAdminConfigFileName,
	)
}

func KeystoneConfigFilePath() string {
	return YAMLConfigFilePath(
		constants.OnecloudKeystoneConfigDir,
		constants.OnecloudKeystoneConfigFileName,
	)
}

func RegionConfigFilePath() string {
	return YAMLConfigFilePath(
		constants.OnecloudConfigDir,
		constants.OnecloudRegionConfigFileName,
	)
}

func GlanceConfigFilePath() string {
	return YAMLConfigFilePath(
		constants.OnecloudGlanceConfigDir,
		constants.OnecloudGlanceConfigFileName,
	)
}

func BaremetalConfigFilePath() string {
	return YAMLConfigFilePath(
		constants.OnecloudConfigDir,
		constants.OnecloudBaremetalConfigFileName,
	)
}

func YAMLConfigFilePath(dir string, fileName string) string {
	return path.Join(dir, fmt.Sprintf("%s%s", fileName, constants.OnecloudConfigFileSuffix))
}

func WriteOnecloudConfigFile(dir string, fileName string, optStruct interface{}) error {
	configFile := YAMLConfigFilePath(dir, fileName)
	content := jsonutils.Marshal(optStruct).YAMLString()
	return writeOnecloudFile(configFile, content)
}

func writeOnecloudFile(filePath string, content string) error {
	parentDir := path.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return errors.Wrapf(err, "create dir %s", parentDir)
	}
	if err := ioutil.WriteFile(filePath, []byte(content), 0755); err != nil {
		return errors.Wrapf(err, "write file %s", filePath)
	}
	return nil
}

func NewRCAdminConfigByFile(configFile string) (*RCAdminConfig, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	return NewRCAdminConfigByBytes(data)
}

func NewRCAdminConfigByBytes(data []byte) (*RCAdminConfig, error) {
	obj, err := jsonutils.ParseYAML(string(data))
	if err != nil {
		return nil, err
	}
	config := new(RCAdminConfig)
	if err := obj.Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}

func ClientSessionFromFile(configFile string) (*mcclient.ClientSession, error) {
	config, err := NewRCAdminConfigByFile(configFile)
	if err != nil {
		return nil, err
	}
	cli := mcclient.NewClient(
		config.AuthUrl,
		config.Timeout,
		config.Debug,
		config.Insecure,
		config.CertFile,
		config.KeyFile,
	)
	token, err := cli.AuthenticateWithSource(
		config.Username,
		config.Password,
		config.DomainName,
		config.ProjectName,
		config.ProjectDomain,
		mcclient.AuthSourceCli,
	)
	if err != nil {
		return nil, err
	}
	session := cli.NewSession(
		context.Background(),
		config.Region,
		"",
		constants.EndpointTypeInternal,
		token,
		"",
	)
	return session, nil
}

func InitServiceAccount(s *mcclient.ClientSession, username string, password string) error {
	obj, exists, err := onecloud.IsUserExists(s, username)
	if err != nil {
		return err
	}
	if exists {
		id, _ := obj.GetString("id")
		if _, err := onecloud.ChangeUserPassword(s, id, password); err != nil {
			return errors.Wrapf(err, "user %s already exists, update password", username)
		}
		return nil
	}
	obj, err = onecloud.CreateUser(s, username, password)
	if err != nil {
		return errors.Wrapf(err, "create user %s", username)
	}
	userId, _ := obj.GetString("id")
	return onecloud.ProjectAddUser(s, constants.SysAdminProject, userId, constants.RoleAdmin)
}

func RegisterServiceEndpointByInterfaces(
	s *mcclient.ClientSession,
	regionId string,
	serviceName string,
	serviceType string,
	endpointUrl string,
	interfaces []string,
) error {
	svc, err := onecloud.EnsureService(s, serviceName, serviceType)
	if err != nil {
		return err
	}
	svcId, err := svc.GetString("id")
	if err != nil {
		return err
	}
	errgrp := &errgroup.Group{}
	for _, inf := range interfaces {
		tmpInf := inf
		errgrp.Go(func() error {
			_, err = onecloud.EnsureEndpoint(s, svcId, regionId, tmpInf, endpointUrl)
			if err != nil {
				return err
			}
			return nil
		})
	}
	return errgrp.Wait()
}

func RegisterServicePublicInternalEndpoint(
	s *mcclient.ClientSession,
	regionId string,
	serviceName string,
	serviceType string,
	endpointUrl string,
) error {
	return RegisterServiceEndpointByInterfaces(s, regionId, serviceName, serviceType,
		endpointUrl, []string{constants.EndpointTypeInternal, constants.EndpointTypePublic})
}

func GetConfigFileObject(filepath string) (jsonutils.JSONObject, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return jsonutils.ParseYAML(string(content))
}
