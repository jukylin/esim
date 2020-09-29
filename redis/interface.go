package redis

import (
	"context"
	"sync"
	"time"
)

var (
	execInfoPool = sync.Pool{
		New: func() interface{} {
			return &execInfo{}
		},
	}
)

type execInfo struct {
	err error

	commandName string

	startTime time.Time

	endTime time.Time

	args []interface{}

	reply interface{}

	method string
}

func newExecInfo() *execInfo {
	ei := execInfoPool.Get().(*execInfo)

	if ei.args == nil {
		ei.args = make([]interface{}, 0)
	}

	return ei
}

// Release initializes the variable and put that to tool.
func (ei *execInfo) Release() {
	ei.err = nil
	ei.commandName = ""
	ei.method = ""
	ei.reply = nil
	ei.startTime = time.Time{}
	ei.endTime = time.Time{}
	ei.args = ei.args[:0]
	execInfoPool.Put(ei)
}

//  CtxConn redefine redis.Conn, Implemented by *redis.Conn..
type ContextConn interface {
	// Close closes the connection.
	Close() error

	// Err returns a non-nil value when the connection is not usable.
	Err() error

	// Do sends a command to the server and returns the received reply.
	Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error)

	// Send writes the command to the client's output buffer.
	Send(ctx context.Context, commandName string, args ...interface{}) error

	// Flush flushes the output buffer to the Redis server.
	Flush(ctx context.Context) error

	// Receive receives a single reply from the Redis server
	Receive(ctx context.Context) (reply interface{}, err error)
}
