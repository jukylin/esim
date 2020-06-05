package log

import (
	"context"
	"os"
	"reflect"
	"testing"

	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

var tracer opentracing.Tracer

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	cfg, _ := jaegerConfig.FromEnv()
	cfg.ServiceName = "logger"
	tracer, _, _ = cfg.NewTracer()
}

func tearDown() {

}

func TestLog(t *testing.T) {
	loggerOptions := LoggerOptions{}
	logger := NewLogger(loggerOptions.WithDebug(false))

	sp := tracer.StartSpan("test")
	ctx := opentracing.ContextWithSpan(context.Background(), sp)

	logger.Debugf("debug")

	logger.Debugc(ctx, "debug")

	logger.Infof("info")

	logger.Infoc(ctx, "info")

	logger.Warnf("warn")

	logger.Warnc(ctx, "warn")
}

func Test_logger_getArgs(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	log := new(logger)
	sp := tracer.StartSpan("test")
	jaegerCtx := opentracing.ContextWithSpan(context.Background(), sp)

	esimTracerID := tracerid.TracerID()()
	esimCtx := context.WithValue(context.Background(), tracerid.ActiveEsimKey, esimTracerID)

	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{"jaeger_tracer_id", args{jaegerCtx},
			[]interface{}{
				"caller", "testing/testing.go:991", "tracer_id",
				sp.Context().(jaeger.SpanContext).TraceID().String(),
			}},
		{"esim_tracer_id", args{esimCtx},
			[]interface{}{
				"caller", "testing/testing.go:991", "tracer_id", esimTracerID,
			}},
		{"empty_tracer_id", args{context.Background()},
			[]interface{}{
				"caller", "testing/testing.go:991",
			}},
		{"nil_ctx", args{context.Background()},
			[]interface{}{
				"caller", "testing/testing.go:991",
			}},
	}

	for k := range tests {
		test := tests[k]
		t.Run(test.name, func(t *testing.T) {
			if got := log.getArgs(test.args.ctx); !reflect.DeepEqual(got, test.want) {
				t.Errorf("logger.getArgs() = %v, want %v", got, test.want)
			}
		})
	}
}
