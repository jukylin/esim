package redis

import (
	"context"
	"github.com/gomodule/redigo/redis"
)

// FacadeProxy implement ContextConn interface, but nextConn is redis.Conn
type FacadeProxy struct {
	nextConn redis.Conn

	name string
}

type FacadeProxyOption func(c *FacadeProxy)

type FacadeProxyOptions struct{}

func NewFacadeProxy(options ...FacadeProxyOption) *FacadeProxy {
	FacadeProxy := &FacadeProxy{}
	for _, option := range options {
		option(FacadeProxy)
	}

	FacadeProxy.name = "Facade_proxy"
	return FacadeProxy
}

//implement Proxy interface
func (this *FacadeProxy) NextProxy(conn interface{}) {
	this.nextConn = conn.(redis.Conn)
}

//implement Proxy interface
func (this *FacadeProxy) ProxyName() string {
	return this.name
}

func (this *FacadeProxy) Close() error {
	err := this.nextConn.Close()
	if err == redis.ErrNil {
		err = nil
	}
	return err
}

func (this *FacadeProxy) Err() (err error) {
	err = this.nextConn.Err()
	return
}

func (this *FacadeProxy) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	reply, err = this.nextConn.Do(commandName, args...)
	return
}

func (this *FacadeProxy) Send(ctx context.Context, commandName string, args ...interface{}) (err error) {
	err = this.nextConn.Send(commandName, args...)
	return
}

func (this *FacadeProxy) Flush(ctx context.Context) (err error) {
	err = this.nextConn.Flush()
	return
}

func (this *FacadeProxy) Receive(ctx context.Context) (reply interface{}, err error) {

	reply, err = this.nextConn.Receive()

	return
}
