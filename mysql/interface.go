package mysql

import (
	"context"
	"database/sql"

	"github.com/jinzhu/gorm"
)

type SQLCommon interface {
	gorm.SQLCommon

	sqlClose

	Begin() (*sql.Tx, error)

	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type sqlClose interface {
	Close() error
}
