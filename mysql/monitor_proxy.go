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

type monitorProxy struct {
	//proxy name
	name string

	nextProxy SqlCommon

	tracer opentracing2.Tracer

	conf config.Config

	log log.Logger

	afterEvents []afterEvents
}

type afterEvents func(string, time.Time, time.Time)

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
		monitorProxy.tracer = opentracing.NewTracer("mysql", monitorProxy.log)
	}

	monitorProxy.name = "monitor_proxy"

	monitorProxy.registerAfterEvent()

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
func (mp *monitorProxy) NextProxy(db interface{}) {
	mp.nextProxy = db.(SqlCommon)
}

//implement Proxy interface
func (mp *monitorProxy) ProxyName() string {
	return mp.name
}

func (mp *monitorProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	startTime := time.Now()
	result, err := mp.nextProxy.Exec(query, args...)
	mp.after(query, startTime)
	return result, err
}

func (mp *monitorProxy) Prepare(query string) (*sql.Stmt, error) {
	startTime := time.Now()
	stmt, err := mp.nextProxy.Prepare(query)
	mp.after(query, startTime)

	return stmt, err
}

func (mp *monitorProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	startTime := time.Now()
	rows, err := mp.nextProxy.Query(query, args...)
	mp.after(query, startTime)

	return rows, err
}

func (mp *monitorProxy) QueryRow(query string, args ...interface{}) *sql.Row {
	startTime := time.Now()
	row := mp.nextProxy.QueryRow(query, args...)
	mp.after(query, startTime)

	return row
}

func (mp *monitorProxy) Close() error {
	return mp.nextProxy.Close()
}

func (mp *monitorProxy) Begin() (*sql.Tx, error) {
	return mp.nextProxy.Begin()
}

func (mp *monitorProxy) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return mp.nextProxy.BeginTx(ctx, opts)
}

func (mp *monitorProxy) registerAfterEvent() {
	if mp.conf.GetBool("mysql_tracer") {
		mp.afterEvents = append(mp.afterEvents, mp.withMysqlTracer)
	}

	if mp.conf.GetBool("mysql_check_slow") {
		mp.afterEvents = append(mp.afterEvents, mp.withSlowSql)
	}

	if mp.conf.GetBool("mysql_metrics") {
		mp.afterEvents = append(mp.afterEvents, mp.withMysqlMetrics)
	}
}

func (mp *monitorProxy) after(query string, beginTime time.Time) {
	now := time.Now()
	for _, event := range mp.afterEvents {
		event(query, beginTime, now)
	}
}

func (mp *monitorProxy) withSlowSql(query string, beginTime, endTime time.Time) {
	mysqlSlowTime := mp.conf.GetInt64("mysql_slow_time")

	if mysqlSlowTime != 0 {
		if endTime.Sub(beginTime) > time.Duration(mysqlSlowTime)*time.Millisecond {
			mp.log.Warnf("slow sql %s", query)
		}
	}
}

func (mp *monitorProxy) withMysqlMetrics(query string, beginTime, endTime time.Time) {
	lab := prometheus.Labels{"sql": query}
	mysqlTotal.With(lab).Inc()
	mysqlDuration.With(lab).Observe(endTime.Sub(beginTime).Seconds())
}

//要等2.0
func (mp *monitorProxy) withMysqlTracer(query string, beginTime, endTime time.Time) {
	//span := opentracing.GetSpan(ctx, m.tracer,
	//	query, beginTime)
	//span.LogKV("sql", query)
	//span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: endTime})
}
