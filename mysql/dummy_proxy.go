//nolint:unused,deadcode
package mysql

import (
	"database/sql"

	"github.com/jukylin/esim/log"
)

// dummyProxy must as last proxy.
type dummyProxy struct {
	nextProxy ConnPool

	logger log.Logger

	name string
}

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

func (this *dummyProxy) ExecContext(query string, args ...interface{}) (sql.Result, error) {
	result := &dummySQLResult{}
	return result, nil
}

func (this *dummyProxy) PrepareContext(query string) (*sql.Stmt, error) {
	stmt := &sql.Stmt{}

	return stmt, nil
}

func (this *dummyProxy) QueryContext(query string, args ...interface{}) (*sql.Rows, error) {
	rows := &sql.Rows{}
	return rows, nil
}

func (this *dummyProxy) QueryRowContext(query string, args ...interface{}) *sql.Row {
	row := &sql.Row{}
	return row
}

func (this *dummyProxy) Close() error {
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
