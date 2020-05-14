package mysql

import (
	"context"
	"database/sql"

	"github.com/jinzhu/gorm"
)

type SQLCommon interface {
	gorm.SQLCommon

	SQLClose

	Begin() (*sql.Tx, error)

	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type SQLClose interface {
	Close() error
}
