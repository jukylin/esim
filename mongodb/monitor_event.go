package mongodb

import (
	"context"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/event"
)

type MonitorEvent struct {
	nextEvent MgoEvent

	conf config.Config

	logger log.Logger

	tracer opentracing2.Tracer

	afterEvents []afterEvents
}

type afterEvents func(context.Context, *mongoBackEvent, time.Time, time.Time)

type MonitorEventOptions struct{}

type EventOption func(c *MonitorEvent)

func NewMonitorEvent(options ...EventOption) MgoEvent {
	m := &MonitorEvent{}

	for _, option := range options {
		option(m)
	}

	if m.conf == nil {
		m.conf = config.NewNullConfig()
	}

	if m.logger == nil {
		m.logger = log.NewLogger()
	}

	if m.tracer == nil {
		m.tracer = opentracing.NewTracer("mongodb", m.logger)
	}

	m.registerAfterEvent()

	return m
}

func (MonitorEventOptions) WithConf(conf config.Config) EventOption {
	return func(m *MonitorEvent) {
		m.conf = conf
	}
}

func (MonitorEventOptions) WithLogger(logger log.Logger) EventOption {
	return func(m *MonitorEvent) {
		m.logger = logger
	}
}

func (MonitorEventOptions) WithTracer(tracer opentracing2.Tracer) EventOption {
	return func(m *MonitorEvent) {
		m.tracer = tracer
	}
}

func (m *MonitorEvent) NextEvent(me MgoEvent) {
	m.nextEvent = me
}

func (m *MonitorEvent) EventName() string {
	return "monitor_proxy"
}

func (m *MonitorEvent) Start(ctx context.Context, starEv *event.CommandStartedEvent) {
	if m.nextEvent != nil {
		m.nextEvent.Start(ctx, starEv)
	}
}

func (m *MonitorEvent) SucceededEvent(ctx context.Context,
	succEvent *event.CommandSucceededEvent) {
	var beginTime time.Time
	monBackEvent := &mongoBackEvent{}
	endTime := time.Now()

	monBackEvent.succEvent = succEvent
	beginTime = endTime.Add(-time.Duration(succEvent.DurationNanos))

	if m.nextEvent != nil {
		m.nextEvent.SucceededEvent(ctx, succEvent)
	}

	for _, ev := range m.afterEvents {
		ev(ctx, monBackEvent, beginTime, endTime)
	}
}

func (m *MonitorEvent) FailedEvent(ctx context.Context, failedEvent *event.CommandFailedEvent) {
	var beginTime time.Time
	monBackEvent := &mongoBackEvent{}
	endTime := time.Now()

	monBackEvent.failedEvent = failedEvent
	beginTime = endTime.Add(-time.Duration(failedEvent.DurationNanos))

	if m.nextEvent != nil {
		m.nextEvent.FailedEvent(ctx, failedEvent)
	}

	for _, ev := range m.afterEvents {
		ev(ctx, monBackEvent, beginTime, endTime)
	}
}

func (m *MonitorEvent) registerAfterEvent() {
	if m.conf.GetBool("mgo_tracer") {
		m.afterEvents = append(m.afterEvents, m.withTracer)
	}

	if m.conf.GetBool("mgo_check_slow") {
		m.afterEvents = append(m.afterEvents, m.withSlowCommand)
	}

	if m.conf.GetBool("mgo_metrics") {
		m.afterEvents = append(m.afterEvents, m.withMetrics)
	}

	if m.conf.GetBool("debug") {
		m.afterEvents = append(m.afterEvents, m.withDebug)
	}
}

// 执行慢的命令
// dur_nan 纳秒
// 执行的命令
func (m *MonitorEvent) withSlowCommand(ctx context.Context, backEvent *mongoBackEvent,
	beginTime, endTime time.Time) {
	mgoSlowTime := m.conf.GetInt64("mgo_slow_time")
	execCommand, ok := ctx.Value("command").(*string)

	var durNan int64
	if backEvent.succEvent != nil {
		durNan = backEvent.succEvent.DurationNanos
	} else if backEvent.failedEvent != nil {
		durNan = backEvent.failedEvent.DurationNanos
	}

	if ok && durNan != 0 && mgoSlowTime != 0 {
		if durNan/1000000 >= mgoSlowTime {
			m.logger.Warnf("slow command %s", execCommand)
		}
	}
}

func (m *MonitorEvent) withTracer(ctx context.Context, backEvent *mongoBackEvent,
	beginTime, endTime time.Time) {
	execCommand, ok := ctx.Value("command").(*string)

	if ok {
		var commandName string
		var errStr string

		if backEvent.succEvent != nil {
			commandName = backEvent.succEvent.CommandName
		} else if backEvent.failedEvent != nil {
			commandName = backEvent.failedEvent.CommandName
			errStr = backEvent.failedEvent.Failure
		}

		if commandName != "" {
			span := opentracing.GetSpan(ctx, m.tracer,
				commandName, beginTime)
			if errStr != "" {
				span.SetTag("error", true)
				span.LogKV("error_detailed", errStr)
			}
			span.LogKV("exec_command", *execCommand)
			span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: endTime})
		}
	}
}

func (m *MonitorEvent) withMetrics(ctx context.Context, backEvent *mongoBackEvent,
	beginTime, endTime time.Time) {
	var commandName string

	if backEvent.succEvent != nil {
		commandName = backEvent.succEvent.CommandName
	} else if backEvent.failedEvent != nil {
		commandName = backEvent.failedEvent.CommandName
	}

	if commandName != "" {
		lab := prometheus.Labels{"command": commandName}
		mongodbErrTotal.With(lab).Inc()
		mongodbDuration.With(lab).Observe(endTime.Sub(beginTime).Seconds())
	}
}

func (m *MonitorEvent) withDebug(ctx context.Context, backEvent *mongoBackEvent,
	beginTime, endTime time.Time) {
	command, ok := ctx.Value("command").(*string)
	if ok {
		if backEvent.succEvent != nil {
			m.logger.Debugf("mongodb success [%v] %s",
				endTime.Sub(beginTime).String(), *command)
		} else if backEvent.failedEvent != nil {
			m.logger.Debugf("mongodb fail [%v] %s",
				endTime.Sub(beginTime).String(), *command)
		}
	}
}
