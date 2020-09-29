package grpc

import (
	"context"
	"testing"

	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
)

func TestClientConn_Close(t *testing.T) {
	logger := log.NewLogger()
	clientConn := NewClientConn(logger)

	clientOptions := NewClientOptions()
	client := NewClient(clientOptions)
	ctx := context.Background()
	conn := client.DialContext(ctx, ":50051")
	clientConn.CollectConn(conn)

	conn = client.DialContext(ctx, ":50052")
	clientConn.CollectConn(conn)

	stats := clientConn.State()
	assert.Equal(t, int(2), len(stats))
	logger.Infof(" stats %+v", stats)
	clientConn.Close()
}
