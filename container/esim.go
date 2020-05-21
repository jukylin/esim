package container

import (
	"sync"

	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	eot "github.com/jukylin/esim/opentracing"
	"github.com/jukylin/esim/prometheus"
	"github.com/opentracing/opentracing-go"
)

var esimOnce sync.Once
var onceEsim *Esim

const DefaultAppname = "esim"
const DefaultPrometheusHTTPArrd = "9002"

// esim init start
type Esim struct {
	prometheus *prometheus.Prometheus

	Logger log.Logger

	Conf config.Config

	Tracer opentracing.Tracer
}

//nolint:varcheck,unused,deadcode
var esimSet = wire.NewSet(
	wire.Struct(new(Esim), "*"),
	provideConf,
	provideLogger,
	providePrometheus,
	provideTracer,
)

var confFunc = func() config.Config {
	return config.NewMemConfig()
}

func SetConfFunc(conf func() config.Config) {
	confFunc = conf
}
func provideConf() config.Config {
	return confFunc()
}

var prometheusFunc = func(conf config.Config, logger log.Logger) *prometheus.Prometheus {
	var httpAddr string
	if conf.GetString("prometheus_http_addr") != "" {
		httpAddr = conf.GetString("prometheus_http_addr")
	} else {
		httpAddr = DefaultPrometheusHTTPArrd
	}
	return prometheus.NewPrometheus(httpAddr, logger)
}

func SetPrometheusFunc(pt func(config.Config, log.Logger) *prometheus.Prometheus) {
	prometheusFunc = pt
}
func providePrometheus(conf config.Config, logger log.Logger) *prometheus.Prometheus {
	return prometheusFunc(conf, logger)
}

var loggerFunc = func(conf config.Config) log.Logger {
	var loggerOptions log.LoggerOptions

	logger := log.NewLogger(
		loggerOptions.WithDebug(conf.GetBool("debug")),
	)
	return logger
}

func SetLogger(logger func(config.Config) log.Logger) {
	loggerFunc = logger
}
func provideLogger(conf config.Config) log.Logger {
	return loggerFunc(conf)
}

var tracerFunc = func(conf config.Config, logger log.Logger) opentracing.Tracer {
	var appname string
	if conf.GetString("appname") != "" {
		appname = conf.GetString("appname")
	} else {
		appname = DefaultAppname
	}
	return eot.NewTracer(appname, logger)
}

func SetTracer(tracer func(config.Config, log.Logger) opentracing.Tracer) {
	tracerFunc = tracer
}
func provideTracer(conf config.Config, logger log.Logger) opentracing.Tracer {
	return tracerFunc(conf, logger)
}

// esim init end

func NewEsim() *Esim {
	esimOnce.Do(func() {
		onceEsim = initEsim()
	})

	return onceEsim
}

func (e *Esim) String() string {
	return "相信，相信自己！！！"
}
