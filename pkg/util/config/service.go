package config

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/onecloud/pkg/cloudcommon/options"
	"yunion.io/x/pkg/util/reflectutils"
	"yunion.io/x/structarg"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/util/mysql"
)

func InitDBUser(conn *mysql.Connection, info apis.DBInfo) error {
	dbName := info.Database
	dbExists, err := conn.IsDatabaseExists(dbName)
	if err != nil {
		return errors.Wrap(err, "check db exists")
	}
	if dbExists {
		return errors.Errorf("database %q already exists", dbName)
	}
	if err := conn.CreateDatabase(dbName); err != nil {
		return errors.Wrapf(err, "create database %q", dbName)
	}
	user := info.Username
	password := info.Password
	if err := conn.CreateUser(user, password, dbName); err != nil {
		return errors.Wrapf(err, "create user %q for database %q", user, dbName)
	}
	return nil
}

func SetOptionsDefault(opt interface{}, serviceType string) error {
	parser, err := structarg.NewArgumentParser(opt, constants.ServiceTypeCompute, "", "")
	if err != nil {
		return err
	}
	parser.SetDefault()

	var optionsRef *options.BaseOptions
	if err := reflectutils.FindAnonymouStructPointer(opt, &optionsRef); err != nil {
		return err
	}
	if len(optionsRef.ApplicationID) == 0 {
		optionsRef.ApplicationID = serviceType
	}
	return nil
}

func EnableConfigTLS(config *apis.ServiceBaseOptions, certDir string, ca string, cert string, key string) {
	config.EnableSSL = true
	config.SSLCAFile = path.Join(certDir, ca)
	config.SSLCertFile = path.Join(certDir, cert)
	config.SSLKeyFile = path.Join(certDir, key)
}

func SetServiceBaseOptions(opt *options.BaseOptions, input apis.ServiceBaseOptions) {
	opt.Region = input.Region
	opt.Port = input.Port
	opt.EnableSsl = input.EnableSSL
	opt.EnableRbac = input.EnableRBAC
	opt.SslCaCerts = input.SSLCAFile
	opt.SslCertfile = input.SSLCertFile
	opt.SslKeyfile = input.SSLKeyFile
}

func SetServiceCommonOptions(opt *options.CommonOptions, input apis.ServiceCommonOptions) {
	SetServiceBaseOptions(&opt.BaseOptions, input.ServiceBaseOptions)
	opt.AuthURL = input.AuthURL
	opt.AdminUser = input.AdminUser
	opt.AdminDomain = input.AdminDomain
	opt.AdminPassword = input.AdminPassword
	opt.AdminProject = input.AdminProject
	opt.AdminProjectDomain = input.AdminProjectDomain
}

func SetDBOptions(opt *options.DBOptions, input apis.ServiceDBOptions) {
	opt.SqlConnection = input.ToSQLConnection()
}

func GetAuthURL(config apis.Keystone, address string) string {
	proto := "http"
	if config.EnableSSL {
		proto = "https"
	}
	return fmt.Sprintf("%s://%s:%d/v3", proto, address, config.ServiceBaseOptions.Port)
}

func FillServiceCommonOptions(opt *apis.ServiceCommonOptions, authConfig apis.Keystone, address string) {
	opt.AuthURL = GetAuthURL(authConfig, address)
}
