package grpc

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	ggp "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type GrpcServer struct {
	Server *grpc.Server

	logger log.Logger

	conf config.Config

	unaryServerInterceptors []grpc.UnaryServerInterceptor

	opts []grpc.ServerOption

	target string

	serviceName string

	tracer opentracing2.Tracer
}

type ServerOption func(c *GrpcServer)

type ServerOptions struct{}

func NewGrpcServer(target string, options ...ServerOption) *GrpcServer {

	grpcServer := &GrpcServer{}

	grpcServer.target = target

	for _, option := range options {
		option(grpcServer)
	}

	if grpcServer.logger == nil {
		grpcServer.logger = log.NewLogger()
	}

	if grpcServer.conf == nil {
		grpcServer.conf = config.NewNullConfig()
	}

	if grpcServer.tracer == nil {
		grpcServer.tracer = opentracing2.NoopTracer{}
	}

	keepAliveServer := keepalive.ServerParameters{}
	grpc_server_kp_time := grpcServer.conf.GetInt("grpc_server_kp_time")
	if grpc_server_kp_time == 0 {
		grpc_server_kp_time = 60
	}
	keepAliveServer.Time = time.Duration(grpc_server_kp_time) * time.Second

	grpc_server_kp_time_out := grpcServer.conf.GetInt("grpc_server_kp_time_out")
	if grpc_server_kp_time_out == 0 {
		grpc_server_kp_time_out = 5
	}
	keepAliveServer.Timeout = time.Duration(grpc_server_kp_time_out) * time.Second

	//测试没生效
	grpc_server_conn_time_out := grpcServer.conf.GetInt("grpc_server_conn_time_out")
	if grpc_server_conn_time_out == 0 {
		grpc_server_conn_time_out = 3
	}

	baseOpts := []grpc.ServerOption{
		grpc.ConnectionTimeout(time.Duration(grpc_server_conn_time_out) * time.Second),
		grpc.KeepaliveParams(keepAliveServer),
	}

	unaryServerInterceptors := []grpc.UnaryServerInterceptor{}
	if grpcServer.conf.GetBool("grpc_server_tracer") == true {
		unaryServerInterceptors = append(unaryServerInterceptors,
			otgrpc.OpenTracingServerInterceptor(grpcServer.tracer))
	}

	if grpcServer.conf.GetBool("grpc_server_metrics") == true {
		ggp.EnableHandlingTimeHistogram()
		serverMetrics := ggp.DefaultServerMetrics
		serverMetrics.EnableHandlingTimeHistogram(ggp.WithHistogramBuckets(prometheus.DefBuckets))
		unaryServerInterceptors = append(unaryServerInterceptors, serverMetrics.UnaryServerInterceptor())
	}

	if grpcServer.conf.GetBool("grpc_server_check_slow") == true {
		unaryServerInterceptors = append(unaryServerInterceptors, grpcServer.checkServerSlow())
	}

	if grpcServer.conf.GetBool("grpc_server_debug") == true {
		unaryServerInterceptors = append(unaryServerInterceptors, grpcServer.serverDebug())
	}

	//handle panic
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandlerContext(grpcServer.handelPanic()),
	}
	unaryServerInterceptors = append(unaryServerInterceptors, grpc_recovery.UnaryServerInterceptor(opts...))

	if len(grpcServer.unaryServerInterceptors) > 0 {
		unaryServerInterceptors = append(unaryServerInterceptors, grpcServer.unaryServerInterceptors...)
	}

	if len(unaryServerInterceptors) > 0 {
		ui := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryServerInterceptors...))
		baseOpts = append(baseOpts, ui)
	}

	if len(grpcServer.opts) > 0 {
		baseOpts = append(baseOpts, grpcServer.opts...)
	}

	s := grpc.NewServer(baseOpts...)

	grpcServer.Server = s

	return grpcServer
}

func (ServerOptions) WithServerConf(conf config.Config) ServerOption {
	return func(g *GrpcServer) {
		g.conf = conf
	}
}

func (ServerOptions) WithServerLogger(logger log.Logger) ServerOption {
	return func(g *GrpcServer) {
		g.logger = logger
	}
}

func (ServerOptions) WithUnarySrvItcp(options ...grpc.UnaryServerInterceptor) ServerOption {
	return func(g *GrpcServer) {
		g.unaryServerInterceptors = options
	}
}

func (ServerOptions) WithGrpcServerOption(options ...grpc.ServerOption) ServerOption {
	return func(g *GrpcServer) {
		g.opts = options
	}
}

func (ServerOptions) WithTracer(tracer opentracing2.Tracer) ServerOption {
	return func(g *GrpcServer) {
		g.tracer = tracer
	}
}

func (this *GrpcServer) checkServerSlow() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		begin_time := time.Now()
		resp, err = handler(ctx, req)
		end_time := time.Now()

		grpc_client_slow_time := this.conf.GetInt64("grpc_server_slow_time")
		if grpc_client_slow_time != 0 {
			if end_time.Sub(begin_time) > time.Duration(grpc_client_slow_time)*time.Millisecond {
				this.logger.Warnc(ctx, "slow server %s", info.FullMethod)
			}
		}

		return resp, err
	}
}

func (this *GrpcServer) serverDebug() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		begin_time := time.Now()
		this.logger.Debugc(ctx, "grpc server start %s, req : %s", info.FullMethod, spew.Sdump(req))

		resp, err = handler(ctx, req)

		end_time := time.Now()
		this.logger.Debugc(ctx, "grpc server end [%v] %s, resp : %s", end_time.Sub(begin_time).String(),
			info.FullMethod, spew.Sdump(resp))

		return resp, err
	}
}

func (this *GrpcServer) handelPanic() grpc_recovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) (err error) {
		this.logger.Errorc(ctx, spew.Sdump(p))
		return errors.New(spew.Sdump("server panic : ", p))
	}
}

func nilResp() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		return nil, err
	}
}

func panicResp() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		if req.(*helloworld.HelloRequest).Name == "call_panic" {
			panic("is a test")
		} else if req.(*helloworld.HelloRequest).Name == "call_panic_arr" {
			var arr [1]string
			arr[0] = "is a test"
			panic(arr)
		}
		resp, err = handler(ctx, req)

		return resp, err
	}
}

func ServerStubs(stubsFunc func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error)) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		return stubsFunc(ctx, req, info, handler)
	}
}

func (this *GrpcServer) Start() {

	lis, err := net.Listen("tcp", this.target)
	if err != nil {
		this.logger.Panicf("failed to listen: %s", err.Error())
	}

	// Register reflection service on gRPC server.
	reflection.Register(this.Server)

	this.logger.Infof("grpc server starting %s:%s",
		this.serviceName, this.target)
	go func() {
		if err := this.Server.Serve(lis); err != nil {
			this.logger.Panicf("failed to server: %s", err.Error())
		}
	}()
}

func (this *GrpcServer) GracefulShutDown() {
	this.Server.GracefulStop()
}
