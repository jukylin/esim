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

func NewSpyProxy(logger log.Logger, name string) *spyProxy {
	spyProxy := &spyProxy{}

	if logger == nil {
		spyProxy.log = log.NewLogger()
	}

	spyProxy.name = name

	return spyProxy
}

//implement Proxy interface
func (sp *spyProxy) NextProxy(conn interface{}) {
	sp.nextConn = conn.(ContextConn)
}

//implement Proxy interface
func (sp *spyProxy) ProxyName() string {
	return sp.name
}

func (sp *spyProxy) Close() error {
	err := sp.nextConn.Close()

	return err
}

func (sp *spyProxy) Err() (err error) {
	err = sp.nextConn.Err()
	return
}

func (sp *spyProxy) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	sp.DoWasCalled = true
	reply, err = sp.nextConn.Do(ctx, commandName, args...)

	return
}

func (sp *spyProxy) Send(ctx context.Context, commandName string, args ...interface{}) (err error) {
	sp.SendWasCalled = true
	err = sp.nextConn.Send(ctx, commandName, args...)

	return
}

func (sp *spyProxy) Flush(ctx context.Context) (err error) {
	sp.FlushWasCalled = true
	err = sp.nextConn.Flush(ctx)

	return
}

func (sp *spyProxy) Receive(ctx context.Context) (reply interface{}, err error) {
	sp.ReceiveWasCalled = true
	reply, err = sp.nextConn.Receive(ctx)

	return
}
