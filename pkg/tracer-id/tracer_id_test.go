package tracerid

import (
	"context"
	"os"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
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

func TestTracerID(t *testing.T) {
	assert.NotEmpty(t, TracerID()())
}

func TestExtractTracerID(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	tracerId := TracerID()()
	ctx := context.Background()
	esimKeyCtx := context.WithValue(ctx, ActiveEsimKey, tracerId)

	sp := tracer.StartSpan("test")
	spanCtx := opentracing.ContextWithSpan(context.Background(), sp)

	tests := []struct {
		name string
		args args
		want string
	}{
		{"ctx为nil", args{nil}, ""},
		{"esim 生成的 tracerId", args{esimKeyCtx}, tracerId},
		{"没有tracerId", args{ctx}, ""},
		{"Jaeger 生成的 tracerId", args{spanCtx}, sp.Context().(jaeger.SpanContext).TraceID().String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractTracerID(tt.args.ctx); got != tt.want {
				t.Errorf("ExtractTracerID() = %v, want %v", got, tt.want)
			}
		})
	}
}
