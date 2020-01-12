package log

import (
	"testing"
	opentracing "github.com/opentracing/opentracing-go"
	"context"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"github.com/stretchr/testify/assert"
	jaeger "github.com/uber/jaeger-client-go"
)

func initJaeger() (opentracing.Tracer, error) {
	cfg, err := jaegerConfig.FromEnv()
	if err != nil{
		return nil, err
	}

	cfg.ServiceName = "logger"
	tracer, _, err := cfg.NewTracer()

	return tracer, err
}

func TestLog(t *testing.T)  {

	loggerOptions := LoggerOptions{}

	logger := NewLogger(loggerOptions.WithDebug(false))

	tracer, err := initJaeger()
	assert.Nil(t, err)

	sp := tracer.StartSpan("test")
	ctx := opentracing.ContextWithSpan(context.Background() , sp)

	logger.Debugf("debug")

	logger.Debugc(ctx, "debug")

	logger.Infof("info")

	logger.Infoc(ctx, "info")

	logger.Warnf("warn")

	logger.Warnc(ctx, "warn")
}

func TestGetTracerId(t *testing.T)  {

	loggerOptions := LoggerOptions{}

	log := NewLogger(loggerOptions.WithDebug(false))

	tracer, err := initJaeger()
	assert.Nil(t, err)

	sp := tracer.StartSpan("test")
	ctx := opentracing.ContextWithSpan(context.Background() , sp)

	assert.Equal(t, sp.Context().(jaeger.SpanContext).TraceID().String(),
		log.(*logger).getTracerId(ctx))
}


func TestGetTracerIdEmpty(t *testing.T)  {

	loggerOptions := LoggerOptions{}

	log := NewLogger(loggerOptions.WithDebug(false))

	ctx := context.Background()
	assert.Empty(t, log.(*logger).getTracerId(ctx))
}

