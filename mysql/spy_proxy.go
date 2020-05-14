package mysql

import (
	"database/sql"

	"github.com/jukylin/esim/log"
)

type spyProxy struct {
	nextProxy SQLCommon

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
func (sp *spyProxy) NextProxy(db interface{}) {
	sp.nextProxy = db.(SQLCommon)
}

//implement Proxy interface
func (sp *spyProxy) ProxyName() string {
	return sp.name
}

func (sp *spyProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	sp.ExecWasCalled = true
	sp.logger.Infof("%s ExecWasCalled %s", sp.name, query)
	result, err := sp.nextProxy.Exec(query, args...)
	return result, err
}

func (sp *spyProxy) Prepare(query string) (*sql.Stmt, error) {
	sp.PrepareWasCalled = true
	sp.logger.Infof("%s PrepareWasCalled %s", sp.name, query)
	stmt, err := sp.nextProxy.Prepare(query)

	return stmt, err
}

func (sp *spyProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	sp.QueryWasCalled = true
	sp.logger.Infof("%s QueryWasCalled %s", sp.name, query)
	rows, err := sp.nextProxy.Query(query, args...)
	return rows, err
}

func (sp *spyProxy) QueryRow(query string, args ...interface{}) *sql.Row {
	sp.QueryRowWasCalled = true
	sp.logger.Infof("%s QueryRowWasCalled %s", sp.name, query)
	row := sp.nextProxy.QueryRow(query, args...)
	return row
}

func (sp *spyProxy) Close() error {
	return sp.nextProxy.Close()
}
