package opentracing

import (
	"time"

	"github.com/jukylin/esim/log"
	"github.com/opentracing/opentracing-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"golang.org/x/net/context"
)

func NewTracer(serviceName string, logger log.Logger) opentracing.Tracer {
	var tracer opentracing.Tracer

	cfg, err := jaegerconfig.FromEnv()
	if err != nil {
		logger.Panicf(err.Error())
	}

	cfg.ServiceName = serviceName
	//cfg.Sampler.Type = "const"
	//cfg.Sampler.Param = 1
	tracer, _, err = cfg.NewTracer(jaegerconfig.Logger(logger))
	if err != nil {
		logger.Panicf(err.Error())
	}

	return tracer
}

func GetSpan(ctx context.Context, tracer opentracing.Tracer,
	operationName string, beginTime time.Time) opentracing.Span {

	if parSpan := opentracing.SpanFromContext(ctx); parSpan != nil {
		spanOption := opentracing.StartSpanOptions{}
		spanOption.StartTime = beginTime

		span := tracer.StartSpan(operationName, opentracing.ChildOf(parSpan.Context()),
			opentracing.StartTime(beginTime))
		return span
	}
	return nil
}

func FinishWithOptions(ctx context.Context, tracer opentracing.Tracer, operationName string,
	beginTime time.Time) opentracing.Span {

	if parSpan := opentracing.SpanFromContext(ctx); parSpan != nil {
		spanOption := opentracing.StartSpanOptions{}
		spanOption.StartTime = beginTime

		span := tracer.StartSpan(operationName, opentracing.ChildOf(parSpan.Context()),
			opentracing.StartTime(beginTime))
		return span
	}
	return nil
}
