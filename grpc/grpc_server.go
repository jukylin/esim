package grpc

import (
	"errors"
	"net"
	"time"

	"github.com/davecgh/go-spew/spew"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	ggp "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/tracer-id"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	address = "0.0.0.0"

	port = 50051

	isTest = "is test"

	callPanic = "call_panic"

	callNil = "call_nil"

	callPanicArr = "callPanciArr"

	esim = "esim"
)

type Server struct {
	Server *grpc.Server

	logger log.Logger

	conf config.Config

	unaryServerInterceptors []grpc.UnaryServerInterceptor

	opts []grpc.ServerOption

	target string

	serviceName string

	tracer opentracing2.Tracer
}

type ServerOption func(c *Server)

type ServerOptions struct{}

func NewServer(target string, options ...ServerOption) *Server {
	Server := &Server{}

	Server.target = target

	for _, option := range options {
		option(Server)
	}

	if Server.logger == nil {
		Server.logger = log.NewLogger()
	}

	if Server.conf == nil {
		Server.conf = config.NewNullConfig()
	}

	if Server.tracer == nil {
		Server.tracer = opentracing2.NoopTracer{}
	}

	keepAliveServer := keepalive.ServerParameters{}
	ServerKpTime := Server.conf.GetInt("grpc_server_kp_time")
	if ServerKpTime == 0 {
		ServerKpTime = 60
	}
	keepAliveServer.Time = time.Duration(ServerKpTime) * time.Second

	ServerKpTimeOut := Server.conf.GetInt("grpc_server_kp_time_out")
	if ServerKpTimeOut == 0 {
		ServerKpTimeOut = 5
	}
	keepAliveServer.Timeout = time.Duration(ServerKpTimeOut) * time.Second

	ServerConnTimeOut := Server.conf.GetInt("grpc_server_conn_time_out")
	if ServerConnTimeOut == 0 {
		ServerConnTimeOut = 3
	}

	baseOpts := []grpc.ServerOption{
		grpc.ConnectionTimeout(time.Duration(ServerConnTimeOut) * time.Second),
		grpc.KeepaliveParams(keepAliveServer),
	}

	unaryServerInterceptors := make([]grpc.UnaryServerInterceptor, 0)
	if Server.conf.GetBool("grpc_server_tracer") {
		unaryServerInterceptors = append(unaryServerInterceptors,
			otgrpc.OpenTracingServerInterceptor(Server.tracer))
	}

	if Server.conf.GetBool("grpc_server_metrics") {
		ggp.EnableHandlingTimeHistogram()
		serverMetrics := ggp.DefaultServerMetrics
		serverMetrics.EnableHandlingTimeHistogram(ggp.WithHistogramBuckets(prometheus.DefBuckets))
		unaryServerInterceptors = append(unaryServerInterceptors,
			serverMetrics.UnaryServerInterceptor())
	}

	if Server.conf.GetBool("grpc_server_check_slow") {
		unaryServerInterceptors = append(unaryServerInterceptors, Server.checkServerSlow())
	}

	if Server.conf.GetBool("grpc_server_debug") {
		unaryServerInterceptors = append(unaryServerInterceptors, Server.serverDebug())
	}

	unaryServerInterceptors = append(unaryServerInterceptors, Server.tracerID())

	// handle panic
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandlerContext(Server.handelPanic()),
	}
	unaryServerInterceptors = append(unaryServerInterceptors,
		grpc_recovery.UnaryServerInterceptor(opts...))

	if len(Server.unaryServerInterceptors) > 0 {
		unaryServerInterceptors = append(unaryServerInterceptors,
			Server.unaryServerInterceptors...)
	}

	if len(unaryServerInterceptors) > 0 {
		ui := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryServerInterceptors...))
		baseOpts = append(baseOpts, ui)
	}

	if len(Server.opts) > 0 {
		baseOpts = append(baseOpts, Server.opts...)
	}

	s := grpc.NewServer(baseOpts...)

	Server.Server = s

	return Server
}

func (ServerOptions) WithServerConf(conf config.Config) ServerOption {
	return func(g *Server) {
		g.conf = conf
	}
}

func (ServerOptions) WithServerLogger(logger log.Logger) ServerOption {
	return func(g *Server) {
		g.logger = logger
	}
}

func (ServerOptions) WithUnarySrvItcp(options ...grpc.UnaryServerInterceptor) ServerOption {
	return func(g *Server) {
		g.unaryServerInterceptors = options
	}
}

func (ServerOptions) WithServerOption(options ...grpc.ServerOption) ServerOption {
	return func(g *Server) {
		g.opts = options
	}
}

func (ServerOptions) WithTracer(tracer opentracing2.Tracer) ServerOption {
	return func(g *Server) {
		g.tracer = tracer
	}
}

func (gs *Server) checkServerSlow() grpc.UnaryServerInterceptor {
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
				gs.logger.Warnc(ctx, "Slow server %s", info.FullMethod)
			}
		}

		return resp, err
	}
}

func (gs *Server) serverDebug() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		beginTime := time.Now()
		gs.logger.Debugc(ctx, "Grpc server start %s, req : %s", info.FullMethod, spew.Sdump(req))

		resp, err = handler(ctx, req)

		endTime := time.Now()
		gs.logger.Debugc(ctx, "Grpc server end [%v] %s, resp : %s",
			endTime.Sub(beginTime).String(),
			info.FullMethod, spew.Sdump(resp))

		return resp, err
	}
}

func (gs *Server) handelPanic() grpc_recovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) (err error) {
		gs.logger.Errorc(ctx, spew.Sdump(p))
		return errors.New(spew.Sdump("Server panic : ", p))
	}
}

// tracerId If not found opentracing's tracer_id then generate a new tracer_id
// Recommend to the end of the Interceptor
func (gs *Server) tracerID() grpc.UnaryServerInterceptor {
	tracerID := tracerid.TracerID()
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		sp := opentracing2.SpanFromContext(ctx)
		if sp == nil {
			ctx = context.WithValue(ctx, tracerid.ActiveEsimKey, tracerID())
		}

		resp, err = handler(ctx, req)

		return resp, err
	}
}

//nolint:deadcode,unused
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
		if req.(*helloworld.HelloRequest).Name == callPanic {
			panic(isTest)
		} else if req.(*helloworld.HelloRequest).Name == callPanicArr {
			var arr [1]string
			arr[0] = isTest
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

func (gs *Server) Start() {
	lis, err := net.Listen("tcp", gs.target)
	if err != nil {
		gs.logger.Panicf("Failed to listen: %s", err.Error())
	}

	// Register reflection service on gRPC server.
	reflection.Register(gs.Server)

	gs.logger.Infof("Grpc server starting %s:%s",
		gs.serviceName, gs.target)
	go func() {
		if err := gs.Server.Serve(lis); err != nil {
			gs.logger.Panicf("Failed to start server: %s", err.Error())
		}
	}()
}

func (gs *Server) GracefulShutDown() {
	gs.Server.GracefulStop()
}
