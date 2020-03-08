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
func (this *dummyProxy) NextProxy(db interface{}) {
	this.nextProxy = db.(SqlCommon)
}

//implement Proxy interface
func (this *dummyProxy) ProxyName() string {
	return this.name
}

func (this *dummyProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	result := &dummySqlResult{}
	return result, nil
}

func (this *dummyProxy) Prepare(query string) (*sql.Stmt, error) {
	stmt := &sql.Stmt{}

	return stmt, nil
}

func (this *dummyProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows := &sql.Rows{}
	return rows, nil
}

func (this *dummyProxy) QueryRow(query string, args ...interface{}) *sql.Row {
	row := &sql.Row{}
	return row
}

func (this *dummyProxy) Close() error {
	return this.nextProxy.Close()
}

// implement sql.Result interface
type dummySqlResult struct {
}

// implement sql.Result interface
func (this *dummySqlResult) LastInsertId() (int64, error) {
	return 0, nil
}

// implement sql.Result interface
func (this *dummySqlResult) RowsAffected() (int64, error) {
	return 0, nil
}
