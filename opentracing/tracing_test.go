package opentracing

import (
	"context"
	"reflect"
	"testing"
	"time"

	"os"

	"github.com/jukylin/esim/log"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

const (
	getEnvPanic    = "get_env_panic"
	notServiceName = "not_service_name"
)

func TestNewTracer(t *testing.T) {
	type args struct {
		serviceName string
		logger      log.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		{"test", args{"test",
			log.NewLogger(log.WithDebug(true))}},
		{getEnvPanic, args{"test",
			log.NewLogger(log.WithDebug(true))}},
		{notServiceName, args{"",
			log.NewLogger(log.WithDebug(true))}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == getEnvPanic {
				os.Setenv("JAEGER_DISABLED", "100")
				assert.Panics(t, func() {
					NewTracer(tt.args.serviceName, tt.args.logger)
				})
			} else if tt.name == notServiceName {
				assert.Panics(t, func() {
					NewTracer(tt.args.serviceName, tt.args.logger)
				})
			} else {
				tracer := NewTracer(tt.args.serviceName, tt.args.logger)
				assert.Implements(t, (*opentracing.Tracer)(nil), tracer)
			}

			if tt.name == getEnvPanic {
				os.Setenv("JAEGER_DISABLED", "false")
			}
		})
	}
}

func TestGetSpan(t *testing.T) {
	type args struct {
		ctx           context.Context
		tracer        opentracing.Tracer
		operationName string
		beginTime     time.Time
	}

	ctx := context.Background()
	tracer := NewTracer("test", log.NewNullLogger())
	spanCtx := opentracing.ContextWithSpan(ctx, tracer.StartSpan("withSpan"))

	tests := []struct {
		name string
		args args
		want opentracing.Span
	}{
		{"noopTracer", args{ctx, &opentracing.NoopTracer{},
			"test", time.Now()},
			opentracing.StartSpan("test")},
		{"notParentSpan", args{ctx,
			&opentracing.NoopTracer{},
			"test", time.Now()},
			opentracing.StartSpan("test")},
		{"ParentSpan", args{spanCtx,
			&opentracing.NoopTracer{},
			"test", time.Now()},
			opentracing.StartSpan("test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSpan(tt.args.ctx, tt.args.tracer,
				tt.args.operationName, tt.args.beginTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSpan() = %v, want %v", got, tt.want)
			}
		})
	}
}
