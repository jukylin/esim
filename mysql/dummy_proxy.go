package mysql

import (
	"database/sql"

	"github.com/jukylin/esim/log"
)

//last proxy
type dummyProxy struct {
	nextProxy SqlCommon

	logger log.Logger

	name string
}

func newDummyProxy(logger log.Logger, name string) *dummyProxy {
	dummyProxy := &dummyProxy{}

	dummyProxy.logger = logger

	dummyProxy.name = name
	return dummyProxy
}

//implement Proxy interface
func (dp *dummyProxy) NextProxy(db interface{}) {
	dp.nextProxy = db.(SqlCommon)
}

//implement Proxy interface
func (dp *dummyProxy) ProxyName() string {
	return dp.name
}

func (dp *dummyProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	result := &dummySqlResult{}
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
type dummySqlResult struct {
}

// implement sql.Result interface
func (dp *dummySqlResult) LastInsertId() (int64, error) {
	return 0, nil
}

// implement sql.Result interface
func (dp *dummySqlResult) RowsAffected() (int64, error) {
	return 0, nil
}
