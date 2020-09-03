package mysql

import (
	"gorm.io/gorm"
)

type ConnPool interface {
	gorm.ConnPool

	SQLClose
}

type SQLClose interface {
	Close() error
}
