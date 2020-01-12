package grpc

import (
	"net"
	"context"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/reflection"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/config"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func startServer(){

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	serverOptions := ServerOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_server_debug", true)

	s := NewGrpcServer("test",
		serverOptions.WithServerLogger(logger),
		serverOptions.WithServerConf(memConfig),
			)

	pb.RegisterGreeterServer(s.Server, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s.Server)
	if err := s.Server.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}



func startPanicServer(){

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	serverOptions := ServerOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_server_debug", true)

	s := NewGrpcServer("test",
		serverOptions.WithServerLogger(logger),
		serverOptions.WithServerConf(memConfig),
		serverOptions.WithUnarySrvItcp(
			panicResp(),
			),
	)

	pb.RegisterGreeterServer(s.Server, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s.Server)
	if err := s.Server.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}



func startPanicArrServer(){

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	serverOptions := ServerOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_server_debug", true)

	s := NewGrpcServer("test",
		serverOptions.WithServerLogger(logger),
		serverOptions.WithServerConf(memConfig),
		serverOptions.WithUnarySrvItcp(
			panicArrayResp(),
		),
	)

	pb.RegisterGreeterServer(s.Server, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s.Server)
	if err := s.Server.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}