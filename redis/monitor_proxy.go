package redis

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
)

const sendMethod = "Send"
const receiveMethod = "Receive"

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

// Implement Proxy interface.
func (mp *MonitorProxy) NextProxy(conn interface{}) {
	mp.nextConn = conn.(ContextConn)
}

// Implement Proxy interface.
func (mp *MonitorProxy) ProxyName() string {
	return mp.name
}

func (mp *MonitorProxy) Close() error {
	now := time.Now()

	err := mp.nextConn.Close()

	execInfo := newExecInfo()
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = ""
	execInfo.method = "Close"
	execInfo.args = nil
	execInfo.reply = nil

	mp.after(context.Background(), execInfo)
	execInfo.Release()

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

	execInfo := newExecInfo()
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.commandName = commandName
	execInfo.method = "Do"
	execInfo.args = args
	execInfo.reply = reply

	mp.after(ctx, execInfo)
	execInfo.Release()

	return
}

func (mp *MonitorProxy) Send(ctx context.Context, commandName string,
	args ...interface{}) (err error) {
	now := time.Now()
	err = mp.nextConn.Send(ctx, commandName, args...)

	execInfo := newExecInfo()
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.method = sendMethod
	execInfo.commandName = commandName
	execInfo.args = args
	execInfo.reply = nil

	mp.after(ctx, execInfo)
	execInfo.Release()

	return
}

func (mp *MonitorProxy) Flush(ctx context.Context) (err error) {
	now := time.Now()
	err = mp.nextConn.Flush(ctx)

	execInfo := newExecInfo()
	execInfo.err = err
	execInfo.startTime = now
	execInfo.method = "Flush"
	execInfo.endTime = time.Now()
	execInfo.commandName = ""
	execInfo.reply = nil

	mp.after(ctx, execInfo)
	execInfo.Release()

	return
}

func (mp *MonitorProxy) Receive(ctx context.Context) (reply interface{}, err error) {
	now := time.Now()
	reply, err = mp.nextConn.Receive(ctx)

	execInfo := newExecInfo()
	execInfo.err = err
	execInfo.startTime = now
	execInfo.endTime = time.Now()
	execInfo.method = receiveMethod
	execInfo.commandName = ""
	execInfo.reply = reply

	mp.after(ctx, execInfo)
	execInfo.Release()

	return
}

func (mp *MonitorProxy) registerAfterEvent() {
	if mp.conf.GetBool("redis_trace") {
		mp.afterEvents = append(mp.afterEvents, mp.redisTracer)
	}

	if mp.conf.GetBool("redis_check_slow") {
		mp.afterEvents = append(mp.afterEvents, mp.redisSlowCommand)
	}

	if mp.conf.GetBool("redis_metrics") {
		mp.afterEvents = append(mp.afterEvents, mp.redisMetrics)
	}

	if mp.conf.GetBool("debug") {
		mp.afterEvents = append(mp.afterEvents, mp.redisDebug)
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

	if span != nil {
		span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: info.endTime})
	}
}

func (mp *MonitorProxy) redisSlowCommand(ctx context.Context, info *execInfo) {
	redisSlowTime := mp.conf.GetInt64("redis_slow_time")
	if info.endTime.Sub(info.startTime) > time.Duration(redisSlowTime)*time.Millisecond {
		mp.logger.Warnf("Slow redis command %s [%s]",
			info.commandName, info.endTime.Sub(info.startTime).String())
	}
}

func (mp *MonitorProxy) redisMetrics(ctx context.Context, info *execInfo) {
	redisTotalLab := prometheus.Labels{"cmd": info.commandName}
	redisDurationLab := prometheus.Labels{"cmd": info.commandName}

	redisTotal.With(redisTotalLab).Inc()
	redisDuration.With(redisDurationLab).Observe(info.endTime.Sub(info.startTime).Seconds())
}

func (mp *MonitorProxy) redisDebug(ctx context.Context, info *execInfo) {
	mp.print(ctx, info.method, info.commandName, info.args, info.reply, info.err)
}

func (mp *MonitorProxy) printValue(buf *bytes.Buffer, v interface{}) {
	const chop = 32
	switch v := v.(type) {
	case []byte:
		if len(v) > chop {
			fmt.Fprintf(buf, "%q...", v[:chop])
		} else {
			fmt.Fprintf(buf, "%q", v)
		}
	case string:
		if len(v) > chop {
			fmt.Fprintf(buf, "%q...", v[:chop])
		} else {
			fmt.Fprintf(buf, "%q", v)
		}
	case []interface{}:
		if len(v) == 0 {
			buf.WriteString("[]")
		} else {
			sep := "["
			fin := "]"
			if len(v) > chop {
				v = v[:chop]
				fin = "...]"
			}
			for _, vv := range v {
				buf.WriteString(sep)
				mp.printValue(buf, vv)
				sep = ", "
			}
			buf.WriteString(fin)
		}
	default:
		fmt.Fprint(buf, v)
	}
}

func (mp *MonitorProxy) print(ctx context.Context, method, commandName string,
	args []interface{}, reply interface{}, err error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s(", method)
	if method != receiveMethod {
		buf.WriteString(commandName)
		for _, arg := range args {
			buf.WriteString(", ")
			mp.printValue(&buf, arg)
		}
	}
	buf.WriteString(") -> (")
	if method != sendMethod {
		mp.printValue(&buf, reply)
		buf.WriteString(", ")
	}
	fmt.Fprintf(&buf, "%v)", err)
	mp.logger.Debugc(ctx, buf.String())
}
