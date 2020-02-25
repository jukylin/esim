package mysql

import (
	"github.com/jinzhu/gorm"
)

type SqlCommon interface {
	gorm.CommonDB

	gorm.TxBeginner

	sqlClose
}

type sqlClose interface {
	Close() error
}

type sqlPing interface {
	Ping() error
}