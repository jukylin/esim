package container

import (
	"sync"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/prometheus"
	"github.com/opentracing/opentracing-go"
)

var esimOnce sync.Once
var onceEsim *Esim

type Option func(c *Esim)

type Esim struct {
	Z *log.EsimZap

	prometheus *prometheus.Prometheus

	Logger log.Logger

	Conf config.Config

	Tracer opentracing.Tracer
}

func WithTracer(tracer opentracing.Tracer) Option {
	return func(e *Esim) {
		e.Tracer = tracer
	}
}

func WithLogger(logger log.Logger) Option {
	return func(e *Esim) {
		e.Logger = logger
	}
}

func WithConf(conf config.Config) Option {
	return func(e *Esim) {
		e.Conf = conf
	}
}

func WithPromer(promer *prometheus.Prometheus) Option {
	return func(e *Esim) {
		e.prometheus = promer
	}
}

func WithEsimZap(ez *log.EsimZap) Option {
	return func(e *Esim) {
		e.Z = ez
	}
}

func NewEsim(options ...Option) *Esim {
	esimOnce.Do(func() {
		onceEsim = &Esim{}
		for _, option := range options {
			option(onceEsim)
		}
	})
	return onceEsim
}

func (e *Esim) String() string {
	return "Working with Esim."
}
