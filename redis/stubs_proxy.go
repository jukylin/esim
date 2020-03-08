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

type stubsProxyOption func(c *stubsProxy)

type stubsProxyOptions struct{}

func NewStubsProxy(logger log.Logger, name string) *stubsProxy {
	stubsProxy := &stubsProxy{}

	if logger == nil {
		stubsProxy.log = log.NewLogger()
	}

	stubsProxy.name = name

	return stubsProxy
}

//implement Proxy interface
func (this *stubsProxy) NextProxy(conn interface{}) {
	this.nextConn = conn.(ContextConn)
}

//implement Proxy interface
func (this *stubsProxy) ProxyName() string {
	return this.name
}

func (this *stubsProxy) Close() error {
	err := this.nextConn.Close()

	return err
}

func (this *stubsProxy) Err() (err error) {
	err = this.nextConn.Err()
	return
}

func (this *stubsProxy) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	if args[0] == "name" {
		return "test", nil
	}

	if args[0] == "version" {
		return "2.0", nil
	}

	return
}

func (this *stubsProxy) Send(ctx context.Context, commandName string, args ...interface{}) (err error) {
	err = this.nextConn.Send(ctx, commandName, args...)

	return
}

func (this *stubsProxy) Flush(ctx context.Context) (err error) {
	err = this.nextConn.Flush(ctx)

	return
}

func (this *stubsProxy) Receive(ctx context.Context) (reply interface{}, err error) {
	reply, err = this.nextConn.Receive(ctx)

	return
}
