package redis

import (
	"context"
	"github.com/jukylin/esim/log"
)

type spyProxy struct {

	DoWasCalled bool

	SendWasCalled bool

	FlushWasCalled bool

	ReceiveWasCalled bool
	
	nextConn ContextConn
	
	name string

	log log.Logger

}

type spyProxyOption func(c *spyProxy)

type spyProxyOptions struct{}

func NewSpyProxy(logger log.Logger, name string) *spyProxy {
	spyProxy := &spyProxy{}

	if logger == nil{
		spyProxy.log = log.NewLogger()
	}

	spyProxy.name = name

	return spyProxy
}


//implement Proxy interface
func (this *spyProxy) NextProxy(conn interface{}) {
	this.nextConn = conn.(ContextConn)
}

//implement Proxy interface
func (this *spyProxy) ProxyName() string {
	return this.name
}

func (this *spyProxy) Close() error {
	err := this.nextConn.Close()

	return err
}

func (this *spyProxy) Err() (err error) {
	err = this.nextConn.Err()
	return
}

func (this *spyProxy) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	this.DoWasCalled = true
	reply, err = this.nextConn.Do(ctx, commandName, args...)

	return
}

func (this *spyProxy) Send(ctx context.Context, commandName string, args ...interface{}) (err error) {
	this.SendWasCalled = true
	err = this.nextConn.Send(ctx, commandName, args...)

	return
}

func (this *spyProxy) Flush(ctx context.Context) (err error) {
	this.FlushWasCalled = true
	err = this.nextConn.Flush(ctx)

	return
}

func (this *spyProxy) Receive(ctx context.Context) (reply interface{}, err error) {
	this.ReceiveWasCalled = true
	reply, err = this.nextConn.Receive(ctx)

	return
}