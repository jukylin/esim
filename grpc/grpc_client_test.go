package grpc

import (
	"testing"
	"context"
	"time"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/config"
)

func TestNewGrpcClient(t *testing.T) {

	go func() {
		startServer()
	}()

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))

	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOptions := ClientOptions{}
	grpcClient := NewGrpcClient(
		clientOptions.WithClientLogger(logger),
		clientOptions.WithClientConf(memConfig),
	)

	ctx := context.Background()
	conn := grpcClient.DialContext(ctx, ":50051")

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "esim"})
	if err != nil {
		logger.Errorf(err.Error())
	}else {
		assert.NotEmpty(t, r.Message)
	}
}



func TestSlowClient(t *testing.T) {

	go func() {
		startServer()
	}()

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))

	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)
	memConfig.Set("grpc_client_check_slow", true)
	memConfig.Set("grpc_client_slow_time", 10)

	clientOptions := ClientOptions{}
	grpcClient := NewGrpcClient(
		clientOptions.WithClientLogger(logger),
		clientOptions.WithClientConf(memConfig),
		clientOptions.WithClientDialOptions(
			grpc.WithBlock(),
			grpc.WithChainUnaryInterceptor(slowRequest),
			),
	)

	ctx := context.Background()
	conn := grpcClient.DialContext(ctx, ":50051")

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "esim"})
	if err != nil {
		logger.Errorf(err.Error())
	}else {
		assert.NotEmpty(t, r.Message)
	}
}



func TestServerPanic(t *testing.T) {

	go func() {
		startPanicServer()
	}()

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))

	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOptions := ClientOptions{}
	grpcClient := NewGrpcClient(
		clientOptions.WithClientLogger(logger),
		clientOptions.WithClientConf(memConfig),
	)

	ctx := context.Background()
	conn := grpcClient.DialContext(ctx, ":50051")

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "esim"})
	if err != nil {
		logger.Errorf(err.Error())
	}else {
		assert.NotEmpty(t, r.Message)
	}
}




func TestServerPanicArr(t *testing.T) {

	go func() {
		startPanicArrServer()
	}()

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))

	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOptions := ClientOptions{}
	grpcClient := NewGrpcClient(
		clientOptions.WithClientLogger(logger),
		clientOptions.WithClientConf(memConfig),
	)

	ctx := context.Background()
	conn := grpcClient.DialContext(ctx, ":50051")

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "esim"})
	if err != nil {
		logger.Errorf(err.Error())
	}else {
		assert.NotEmpty(t, r.Message)
	}
}
