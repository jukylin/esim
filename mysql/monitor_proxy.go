package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
)

type MonitorProxy struct {
	// proxy name
	name string

	nextProxy ConnPool

	tracer opentracing2.Tracer

	conf config.Config

	logger log.Logger

	afterEvents []afterEvents
}

type afterEvents func(context.Context, string, time.Time, time.Time)

type MonitorProxyOption func(c *MonitorProxy)

type MonitorProxyOptions struct{}

func NewMonitorProxy(options ...MonitorProxyOption) *MonitorProxy {
	MonitorProxy := &MonitorProxy{}

	for _, option := range options {
		option(MonitorProxy)
	}

	if MonitorProxy.conf == nil {
		MonitorProxy.conf = config.NewNullConfig()
	}

	if MonitorProxy.logger == nil {
		MonitorProxy.logger = log.NewLogger()
	}

	if MonitorProxy.tracer == nil {
		MonitorProxy.tracer = opentracing.NewTracer("mysql", MonitorProxy.logger)
	}

	MonitorProxy.name = "monitor_proxy"

	MonitorProxy.registerAfterEvent()

	return MonitorProxy
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
func (mp *MonitorProxy) NextProxy(db interface{}) {
	mp.nextProxy = db.(ConnPool)
}

// Implement Proxy interface.
func (mp *MonitorProxy) ProxyName() string {
	return mp.name
}

func (mp *MonitorProxy) ExecContext(ctx context.Context, query string,
	args ...interface{}) (sql.Result, error) {
	startTime := time.Now()
	result, err := mp.nextProxy.ExecContext(ctx, query, args...)
	mp.after(ctx, query, startTime)
	return result, err
}

func (mp *MonitorProxy) PrepareContext(ctx context.Context,
	query string) (*sql.Stmt, error) {
	startTime := time.Now()
	stmt, err := mp.nextProxy.PrepareContext(ctx, query)
	mp.after(ctx, query, startTime)

	return stmt, err
}

func (mp *MonitorProxy) QueryContext(ctx context.Context, query string,
	args ...interface{}) (*sql.Rows, error) {
	startTime := time.Now()
	rows, err := mp.nextProxy.QueryContext(ctx, query, args...)
	mp.after(ctx, query, startTime)

	return rows, err
}

func (mp *MonitorProxy) QueryRowContext(ctx context.Context, query string,
	args ...interface{}) *sql.Row {
	startTime := time.Now()
	row := mp.nextProxy.QueryRowContext(ctx, query, args...)
	mp.after(ctx, query, startTime)

	return row
}

func (mp *MonitorProxy) Close() error {
	return mp.nextProxy.Close()
}

func (mp *MonitorProxy) registerAfterEvent() {
	if mp.conf.GetBool("mysql_trace") {
		mp.afterEvents = append(mp.afterEvents, mp.withMysqlTracer)
	}

	if mp.conf.GetBool("mysql_check_slow") {
		mp.afterEvents = append(mp.afterEvents, mp.withSlowSQL)
	}

	if mp.conf.GetBool("mysql_metrics") {
		mp.afterEvents = append(mp.afterEvents, mp.withMysqlMetrics)
	}
}

func (mp *MonitorProxy) after(ctx context.Context, query string, beginTime time.Time) {
	now := time.Now()
	for _, event := range mp.afterEvents {
		event(ctx, query, beginTime, now)
	}
}

func (mp *MonitorProxy) withSlowSQL(ctx context.Context, query string,
	beginTime, endTime time.Time) {
	mysqlSlowTime := mp.conf.GetInt64("mysql_slow_time")
	if mysqlSlowTime != 0 {
		diffTime := endTime.Sub(beginTime)
		if diffTime > time.Duration(mysqlSlowTime)*time.Millisecond {
			mp.logger.Warnf("Slow sql %d : %s", diffTime, query)
		}
	}
}

func (mp *MonitorProxy) withMysqlMetrics(ctx context.Context, query string,
	beginTime, endTime time.Time) {
	lab := prometheus.Labels{"sql": query}
	mysqlTotal.With(lab).Inc()
	mysqlDuration.With(lab).Observe(endTime.Sub(beginTime).Seconds())
}

func (mp *MonitorProxy) withMysqlTracer(ctx context.Context, query string,
	beginTime, endTime time.Time) {
	span := opentracing.GetSpan(ctx, mp.tracer, "sql", beginTime)
	span.LogKV("sql", query)
	span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: endTime})
}
