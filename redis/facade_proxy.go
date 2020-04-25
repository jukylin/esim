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
func (fp *FacadeProxy) NextProxy(conn interface{}) {
	fp.nextConn = conn.(redis.Conn)
}

//implement Proxy interface
func (fp *FacadeProxy) ProxyName() string {
	return fp.name
}

func (fp *FacadeProxy) Close() error {
	err := fp.nextConn.Close()
	if err == redis.ErrNil {
		err = nil
	}
	return err
}

func (fp *FacadeProxy) Err() (err error) {
	err = fp.nextConn.Err()
	return
}

func (fp *FacadeProxy) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	reply, err = fp.nextConn.Do(commandName, args...)
	return
}

func (fp *FacadeProxy) Send(ctx context.Context, commandName string, args ...interface{}) (err error) {
	err = fp.nextConn.Send(commandName, args...)
	return
}

func (fp *FacadeProxy) Flush(ctx context.Context) (err error) {
	err = fp.nextConn.Flush()
	return
}

func (fp *FacadeProxy) Receive(ctx context.Context) (reply interface{}, err error) {

	reply, err = fp.nextConn.Receive()

	return
}
