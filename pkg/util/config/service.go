package config

import (
	"github.com/pkg/errors"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/util/mysql"
)

func InitDBUser(conn *mysql.Connection, info apis.DBInfo) error {
	dbName := info.Database
	dbExists, err := conn.IsDatabaseExists(dbName)
	if err != nil {
		return errors.Wrap(err, "check db exists")
	}
	if !dbExists {
		//return errors.Errorf("database %q already exists", dbName)
		if err := conn.CreateDatabase(dbName); err != nil {
			return errors.Wrapf(err, "create database %q", dbName)
		}
	}
	user := info.Username
	password := info.Password
	if err := conn.CreateUser(user, password, dbName); err != nil {
		return errors.Wrapf(err, "create user %q for database %q", user, dbName)
	}
	return nil
}
