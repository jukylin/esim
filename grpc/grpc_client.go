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

	clientMetrics *ggp.ClientMetrics

	tracer opentracing2.Tracer

	log log.Logger

	conf config.Config

	opts []grpc.DialOption
}

type ClientOption func(c *GrpcClient)

type ClientOptions struct{}

//NewGrpcClient create GrpcClient for business.
func NewGrpcClient(options ...ClientOption) *GrpcClient {

	grpcClient := &GrpcClient{}

	for _, option := range options {
		option(grpcClient)
	}

	if grpcClient.log == nil {
		grpcClient.log = log.NewLogger()
	}

	if grpcClient.conf == nil {
		grpcClient.conf = config.NewNullConfig()
	}

	if grpcClient.tracer == nil {
		grpcClient.tracer = opentracing.NewTracer("grpc_client", grpcClient.log)
	}

	if grpcClient.clientMetrics == nil {
		ggp.EnableClientHandlingTimeHistogram(ggp.WithHistogramBuckets(prometheus.DefBuckets))
		grpcClient.clientMetrics = ggp.DefaultClientMetrics
	}

	return grpcClient
}

func (ClientOptions) WithClientConf(conf config.Config) ClientOption {
	return func(g *GrpcClient) {
		g.conf = conf
	}
}

func (ClientOptions) WithClientLogger(log log.Logger) ClientOption {
	return func(g *GrpcClient) {
		g.log = log
	}
}

func (ClientOptions) WithClientTracer(tracer opentracing2.Tracer) ClientOption {
	return func(g *GrpcClient) {
		g.tracer = tracer
	}
}

func (ClientOptions) WithClientMetrics(metrics *ggp.ClientMetrics) ClientOption {
	return func(g *GrpcClient) {
		g.clientMetrics = metrics
	}
}

func (ClientOptions) WithClientDialOptions(options ...grpc.DialOption) ClientOption {
	return func(g *GrpcClient) {
		g.opts = options
	}
}

func (this *GrpcClient) DialContext(ctx context.Context, target string) *grpc.ClientConn {

	keepAliveClient := keepalive.ClientParameters{}
	grpc_client_kp_time := this.conf.GetInt("grpc_client_kp_time")
	if grpc_client_kp_time == 0 {
		grpc_client_kp_time = 60
	}
	keepAliveClient.Time = time.Duration(grpc_client_kp_time) * time.Second

	grpc_client_kp_time_out := this.conf.GetInt("grpc_client_kp_time_out")
	if grpc_client_kp_time_out == 0 {
		grpc_client_kp_time_out = 5
	}
	keepAliveClient.Timeout = time.Duration(grpc_client_kp_time_out) * time.Second

	grpc_client_permit_without_stream := this.conf.GetBool("grpc_client_permit_without_stream")
	keepAliveClient.PermitWithoutStream = grpc_client_permit_without_stream

	baseOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepAliveClient),
	}

	if this.conf.GetBool("grpc_client_tracer") == true {
		baseOpts = append(baseOpts, grpc.WithChainUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(this.tracer)))
	}

	if this.conf.GetBool("grpc_client_metrics") == true {
		baseOpts = append(baseOpts, grpc.WithChainUnaryInterceptor(
			this.clientMetrics.UnaryClientInterceptor()))
	}

	if this.conf.GetBool("grpc_client_check_slow") == true {
		baseOpts = append(baseOpts, grpc.WithChainUnaryInterceptor(this.checkClientSlow()))
	}

	if this.conf.GetBool("grpc_client_debug") == true {
		baseOpts = append(baseOpts, grpc.WithChainUnaryInterceptor(this.clientDebug()))
	}

	var cancel context.CancelFunc

	grpc_client_conn_time_out := this.conf.GetInt("grpc_client_conn_time_out")
	if grpc_client_conn_time_out == 0 {
		grpc_client_conn_time_out = 3
		ctx, cancel = context.WithTimeout(ctx, time.Duration(grpc_client_conn_time_out)*time.Second)
		this.cancel = cancel
	}

	baseOpts = append(baseOpts, this.opts...)

	conn, err := grpc.DialContext(ctx, target, baseOpts...)
	if err != nil {
		this.log.Errorf(err.Error())
		return nil
	}
	this.conn = conn

	return conn
}

func (this *GrpcClient) Close() {
	this.conn.Close()
	this.cancel()
}

func (this *GrpcClient) checkClientSlow() func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		grpc_client_slow_time := this.conf.GetInt64("grpc_client_slow_time")

		begin_time := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		end_time := time.Now()

		if grpc_client_slow_time != 0 {
			if end_time.Sub(begin_time) > time.Duration(grpc_client_slow_time)*time.Millisecond {
				this.log.Warnc(ctx, "slow client grpc_handle %s", method)
			}
		}

		return err
	}
}

func (this *GrpcClient) clientDebug() func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		begin_time := time.Now()
		this.log.Debugc(ctx, "grpc client start %s : %s, req : %s",
			cc.Target(), method, spew.Sdump(req))

		err := invoker(ctx, method, req, reply, cc, opts...)

		end_time := time.Now()
		this.log.Debugc(ctx, "grpc client end [%v] %s : %s, reply : %s",
			end_time.Sub(begin_time).String(), cc.Target(), method, spew.Sdump(reply))

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
