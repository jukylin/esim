package mysql

import (
	"context"
	"github.com/jinzhu/gorm"
)

type MysqlClient interface {

	GetDb(string) *gorm.DB

	GetCtxDb(context.Context, string) *gorm.DB

	Ping() []error

	Close()
}


type SqlCommon interface {
	gorm.SQLCommon
	sqlClose
}

type sqlClose interface {
	Close() error
}

type sqlPing interface {
	Ping() error
}