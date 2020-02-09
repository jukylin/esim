package grpc

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	ggp "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GrpcClient struct {
	conn *grpc.ClientConn

	cancel context.CancelFunc

	logger log.Logger

	conf config.Config

	clientOpts *ClientOptions
}

type ClientOptions struct {
	cancel context.CancelFunc

	clientMetrics *ggp.ClientMetrics

	tracer opentracing2.Tracer

	logger log.Logger

	conf config.Config

	opts []grpc.DialOption
}


type ClientOptional func(c *ClientOptions)

type ClientOptionals struct{}

func NewClientOptions(options ...ClientOptional) *ClientOptions {
	clientOptions := &ClientOptions{}

	for _, option := range options {
		option(clientOptions)
	}

	if clientOptions.logger == nil {
		clientOptions.logger = log.NewLogger()
	}

	if clientOptions.conf == nil {
		clientOptions.conf = config.NewNullConfig()
	}

	if clientOptions.tracer == nil {
		clientOptions.tracer = opentracing.NewTracer("grpc_client", clientOptions.logger)
	}

	if clientOptions.clientMetrics == nil {
		ggp.EnableClientHandlingTimeHistogram(ggp.WithHistogramBuckets(prometheus.DefBuckets))
		clientOptions.clientMetrics = ggp.DefaultClientMetrics
	}

	keepAliveClient := keepalive.ClientParameters{}
	grpc_client_kp_time := clientOptions.conf.GetInt("grpc_client_kp_time")
	if grpc_client_kp_time == 0 {
		grpc_client_kp_time = 60
	}
	keepAliveClient.Time = time.Duration(grpc_client_kp_time) * time.Second

	grpc_client_kp_time_out := clientOptions.conf.GetInt("grpc_client_kp_time_out")
	if grpc_client_kp_time_out == 0 {
		grpc_client_kp_time_out = 5
	}
	keepAliveClient.Timeout = time.Duration(grpc_client_kp_time_out) * time.Second

	grpc_client_permit_without_stream := clientOptions.conf.GetBool("grpc_client_permit_without_stream")
	keepAliveClient.PermitWithoutStream = grpc_client_permit_without_stream

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepAliveClient),
	}

	if clientOptions.conf.GetBool("grpc_client_tracer") == true {
		opts = append(opts, grpc.WithChainUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(clientOptions.tracer)))
	}

	if clientOptions.conf.GetBool("grpc_client_metrics") == true {
		opts = append(opts, grpc.WithChainUnaryInterceptor(
			clientOptions.clientMetrics.UnaryClientInterceptor()))
	}

	if clientOptions.conf.GetBool("grpc_client_check_slow") == true {
		opts = append(opts, grpc.WithChainUnaryInterceptor(clientOptions.checkClientSlow()))
	}

	if clientOptions.conf.GetBool("grpc_client_debug") == true {
		opts = append(opts, grpc.WithChainUnaryInterceptor(clientOptions.clientDebug()))
	}

	clientOptions.opts = append(opts, clientOptions.opts...)

	return clientOptions
}

func (ClientOptionals) WithConf(conf config.Config) ClientOptional {
	return func(g *ClientOptions) {
		g.conf = conf
	}
}

func (ClientOptionals) WithLogger(logger log.Logger) ClientOptional {
	return func(g *ClientOptions) {
		g.logger = logger
	}
}

func (ClientOptionals) WithTracer(tracer opentracing2.Tracer) ClientOptional {
	return func(g *ClientOptions) {
		g.tracer = tracer
	}
}

func (ClientOptionals) WithMetrics(metrics *ggp.ClientMetrics) ClientOptional {
	return func(g *ClientOptions) {
		g.clientMetrics = metrics
	}
}

func (ClientOptionals) WithDialOptions(options ...grpc.DialOption) ClientOptional {
	return func(g *ClientOptions) {
		g.opts = options
	}
}


//NewGrpcClient create GrpcClient for business.
func NewClient(clientOptions *ClientOptions) *GrpcClient {

	grpcClient := &GrpcClient{}
	if clientOptions == nil {
		grpcClient.clientOpts = &ClientOptions{}
	}else{
		grpcClient.clientOpts = clientOptions
	}

	return grpcClient
}


func (this *GrpcClient) DialContext(ctx context.Context, target string) *grpc.ClientConn {

	var cancel context.CancelFunc

	grpc_client_conn_time_out := this.clientOpts.conf.GetInt("grpc_client_conn_time_out")
	if grpc_client_conn_time_out == 0 {
		grpc_client_conn_time_out = 3
		ctx, cancel = context.WithTimeout(ctx, time.Duration(grpc_client_conn_time_out)*time.Second)
		this.cancel = cancel
	}

	conn, err := grpc.DialContext(ctx, target, this.clientOpts.opts...)
	if err != nil {
		this.logger.Errorf(err.Error())
		return nil
	}
	this.conn = conn

	return conn
}

func (this *GrpcClient) Close() {
	this.conn.Close()
	this.cancel()
}

func (this *ClientOptions) checkClientSlow() func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		grpc_client_slow_time := this.conf.GetInt64("grpc_client_slow_time")

		begin_time := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		end_time := time.Now()

		if grpc_client_slow_time != 0 {
			if end_time.Sub(begin_time) > time.Duration(grpc_client_slow_time)*time.Millisecond {
				this.logger.Warnc(ctx, "slow client grpc_handle %s", method)
			}
		}

		return err
	}
}

func (this *ClientOptions) clientDebug() func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		begin_time := time.Now()
		this.logger.Debugc(ctx, "grpc client start %s : %s, req : %s",
			cc.Target(), method, spew.Sdump(req))

		err := invoker(ctx, method, req, reply, cc, opts...)

		end_time := time.Now()
		this.logger.Debugc(ctx, "grpc client end [%v] %s : %s, reply : %s",
			end_time.Sub(begin_time).String(), cc.Target(), method, spew.Sdump(reply))

		return err
	}
}


func ClientStubs(stubsFunc func(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error) func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			err := stubsFunc(ctx, method, req, reply, cc, invoker, opts...)
		return err
	}
}


func slowRequest(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	time.Sleep(20 * time.Millisecond)
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}


func timeoutRequest(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	time.Sleep(10 * time.Second)
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}
