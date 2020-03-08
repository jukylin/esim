package redis

import (
	"context"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type monitorProxy struct {
	name string

	nextConn ContextConn

	tracer opentracing2.Tracer

	conf config.Config

	log log.Logger

	afterEvents []afterEvents
}

type afterEvents func(context.Context, RedisExecInfo)

type MonitorProxyOption func(c *monitorProxy)

type MonitorProxyOptions struct{}

func NewMonitorProxy(options ...MonitorProxyOption) *monitorProxy {

	monitorProxy := &monitorProxy{}
	for _, option := range options {
		option(monitorProxy)
	}

	if monitorProxy.conf == nil {
		monitorProxy.conf = config.NewNullConfig()
	}

	if monitorProxy.log == nil {
		monitorProxy.log = log.NewLogger()
	}

	if monitorProxy.tracer == nil {
		monitorProxy.tracer = opentracing.NewTracer("redis", monitorProxy.log)
	}

	monitorProxy.name = "monitor_proxy"

	return monitorProxy
}

func (MonitorProxyOptions) WithConf(conf config.Config) MonitorProxyOption {
	return func(r *monitorProxy) {
		r.conf = conf
	}
}

func (MonitorProxyOptions) WithLogger(log log.Logger) MonitorProxyOption {
	return func(r *monitorProxy) {
		r.log = log
	}
}

func (MonitorProxyOptions) WithTracer(tracer opentracing2.Tracer) MonitorProxyOption {
	return func(r *monitorProxy) {
		r.tracer = tracer
	}
}

//implement Proxy interface
func (pc *monitorProxy) NextProxy(conn interface{}) {
	pc.nextConn = conn.(ContextConn)
}

//implement Proxy interface
func (pc *monitorProxy) ProxyName() string {
	return pc.name
}

func (pc *monitorProxy) Close() error {
	err := pc.nextConn.Close()

	return err
}

func (pc *monitorProxy) Err() (err error) {
	err = pc.nextConn.Err()
	return
}

func (pc *monitorProxy) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	now := time.Now()

	reply, err = pc.nextConn.Do(ctx, commandName, args...)

	execInfo := RedisExecInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = commandName
	execInfo.args = args

	pc.after(ctx, execInfo)

	return
}

func (pc *monitorProxy) Send(ctx context.Context, commandName string, args ...interface{}) (err error) {

	now := time.Now()
	err = pc.nextConn.Send(ctx, commandName, args...)

	execInfo := RedisExecInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = commandName
	execInfo.args = args

	pc.after(ctx, execInfo)

	return
}

func (pc *monitorProxy) Flush(ctx context.Context) (err error) {

	now := time.Now()
	err = pc.nextConn.Flush(ctx)

	execInfo := RedisExecInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = "flush"

	pc.after(ctx, execInfo)

	return
}

func (pc *monitorProxy) Receive(ctx context.Context) (reply interface{}, err error) {

	now := time.Now()
	reply, err = pc.nextConn.Receive(ctx)

	execInfo := RedisExecInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = "receive"

	pc.after(ctx, execInfo)

	return
}

//初始化回调事件
func (pc *monitorProxy) registerAfterEvent() {
	if pc.conf.GetBool("redis_tracer") == true {
		pc.afterEvents = append(pc.afterEvents, pc.redisTracer)
	}

	if pc.conf.GetBool("redis_check_slow") == true {
		pc.afterEvents = append(pc.afterEvents, pc.redisSlowCommand)
	}

	if pc.conf.GetBool("redis_metrics") == true {
		pc.afterEvents = append(pc.afterEvents, pc.redisMetrics)
	}
}

func (pc *monitorProxy) after(ctx context.Context, execInfo RedisExecInfo) {
	for _, e := range pc.afterEvents {
		e(ctx, execInfo)
	}
}

func (pc *monitorProxy) redisTracer(ctx context.Context, info RedisExecInfo) {
	span := opentracing.GetSpan(ctx, pc.tracer, info.commandName, info.startTime)
	if info.err != nil {
		span.SetTag("error", true)
		span.LogKV("error_detailed", info.err.Error())
	}
	span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: info.endTime})
}

//慢命令
func (pc *monitorProxy) redisSlowCommand(ctx context.Context, info RedisExecInfo) {
	redis_slow_time := pc.conf.GetInt64("redis_slow_time")
	if info.endTime.Sub(info.startTime) > time.Duration(redis_slow_time)*time.Millisecond {
		pc.log.Warnf("slow redis command %s", info.commandName)
	}
}

func (pc *monitorProxy) redisMetrics(ctx context.Context, info RedisExecInfo) {
	redisTotalLab := prometheus.Labels{"cmd": info.commandName}
	redisDurationLab := prometheus.Labels{"cmd": info.commandName}

	redisTotal.With(redisTotalLab).Inc()
	redisDuration.With(redisDurationLab).Observe(info.endTime.Sub(info.startTime).Seconds())
}
