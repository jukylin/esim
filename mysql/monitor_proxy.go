package mysql

import (
	"context"
	"database/sql"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	"time"
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
func (this *monitorProxy) NextProxy(db interface{}) {
	this.nextProxy = db.(SqlCommon)
}


//implement Proxy interface
func (this *monitorProxy) ProxyName() string {
	return this.name
}


func (this *monitorProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	startTime := time.Now()
	result, err := this.nextProxy.Exec(query, args...)
	this.after(query, startTime)
	return result, err
}


func (this *monitorProxy) Prepare(query string) (*sql.Stmt, error) {
	startTime := time.Now()
	stmt, err := this.nextProxy.Prepare(query)
	this.after(query, startTime)

	return stmt, err
}


func (this *monitorProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	startTime := time.Now()
	rows, err := this.nextProxy.Query(query, args...)
	this.after(query, startTime)

	return rows, err
}


func (this *monitorProxy) QueryRow(query string, args ...interface{}) *sql.Row {
	startTime := time.Now()
	row := this.nextProxy.QueryRow(query, args...)
	this.after(query, startTime)

	return row
}


func (this *monitorProxy) Close() error {
	return this.nextProxy.Close()
}


func (this *monitorProxy) Begin() (*sql.Tx, error){
	return this.nextProxy.Begin()
}


func (this *monitorProxy) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error){
	return this.nextProxy.BeginTx(ctx, opts)
}


func (this *monitorProxy) registerAfterEvent() {
	if this.conf.GetBool("mysql_tracer") == true {
		this.afterEvents = append(this.afterEvents, this.withMysqlTracer)
	}

	if this.conf.GetBool("mysql_check_slow") == true {
		this.afterEvents = append(this.afterEvents, this.withSlowSql)
	}

	if this.conf.GetBool("mysql_metrics") == true {
		this.afterEvents = append(this.afterEvents, this.withMysqlMetrics)
	}
}


func (this *monitorProxy) after(query string, beginTime time.Time) {
	now := time.Now()
	for _, event := range this.afterEvents {
		event(query, beginTime, now)
	}
}


func (this *monitorProxy) withSlowSql(query string, beginTime, endTime time.Time) {
	mysql_slow_time := this.conf.GetInt64("mysql_slow_time")

	if mysql_slow_time != 0 {
		if endTime.Sub(beginTime) > time.Duration(mysql_slow_time)*time.Millisecond {
			this.log.Warnf("slow sql %s", query)
		}
	}
}

func (this *monitorProxy) withMysqlMetrics(query string, beginTime, endTime time.Time) {
	lab := prometheus.Labels{"sql": query}
	mysqlTotal.With(lab).Inc()
	mysqlDuration.With(lab).Observe(endTime.Sub(beginTime).Seconds())
}

//要等2.0
func (this *monitorProxy) withMysqlTracer(query string, beginTime, endTime time.Time) {
	//span := opentracing.GetSpan(ctx, m.tracer,
	//	query, beginTime)
	//span.LogKV("sql", query)
	//span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: endTime})
}
