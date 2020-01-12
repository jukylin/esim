package redis

import (
	"context"
	"time"
)

type RedisExecInfo struct {
	err error

	commandName string

	startTime time.Time

	endTime time.Time

	args []interface{}
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
