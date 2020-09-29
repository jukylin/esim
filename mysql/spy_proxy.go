package mysql

import (
	"context"
	"database/sql"

	"github.com/jukylin/esim/log"
)

type spyProxy struct {
	nextProxy ConnPool

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

// implement Proxy interface.
func (sp *spyProxy) NextProxy(db interface{}) {
	sp.nextProxy = db.(ConnPool)
}

// implement Proxy interface.
func (sp *spyProxy) ProxyName() string {
	return sp.name
}

func (sp *spyProxy) ExecContext(ctx context.Context, query string,
	args ...interface{}) (sql.Result, error) {
	sp.ExecWasCalled = true
	sp.logger.Infof("%s ExecContextWasCalled %s", sp.name, query)
	result, err := sp.nextProxy.ExecContext(ctx, query, args...)
	return result, err
}

func (sp *spyProxy) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	sp.PrepareWasCalled = true
	sp.logger.Infof("%s PrepareContextWasCalled %s", sp.name, query)
	stmt, err := sp.nextProxy.PrepareContext(ctx, query)
	return stmt, err
}

func (sp *spyProxy) QueryContext(ctx context.Context, query string,
	args ...interface{}) (*sql.Rows, error) {
	sp.QueryWasCalled = true
	sp.logger.Infof("%s QueryContextWasCalled %s", sp.name, query)
	rows, err := sp.nextProxy.QueryContext(ctx, query, args...)
	return rows, err
}

func (sp *spyProxy) QueryRowContext(ctx context.Context, query string,
	args ...interface{}) *sql.Row {
	sp.QueryRowWasCalled = true
	sp.logger.Infof("%s QueryContextRowWasCalled %s", sp.name, query)
	row := sp.nextProxy.QueryRowContext(ctx, query, args...)
	return row
}

func (sp *spyProxy) Close() error {
	sp.CloseWasCalled = true
	sp.logger.Infof("%s CloseWasCalled %s", sp.name)
	return sp.nextProxy.Close()
}
