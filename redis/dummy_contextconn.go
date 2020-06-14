package redis

import (
	context "context"
)

type DummyContextConn struct{}

func (dcc DummyContextConn) Close() error {
	var r0 error
	return r0
}

func (dcc DummyContextConn) Do(ctx context.Context,
	commandName string, args ...interface{}) (reply interface{}, err error) {
	return
}

func (dcc DummyContextConn) Err() error {
	var r0 error
	return r0
}

func (dcc DummyContextConn) Flush(ctx context.Context) error {
	var r0 error
	return r0
}

func (dcc DummyContextConn) Receive(ctx context.Context) (reply interface{}, err error) {
	return
}

func (dcc DummyContextConn) Send(ctx context.Context, commandName string,
	args ...interface{}) error {
	var r0 error
	return r0
}
