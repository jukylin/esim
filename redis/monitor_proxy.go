package redis

import (
	"context"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
)

type MonitorProxy struct {
	name string

	nextConn ContextConn

	tracer opentracing2.Tracer

	conf config.Config

	logger log.Logger

	afterEvents []afterEvents
}

type afterEvents func(context.Context, *execInfo)

type MonitorProxyOption func(c *MonitorProxy)

type MonitorProxyOptions struct{}

func NewMonitorProxy(options ...MonitorProxyOption) *MonitorProxy {
	monitorProxy := &MonitorProxy{}
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

	monitorProxy.registerAfterEvent()

	monitorProxy.name = "monitor_proxy"

	return monitorProxy
}

func (MonitorProxyOptions) WithConf(conf config.Config) MonitorProxyOption {
	return func(r *MonitorProxy) {
		r.conf = conf
	}
}

func (MonitorProxyOptions) WithLogger(logger log.Logger) MonitorProxyOption {
	return func(r *MonitorProxy) {
		r.logger = logger
	}
}

func (MonitorProxyOptions) WithTracer(tracer opentracing2.Tracer) MonitorProxyOption {
	return func(r *MonitorProxy) {
		r.tracer = tracer
	}
}

// implement Proxy interface
func (mp *MonitorProxy) NextProxy(conn interface{}) {
	mp.nextConn = conn.(ContextConn)
}

// implement Proxy interface
func (mp *MonitorProxy) ProxyName() string {
	return mp.name
}

func (mp *MonitorProxy) Close() error {
	err := mp.nextConn.Close()

	return err
}

func (mp *MonitorProxy) Err() (err error) {
	err = mp.nextConn.Err()
	return
}

func (mp *MonitorProxy) Do(ctx context.Context, commandName string,
	args ...interface{}) (reply interface{}, err error) {
	now := time.Now()

	reply, err = mp.nextConn.Do(ctx, commandName, args...)

	execInfo := &execInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = commandName
	execInfo.args = args

	mp.after(ctx, execInfo)

	return
}

func (mp *MonitorProxy) Send(ctx context.Context, commandName string,
	args ...interface{}) (err error) {
	now := time.Now()
	err = mp.nextConn.Send(ctx, commandName, args...)

	execInfo := &execInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = commandName
	execInfo.args = args

	mp.after(ctx, execInfo)

	return
}

func (mp *MonitorProxy) Flush(ctx context.Context) (err error) {
	now := time.Now()
	err = mp.nextConn.Flush(ctx)

	info := execInfo{}
	info.err = err
	info.startTime = now
	info.endTime = time.Now()
	info.commandName = "flush"

	mp.after(ctx, &info)

	return
}

func (mp *MonitorProxy) Receive(ctx context.Context) (reply interface{}, err error) {
	now := time.Now()
	reply, err = mp.nextConn.Receive(ctx)

	execInfo := &execInfo{}
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = "receive"

	mp.after(ctx, execInfo)

	return
}

// 初始化回调事件
func (mp *MonitorProxy) registerAfterEvent() {
	if mp.conf.GetBool("redis_tracer") {
		mp.afterEvents = append(mp.afterEvents, mp.redisTracer)
	}

	if mp.conf.GetBool("redis_check_slow") {
		mp.afterEvents = append(mp.afterEvents, mp.redisSlowCommand)
	}

	if mp.conf.GetBool("redis_metrics") {
		mp.afterEvents = append(mp.afterEvents, mp.redisMetrics)
	}
}

func (mp *MonitorProxy) after(ctx context.Context, execInfo *execInfo) {
	for _, e := range mp.afterEvents {
		e(ctx, execInfo)
	}
}

func (mp *MonitorProxy) redisTracer(ctx context.Context, info *execInfo) {
	span := opentracing.GetSpan(ctx, mp.tracer, info.commandName, info.startTime)
	if info.err != nil {
		span.SetTag("error", true)
		span.LogKV("error_detailed", info.err.Error())
	}
	span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: info.endTime})
}

// 慢命令
func (mp *MonitorProxy) redisSlowCommand(ctx context.Context, info *execInfo) {
	redisSlowTime := mp.conf.GetInt64("redis_slow_time")
	if info.endTime.Sub(info.startTime) > time.Duration(redisSlowTime)*time.Millisecond {
		mp.logger.Warnf("Slow redis command %s", info.commandName)
	}
}

func (mp *MonitorProxy) redisMetrics(ctx context.Context, info *execInfo) {
	redisTotalLab := prometheus.Labels{"cmd": info.commandName}
	redisDurationLab := prometheus.Labels{"cmd": info.commandName}

	redisTotal.With(redisTotalLab).Inc()
	redisDuration.With(redisDurationLab).Observe(info.endTime.Sub(info.startTime).Seconds())
}
