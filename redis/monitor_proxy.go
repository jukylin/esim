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

	logger log.Logger

	afterEvents []afterEvents
}

type afterEvents func(context.Context, ExecRedisInfo)

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

	if monitorProxy.logger == nil {
		monitorProxy.logger = log.NewLogger()
	}

	if monitorProxy.tracer == nil {
		monitorProxy.tracer = opentracing.NewTracer("redis", monitorProxy.logger)
	}

	monitorProxy.name = "monitor_proxy"

	return monitorProxy
}

func (MonitorProxyOptions) WithConf(conf config.Config) MonitorProxyOption {
	return func(r *monitorProxy) {
		r.conf = conf
	}
}

func (MonitorProxyOptions) WithLogger(logger log.Logger) MonitorProxyOption {
	return func(r *monitorProxy) {
		r.logger = logger
	}
}

func (MonitorProxyOptions) WithTracer(tracer opentracing2.Tracer) MonitorProxyOption {
	return func(r *monitorProxy) {
		r.tracer = tracer
	}
}

//implement Proxy interface
func (mp *monitorProxy) NextProxy(conn interface{}) {
	mp.nextConn = conn.(ContextConn)
}

//implement Proxy interface
func (mp *monitorProxy) ProxyName() string {
	return mp.name
}

func (mp *monitorProxy) Close() error {
	err := mp.nextConn.Close()

	return err
}

func (mp *monitorProxy) Err() (err error) {
	err = mp.nextConn.Err()
	return
}

func (mp *monitorProxy) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	now := time.Now()

	reply, err = mp.nextConn.Do(ctx, commandName, args...)

	execInfo := ExecRedisInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = commandName
	execInfo.args = args

	mp.after(ctx, execInfo)

	return
}

func (mp *monitorProxy) Send(ctx context.Context, commandName string, args ...interface{}) (err error) {

	now := time.Now()
	err = mp.nextConn.Send(ctx, commandName, args...)

	execInfo := ExecRedisInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = commandName
	execInfo.args = args

	mp.after(ctx, execInfo)

	return
}

func (mp *monitorProxy) Flush(ctx context.Context) (err error) {

	now := time.Now()
	err = mp.nextConn.Flush(ctx)

	execInfo := ExecRedisInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = "flush"

	mp.after(ctx, execInfo)

	return
}

func (mp *monitorProxy) Receive(ctx context.Context) (reply interface{}, err error) {

	now := time.Now()
	reply, err = mp.nextConn.Receive(ctx)

	execInfo := ExecRedisInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = "receive"

	mp.after(ctx, execInfo)

	return
}

//初始化回调事件
func (mp *monitorProxy) registerAfterEvent() {
	if mp.conf.GetBool("redis_tracer") == true {
		mp.afterEvents = append(mp.afterEvents, mp.redisTracer)
	}

	if mp.conf.GetBool("redis_check_slow") == true {
		mp.afterEvents = append(mp.afterEvents, mp.redisSlowCommand)
	}

	if mp.conf.GetBool("redis_metrics") == true {
		mp.afterEvents = append(mp.afterEvents, mp.redisMetrics)
	}
}

func (mp *monitorProxy) after(ctx context.Context, execInfo ExecRedisInfo) {
	for _, e := range mp.afterEvents {
		e(ctx, execInfo)
	}
}

func (mp *monitorProxy) redisTracer(ctx context.Context, info ExecRedisInfo) {
	span := opentracing.GetSpan(ctx, mp.tracer, info.commandName, info.startTime)
	if info.err != nil {
		span.SetTag("error", true)
		span.LogKV("error_detailed", info.err.Error())
	}
	span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: info.endTime})
}

//慢命令
func (mp *monitorProxy) redisSlowCommand(ctx context.Context, info ExecRedisInfo) {
	redisSlowTime := mp.conf.GetInt64("redis_slow_time")
	if info.endTime.Sub(info.startTime) > time.Duration(redisSlowTime)*time.Millisecond {
		mp.logger.Warnf("slow redis command %s", info.commandName)
	}
}

func (mp *monitorProxy) redisMetrics(ctx context.Context, info ExecRedisInfo) {
	redisTotalLab := prometheus.Labels{"cmd": info.commandName}
	redisDurationLab := prometheus.Labels{"cmd": info.commandName}

	redisTotal.With(redisTotalLab).Inc()
	redisDuration.With(redisDurationLab).Observe(info.endTime.Sub(info.startTime).Seconds())
}
