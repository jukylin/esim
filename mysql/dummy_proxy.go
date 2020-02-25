package mysql

import (
	"context"
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

func (this *dummyProxy) ExecContext(query string, args ...interface{}) (sql.Result, error) {
	result := &dummySqlResult{}
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


func (this *dummyProxy) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error){
	return &sql.Tx{}, nil
}

// implement sql.Result interface
type dummySqlResult struct {
}

// implement sql.Result interface
func (this *dummySqlResult) LastInsertId() (int64, error){
	return 0, nil
}

// implement sql.Result interface
func (this *dummySqlResult) RowsAffected() (int64, error){
	return 0, nil
}