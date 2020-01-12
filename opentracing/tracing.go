package opentracing

import (
	"time"

	"golang.org/x/net/context"
	"github.com/opentracing/opentracing-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"github.com/jukylin/esim/log"
)


func NewTracer(serviceName string, logger log.Logger) opentracing.Tracer {

	var tracer opentracing.Tracer

	cfg, err := jaegerconfig.FromEnv()
	if err != nil {
		logger.Panicf(err.Error())
	}

	cfg.ServiceName = serviceName
	//default close tracer
	if cfg.Sampler.Type == "" {
		cfg.Disabled = true
	}

	tracer, _, err = cfg.NewTracer(jaegerconfig.Logger(logger))
	if err != nil{
		logger.Panicf(err.Error())
	}

	return tracer
}


func GetSpan(ctx context.Context, tracer opentracing.Tracer,
	operationName string, begin_time time.Time) (opentracing.Span){

	if parSpan := opentracing.SpanFromContext(ctx); parSpan != nil {
		spanOption := opentracing.StartSpanOptions{}
		spanOption.StartTime = begin_time

		span := tracer.StartSpan(operationName, opentracing.ChildOf(parSpan.Context()),
			opentracing.StartTime(begin_time))
		return span
	}
	return nil
}


func FinishWithOptions(ctx context.Context, tracer opentracing.Tracer, operationName string,
	begin_time time.Time) (opentracing.Span){

	if parSpan := opentracing.SpanFromContext(ctx); parSpan != nil {
		spanOption := opentracing.StartSpanOptions{}
		spanOption.StartTime = begin_time

		span := tracer.StartSpan(operationName, opentracing.ChildOf(parSpan.Context()),
			opentracing.StartTime(begin_time))
		return span
	}
	return nil
}