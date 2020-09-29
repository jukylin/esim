package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

var (
	logger log.Logger

	tcpAddr = &net.TCPAddr{IP: net.ParseIP(address).To4(), Port: port}
)

func TestNewGrpcClient(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOptional := ClientOptionals{}
	clientOptions := NewClientOptions(
		clientOptional.WithLogger(logger),
		clientOptional.WithConf(memConfig),
	)

	ctx := context.Background()
	client := NewClient(clientOptions)
	conn := client.DialContext(ctx, tcpAddr.String())

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.NotEmpty(t, r.Message)
	}
}

func TestSlowClient(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)
	memConfig.Set("grpc_client_check_slow", true)
	memConfig.Set("grpc_client_slow_time", 10)

	clientOptional := ClientOptionals{}
	clientOptions := NewClientOptions(
		clientOptional.WithLogger(logger),
		clientOptional.WithConf(memConfig),
		clientOptional.WithDialOptions(
			grpc.WithBlock(),
			grpc.WithChainUnaryInterceptor(slowRequest),
		),
	)

	ctx := context.Background()
	client := NewClient(clientOptions)
	conn := client.DialContext(ctx, tcpAddr.String())

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.NotEmpty(t, r.Message)
	}
}

func TestServerPanic(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOptional := ClientOptionals{}
	clientOptions := NewClientOptions(
		clientOptional.WithLogger(logger),
		clientOptional.WithConf(memConfig),
	)

	ctx := context.Background()
	client := NewClient(clientOptions)
	conn := client.DialContext(ctx, tcpAddr.String())

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: callPanic})
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestServerPanicArr(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOptional := ClientOptionals{}
	clientOptions := NewClientOptions(
		clientOptional.WithLogger(logger),
		clientOptional.WithConf(memConfig),
	)

	ctx := context.Background()
	client := NewClient(clientOptions)
	conn := client.DialContext(ctx, tcpAddr.String())

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: callPanicArr})
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestSubsReply(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOptional := ClientOptionals{}
	clientOptions := NewClientOptions(
		clientOptional.WithLogger(logger),
		clientOptional.WithConf(memConfig),
		clientOptional.WithDialOptions(
			grpc.WithChainUnaryInterceptor(ClientStubs(func(ctx context.Context,
				method string, req, reply interface{}, cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				if method == "/helloworld.Greeter/SayHello" {
					reply.(*pb.HelloReply).Message = esim
				}
				return nil
			})),
		),
	)

	ctx := context.Background()
	client := NewClient(clientOptions)
	conn := client.DialContext(ctx, tcpAddr.String())

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.Equal(t, esim, r.Message)
	}
}

func TestGlobalSubs(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOptional := ClientOptionals{}
	GlobalStub = func(ctx context.Context,
		method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if method == "/helloworld.Greeter/SayHello" {
			reply.(*pb.HelloReply).Message = esim
		}
		return nil
	}

	ctx := context.Background()
	client := NewClientWithOptionals(
		clientOptional.WithLogger(logger),
		clientOptional.WithConf(memConfig),
	)
	conn := client.DialContext(ctx, tcpAddr.String())

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.Equal(t, esim, r.Message)
	}
}
