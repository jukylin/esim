package container

import (
	"sync"
	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/prometheus"
	"github.com/opentracing/opentracing-go"
	eot "github.com/jukylin/esim/opentracing"
)

var esimOnce sync.Once
var onceEsim *Esim

//esim init start
type Esim struct {
	prometheus *prometheus.Prometheus

	Logger log.Logger

	Conf config.Config

	Tracer opentracing.Tracer
}

var esimSet = wire.NewSet(
	wire.Struct(new(Esim), "*"),
	provideConf,
	provideLogger,
	providePrometheus,
	provideTracer,
)

var confFunc  = func() config.Config {
	return config.NewNullConfig()
}
func SetConfFunc(conf func() config.Config) {
	confFunc = conf
}
func provideConf() config.Config {
	return confFunc()
}

var prometheusFunc = func(conf config.Config, log log.Logger) *prometheus.Prometheus {
	return prometheus.NewPrometheus(conf, log)
}
func SetPrometheusFunc(prometheus func(conf config.Config, log log.Logger) *prometheus.Prometheus) {
	prometheusFunc = prometheus
}
func providePrometheus(conf config.Config, log log.Logger) *prometheus.Prometheus {
	return prometheusFunc(conf, log)
}


var loggerFunc = func(conf config.Config) log.Logger {
	var loggerOptions log.LoggerOptions

	logger := log.NewLogger(
		loggerOptions.WithConf(conf),
		loggerOptions.WithDebug(conf.GetBool("debug")),
	)
	return logger
}
func SetLogger(log func(config.Config) log.Logger) {
	loggerFunc = log
}
func provideLogger(conf config.Config) log.Logger {
	return loggerFunc(conf)
}


var tracerFunc = func(conf config.Config, logger log.Logger) opentracing.Tracer {
	return eot.NewTracer(conf.GetString("appname"), logger)
}
func SetTracer(tracer func(conf config.Config, logger log.Logger) opentracing.Tracer) {
	tracerFunc = tracer
}
func provideTracer(conf config.Config, logger log.Logger) opentracing.Tracer {
	return tracerFunc(conf, logger)
}

//esim init end

//使用单例模式，基础设施为全局资源
func NewEsim() *Esim {
	esimOnce.Do(func() {
		onceEsim = initEsim()
	})

	return onceEsim
}

func (this *Esim) String() string {
	return "相信，相信自己！！！"
}
