// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gorm

import (
	"errors"
	v2 "github.com/cebrains/jupiter/pkg/store/gorm/v2"
	"github.com/cebrains/jupiter/pkg/util/xdebug"
	"github.com/cebrains/jupiter/pkg/xlog"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	gormLog "gorm.io/gorm/logger"
	stdLog "log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
	// mysql driver
	_ "gorm.io/driver/mysql"
	// postgres driver
	_ "gorm.io/driver/postgres"
)

// SQLCommon ...
type (
	// Scope ...
	Scope = gorm.Statement
	// DB ...
	DB = gorm.DB
	// Model ...
	Model = gorm.Model
	// Association ...
	Association = gorm.Association
)

var (
	errSlowCommand = errors.New("database query slow command")

	// ErrRecordNotFound returns a "record not found error". Occurs only when attempting to query the database with a struct; querying with a slice won't return this error
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

// Open ...
func Open(dialect string, options *Config) (*DB, error) {
	var dbDialector gorm.Dialector
	if val, err := getDbDialector(dialect, "Write", options.DSN); err != nil {
		xlog.Panic("database not exists", xlog.FieldMod("gorm"), xlog.FieldErr(err), xlog.FieldKey(dialect))
	} else {
		dbDialector = val
	}

	logLevel := gormLog.Error
	if options.Debug || xdebug.IsDevelopmentMode() {
		logLevel = gormLog.Info
	}
	inner, err := gorm.Open(dbDialector, &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 redefineLog(logLevel), //拦截、接管 gorm v2 自带日志
	})
	if err != nil {
		//gorm 数据库驱动初始化失败
		return nil, err
	}

	// 为主连接设置连接池()
	if rawDb, err := inner.DB(); err != nil {
		return nil, err
	} else {
		rawDb.SetConnMaxIdleTime(time.Second * 30)
		rawDb.SetConnMaxLifetime(options.ConnMaxLifetime)
		rawDb.SetMaxIdleConns(options.MaxIdleConns)
		rawDb.SetMaxOpenConns(options.MaxOpenConns)
		return inner, nil
	}
}

// 获取一个数据库方言(Dialector),通俗的说就是根据不同的连接参数，获取具体的一类数据库的连接指针
func getDbDialector(dialect, readWrite, dsn string) (gorm.Dialector, error) {
	var dbDialector gorm.Dialector
	switch strings.ToLower(dialect) {
	case "mysql":
		dbDialector = mysql.Open(dsn)
	case "postgres", "postgresql":
		dbDialector = postgres.Open(dsn)
	default:
		return nil, errors.New("database not exists")
	}
	return dbDialector, nil
}

// 创建自定义日志模块，对 gorm 日志进行拦截、
func redefineLog(logLevel gormLog.LogLevel) gormLog.Interface {
	return v2.New(stdLog.New(os.Stdout, "\r\n", stdLog.LstdFlags), &gormLog.Config{
		LogLevel:      logLevel,
		Colorful:      true,
		SlowThreshold: 100 * time.Millisecond,
	})
}
