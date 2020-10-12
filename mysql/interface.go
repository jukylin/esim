package mysql

import (
	"context"
	"gorm.io/gorm"
	"database/sql"
)

type ConnPool interface {
	gorm.ConnPool

	SQLClose

	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type SQLClose interface {
	Close() error
}
