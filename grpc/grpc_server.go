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
	grpcServerKpTime := grpcServer.conf.GetInt("grpc_server_kp_time")
	if grpcServerKpTime == 0 {
		grpcServerKpTime = 60
	}
	keepAliveServer.Time = time.Duration(grpcServerKpTime) * time.Second

	grpcServerKpTimeOut := grpcServer.conf.GetInt("grpc_server_kp_time_out")
	if grpcServerKpTimeOut == 0 {
		grpcServerKpTimeOut = 5
	}
	keepAliveServer.Timeout = time.Duration(grpcServerKpTimeOut) * time.Second

	//测试没生效
	grpcServerConnTimeOut := grpcServer.conf.GetInt("grpc_server_conn_time_out")
	if grpcServerConnTimeOut == 0 {
		grpcServerConnTimeOut = 3
	}

	baseOpts := []grpc.ServerOption{
		grpc.ConnectionTimeout(time.Duration(grpcServerConnTimeOut) * time.Second),
		grpc.KeepaliveParams(keepAliveServer),
	}

	unaryServerInterceptors := make([]grpc.UnaryServerInterceptor, 0)
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

func (gs *GrpcServer) checkServerSlow() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		beginTime := time.Now()
		resp, err = handler(ctx, req)
		endTime := time.Now()

		grpcClientSlowTime := gs.conf.GetInt64("grpc_server_slow_time")
		if grpcClientSlowTime != 0 {
			if endTime.Sub(beginTime) > time.Duration(grpcClientSlowTime)*time.Millisecond {
				gs.logger.Warnc(ctx, "slow server %s", info.FullMethod)
			}
		}

		return resp, err
	}
}

func (gs *GrpcServer) serverDebug() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		beginTime := time.Now()
		gs.logger.Debugc(ctx, "grpc server start %s, req : %s", info.FullMethod, spew.Sdump(req))

		resp, err = handler(ctx, req)

		endTime := time.Now()
		gs.logger.Debugc(ctx, "grpc server end [%v] %s, resp : %s", endTime.Sub(beginTime).String(),
			info.FullMethod, spew.Sdump(resp))

		return resp, err
	}
}

func (gs *GrpcServer) handelPanic() grpc_recovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) (err error) {
		gs.logger.Errorc(ctx, spew.Sdump(p))
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

func (gs *GrpcServer) Start() {

	lis, err := net.Listen("tcp", gs.target)
	if err != nil {
		gs.logger.Panicf("failed to listen: %s", err.Error())
	}

	// Register reflection service on gRPC server.
	reflection.Register(gs.Server)

	gs.logger.Infof("grpc server starting %s:%s",
		gs.serviceName, gs.target)
	go func() {
		if err := gs.Server.Serve(lis); err != nil {
			gs.logger.Panicf("failed to server: %s", err.Error())
		}
	}()
}

func (gs *GrpcServer) GracefulShutDown() {
	gs.Server.GracefulStop()
}
