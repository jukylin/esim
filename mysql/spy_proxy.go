package mysql

import (
	"context"
	"database/sql"
	"github.com/jukylin/esim/log"
)

type spyProxy struct {
	nextProxy SqlCommon

	ExecWasCalled bool

	PrepareWasCalled bool

	QueryWasCalled bool

	QueryRowWasCalled bool

	CloseWasCalled bool

	BeginTxWasCalled bool

	logger log.Logger

	name string
}


func newSpyProxy(logger log.Logger, name string) *spyProxy {
	spyProxy := &spyProxy{}

	spyProxy.logger = logger

	spyProxy.name = name
	return spyProxy
}


//implement Proxy interface
func (this *spyProxy) NextProxy(db interface{}) {
	this.nextProxy = db.(SqlCommon)
}


//implement Proxy interface
func (this *spyProxy) ProxyName() string {
	return this.name
}


func (this *spyProxy) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	this.ExecWasCalled = true
	this.logger.Infof("%s ExecContextWasCalled %s", this.name, query)
	result, err := this.nextProxy.ExecContext(ctx, query, args...)
	return result, err
}


func (this *spyProxy) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	this.PrepareWasCalled = true
	this.logger.Infof("%s PrepareContextWasCalled %s", this.name, query)
	stmt, err := this.nextProxy.PrepareContext(ctx, query)

	return stmt, err
}


func (this *spyProxy) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	this.QueryWasCalled = true
	this.logger.Infof("%s QueryContextWasCalled %s", this.name, query)
	rows, err := this.nextProxy.QueryContext(ctx, query, args...)
	return rows, err
}


func (this *spyProxy) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	this.QueryRowWasCalled = true
	this.logger.Infof("%s QueryContextRowWasCalled %s", this.name, query)
	row := this.nextProxy.QueryRowContext(ctx, query, args...)
	return row
}


func (this *spyProxy) Close() error {
	this.CloseWasCalled = true
	this.logger.Infof("%s CloseWasCalled %s", this.name)
	return this.nextProxy.Close()
}


func (this *spyProxy) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error){
	this.BeginTxWasCalled = true
	this.logger.Infof("%s BeginTxWasCalled %s", this.name)
	return this.nextProxy.BeginTx(ctx, opts)
}