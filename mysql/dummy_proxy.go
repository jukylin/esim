//nolint:unused,deadcode
package mysql

import (
	"context"
	"database/sql"

	"github.com/jukylin/esim/log"
)

// dummyProxy must as last proxy.
type dummyProxy struct {
	nextProxy ConnPool

	logger log.Logger

	name string
}

//nolint:unused
func newDummyProxy(logger log.Logger, name string) ConnPool {
	dummyProxy := &dummyProxy{}

	dummyProxy.logger = logger

	dummyProxy.name = name
	return dummyProxy
}

// Implement Proxy interface.
func (dp *dummyProxy) NextProxy(db interface{}) {
	dp.nextProxy = db.(ConnPool)
}

// Implement Proxy interface.
func (dp *dummyProxy) ProxyName() string {
	return dp.name
}

func (dp *dummyProxy) ExecContext(ctx context.Context,
	query string, args ...interface{}) (sql.Result, error) {
	result := &dummySQLResult{}
	return result, nil
}

func (dp *dummyProxy) PrepareContext(ctx context.Context,
	query string) (*sql.Stmt, error) {
	stmt := &sql.Stmt{}

	return stmt, nil
}

func (dp *dummyProxy) QueryContext(ctx context.Context, query string,
	args ...interface{}) (*sql.Rows, error) {
	rows := &sql.Rows{}
	return rows, nil
}

func (dp *dummyProxy) QueryRowContext(ctx context.Context, query string,
	args ...interface{}) *sql.Row {
	row := &sql.Row{}
	return row
}

func (dp *dummyProxy) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return dp.nextProxy.BeginTx(ctx, opts)
}

func (dp *dummyProxy) Close() error {
	return nil
}

// Implement sql.Result interface.
type dummySQLResult struct {
}

// Implement sql.Result interface.
func (dp *dummySQLResult) LastInsertId() (int64, error) {
	return 0, nil
}

// Implement sql.Result interface.
func (dp *dummySQLResult) RowsAffected() (int64, error) {
	return 0, nil
}
