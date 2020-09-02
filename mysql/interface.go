package mysql

import (
	"github.com/go-gorm/gorm"
)

type ConnPool interface {
	gorm.ConnPool



	SQLClose
}

type SQLClose interface {
	Close() error
}
