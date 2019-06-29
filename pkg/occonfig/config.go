package occonfig

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"
	"yunion.io/x/onecloud/pkg/mcclient"

	"github.com/pkg/errors"

	"yunion.io/x/jsonutils"
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
	if err := writeOnecloudConfigFile(constants.OnecloudConfigDir, constants.OnecloudAdminConfigFileName, opt); err != nil {
		return err
	}
	return nil
}

func WriteKeystoneConfigFile(opt keystoneoptions.SKeystoneOptions) error {
	return writeOnecloudConfigFile(
		constants.OnecloudKeystoneConfigDir,
		constants.OnecloudKeystoneConfigFileName,
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

func YAMLConfigFilePath(dir string, fileName string) string {
	return path.Join(dir, fmt.Sprintf("%s%s", fileName, constants.OnecloudConfigFileSuffix))
}

func writeOnecloudConfigFile(dir string, fileName string, optStruct interface{}) error {
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

func ClientSessionFromFile(configFile string) (*mcclient.ClientSession, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	obj, err := jsonutils.ParseYAML(string(data))
	if err != nil {
		return nil, err
	}
	config := new(RCAdminConfig)
	if err := obj.Unmarshal(config); err != nil {
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
