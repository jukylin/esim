package container

import (
	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/prometheus"
	"github.com/opentracing/opentracing-go"
)

func provideMockConf() config.Config {
	conf := config.NewMemConfig()
	conf.Set("debug", true)
	return conf
}

func provideMockProme(conf config.Config) *prometheus.Prometheus {
	return prometheus.NewNullProme()
}


func provideNoopTracer() opentracing.Tracer {
	return opentracing.NoopTracer{}
}

var MockSet = wire.NewSet(
	wire.Struct(new(Esim), "*"),
	provideMockConf,
	provideLogger,
	provideMockProme,
	provideNoopTracer,
)
