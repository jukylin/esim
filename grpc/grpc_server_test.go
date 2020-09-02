package grpc

import (
	"os"
	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/opentracing"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer.
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func TestMain(m *testing.M) {
	serverOptions := ServerOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_server_debug", true)
	memConfig.Set("grpc_server_metrics", true)
	memConfig.Set("grpc_server_trace", true)
	memConfig.Set("grpc_server_check_slow", true)

	svr := NewServer(tcpAddr.String(),
		serverOptions.WithServerLogger(logger),
		serverOptions.WithServerConf(memConfig),
		serverOptions.WithTracer(opentracing.NewTracer("grcp_server_test", logger)),
		serverOptions.WithUnarySrvItcp(
			ServerStubs(func(ctx context.Context, req interface{},
				info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				if req.(*pb.HelloRequest).Name == callPanic {
					panic(isTest)
				} else if req.(*pb.HelloRequest).Name == callPanicArr {
					var arr [1]string
					arr[0] = isTest
					panic(arr)
				} else if req.(*pb.HelloRequest).Name == callNil {
					return nil, err
				}
				resp, err = handler(ctx, req)

				return resp, err
			}),
		))

	pb.RegisterGreeterServer(svr.Server, &server{})

	svr.Start()

	code := m.Run()

	svr.GracefulShutDown()

	os.Exit(code)
}
