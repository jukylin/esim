package mysql

import (
	"database/sql"
	"github.com/jukylin/esim/log"
)

type spyProxy struct {
	nextProxy SqlCommon

	ExecWasCalled bool

	PrepareWasCalled bool

	QueryWasCalled bool

	QueryRowWasCalled bool

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

func (this *spyProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	this.ExecWasCalled = true
	this.logger.Infof("%s ExecWasCalled %s", this.name, query)
	result, err := this.nextProxy.Exec(query, args...)
	return result, err
}

func (this *spyProxy) Prepare(query string) (*sql.Stmt, error) {
	this.PrepareWasCalled = true
	this.logger.Infof("%s PrepareWasCalled %s", this.name, query)
	stmt, err := this.nextProxy.Prepare(query)

	return stmt, err
}

func (this *spyProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	this.QueryWasCalled = true
	this.logger.Infof("%s QueryWasCalled %s", this.name, query)
	rows, err := this.nextProxy.Query(query, args...)
	return rows, err
}

func (this *spyProxy) QueryRow(query string, args ...interface{}) *sql.Row {
	this.QueryRowWasCalled = true
	this.logger.Infof("%s QueryRowWasCalled %s", this.name, query)
	row := this.nextProxy.QueryRow(query, args...)
	return row
}

func (this *spyProxy) Close() error {
	return this.nextProxy.Close()
}
