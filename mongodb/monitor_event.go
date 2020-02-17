package mongodb

import (
	"context"
	"time"

	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	"go.mongodb.org/mongo-driver/event"
)

type monitorEvent struct {
	nextEvent MonitorEvent

	conf config.Config

	logger log.Logger

	tracer opentracing2.Tracer

	afterEvents []afterEvents
}

type afterEvents func(context.Context, *mongoBackEvent, time.Time, time.Time)

type MonitorEventOptions struct {}

type EventOption func(c *monitorEvent)

func NewMonitorEvent(options ...EventOption) *monitorEvent {

	m := &monitorEvent{}

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
	return func(m *monitorEvent) {
		m.conf = conf
	}
}

func (MonitorEventOptions) WithLogger(logger log.Logger) EventOption {
	return func(m *monitorEvent) {
		m.logger = logger
	}
}

func (MonitorEventOptions) WithTracer(tracer opentracing2.Tracer) EventOption {
	return func(m *monitorEvent) {
		m.tracer = tracer
	}
}

func (m *monitorEvent) NextEvent(event MonitorEvent) {
	m.nextEvent = event
}

func (m *monitorEvent) EventName() string {
	return "monitor_proxy"
}

func (m *monitorEvent) Start(ctx context.Context, starEv *event.CommandStartedEvent) {
	if m.nextEvent != nil {
		m.nextEvent.Start(ctx, starEv)
	}
}


func (m *monitorEvent) SucceededEvent(ctx context.Context,
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


func (m *monitorEvent) FailedEvent(ctx context.Context, failedEvent *event.CommandFailedEvent) {

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

func (m *monitorEvent) registerAfterEvent() {
	if m.conf.GetBool("mgo_tracer") == true {
		m.afterEvents = append(m.afterEvents, m.withTracer)
	}

	if m.conf.GetBool("mgo_check_slow") == true {
		m.afterEvents = append(m.afterEvents, m.withSlowCommand)
	}

	if m.conf.GetBool("mgo_metrics") == true {
		m.afterEvents = append(m.afterEvents, m.withMetrics)
	}

	if m.conf.GetBool("debug") == true {
		m.afterEvents = append(m.afterEvents, m.withDebug)
	}
}

//执行慢的命令
// dur_nan 纳秒
//执行的命令
func (m *monitorEvent) withSlowCommand(ctx context.Context, backEvent *mongoBackEvent, begin_time time.Time, end_time time.Time) {
	mgo_slow_time := m.conf.GetInt64("mgo_slow_time")
	exec_command, ok := ctx.Value("command").(*string)

	var dur_nan int64
	if backEvent.succEvent != nil {
		dur_nan = backEvent.succEvent.DurationNanos
	} else if backEvent.failedEvent != nil {
		dur_nan = backEvent.failedEvent.DurationNanos
	}

	if ok == true && dur_nan != 0 && mgo_slow_time != 0 {
		if int64(dur_nan/1000000) >= int64(mgo_slow_time) {
			m.logger.Warnf("slow command %s", exec_command)
		}
	}
}

func (m *monitorEvent) withTracer(ctx context.Context, backEvent *mongoBackEvent, begin_time time.Time, end_time time.Time) {
	exec_command, ok := ctx.Value("command").(*string)

	if ok {
		var command_name string
		var err_str string

		if backEvent.succEvent != nil {
			command_name = backEvent.succEvent.CommandName
		} else if backEvent.failedEvent != nil {
			command_name = backEvent.failedEvent.CommandName
			err_str = backEvent.failedEvent.Failure
		}

		if command_name != "" {
			span := opentracing.GetSpan(ctx, m.tracer,
				command_name, begin_time)
			if err_str != "" {
				span.SetTag("error", true)
				span.LogKV("error_detailed", err_str)
			}
			span.LogKV("exec_command", *exec_command)
			span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: end_time})
		}
	}

}

func (m *monitorEvent) withMetrics(ctx context.Context, backEvent *mongoBackEvent, begin_time time.Time, end_time time.Time) {

	var command_name string

	if backEvent.succEvent != nil {
		command_name = backEvent.succEvent.CommandName
	} else if backEvent.failedEvent != nil {
		command_name = backEvent.failedEvent.CommandName
	}

	if command_name != "" {
		lab := prometheus.Labels{"command": command_name}
		mongodbErrTotal.With(lab).Inc()
		mongodbDuration.With(lab).Observe(end_time.Sub(begin_time).Seconds())
	}
}

func (m *monitorEvent) withDebug(ctx context.Context, backEvent *mongoBackEvent, begin_time time.Time, end_time time.Time) {
	command, ok := ctx.Value("command").(*string)
	if ok {
		if backEvent.succEvent != nil {
			m.logger.Debugf("mongodb success [%v] %s",
				end_time.Sub(begin_time).String(), *command)
		} else if backEvent.failedEvent != nil {
			m.logger.Debugf("mongodb fail [%v] %s",
				end_time.Sub(begin_time).String(), *command)
		}
	}
}
