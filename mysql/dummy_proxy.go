//nolint:unused,deadcode
package mysql

import (
	"database/sql"
	"context"

	"github.com/jukylin/esim/log"
)

//last proxy
type dummyProxy struct {
	nextProxy SQLCommon

	logger log.Logger

	name string
}

func newDummyProxy(logger log.Logger, name string) SQLCommon {
	dummyProxy := &dummyProxy{}

	dummyProxy.logger = logger

	dummyProxy.name = name
	return dummyProxy
}

//implement Proxy interface
func (dp *dummyProxy) NextProxy(db interface{}) {
	dp.nextProxy = db.(SQLCommon)
}

//implement Proxy interface
func (dp *dummyProxy) ProxyName() string {
	return dp.name
}

func (dp *dummyProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	result := &dummySQLResult{}
	return result, nil
}

func (dp *dummyProxy) Prepare(query string) (*sql.Stmt, error) {
	stmt := &sql.Stmt{}

	return stmt, nil
}

func (dp *dummyProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows := &sql.Rows{}
	return rows, nil
}

func (dp *dummyProxy) QueryRow(query string, args ...interface{}) *sql.Row {
	row := &sql.Row{}
	return row
}

func (dp *dummyProxy) Close() error {
	return dp.nextProxy.Close()
}

// implement sql.Result interface
type dummySQLResult struct {
}

// implement sql.Result interface
func (dp *dummySQLResult) LastInsertId() (int64, error) {
	return 0, nil
}

// implement sql.Result interface
func (dp *dummySQLResult) RowsAffected() (int64, error) {
	return 0, nil
}


func (dp *dummyProxy) Begin() (*sql.Tx, error) {
	return dp.nextProxy.Begin()
}

func (dp *dummyProxy) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return dp.nextProxy.BeginTx(ctx, opts)
}