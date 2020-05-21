package redis

import (
	"context"

	"github.com/jukylin/esim/log"
)

type stubsProxy struct {
	nextConn ContextConn

	name string

	log log.Logger
}

func newStubsProxy(logger log.Logger, name string) *stubsProxy {
	stubsProxy := &stubsProxy{}

	if logger == nil {
		stubsProxy.log = log.NewLogger()
	}

	stubsProxy.name = name

	return stubsProxy
}

// implement Proxy interface
func (sp *stubsProxy) NextProxy(conn interface{}) {
	sp.nextConn = conn.(ContextConn)
}

// implement Proxy interface
func (sp *stubsProxy) ProxyName() string {
	return sp.name
}

func (sp *stubsProxy) Close() error {
	err := sp.nextConn.Close()

	return err
}

func (sp *stubsProxy) Err() (err error) {
	err = sp.nextConn.Err()
	return
}

func (sp *stubsProxy) Do(ctx context.Context, commandName string,
	args ...interface{}) (reply interface{}, err error) {
	if args[0] == "name" {
		return "test", nil
	}

	if args[0] == "version" {
		return "2.0", nil
	}

	return
}

func (sp *stubsProxy) Send(ctx context.Context, commandName string,
	args ...interface{}) (err error) {
	err = sp.nextConn.Send(ctx, commandName, args...)

	return
}

func (sp *stubsProxy) Flush(ctx context.Context) (err error) {
	err = sp.nextConn.Flush(ctx)

	return
}

func (sp *stubsProxy) Receive(ctx context.Context) (reply interface{}, err error) {
	reply, err = sp.nextConn.Receive(ctx)

	return
}
