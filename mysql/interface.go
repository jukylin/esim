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