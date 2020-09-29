package tracerid

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/utils"
)

type contextKey struct{}

var ActiveEsimKey = contextKey{}

// TracerID 用于进程内，不支持分布式.
func TracerID() func() string {
	seedGenerator := utils.NewRand(time.Now().UnixNano())
	pool := sync.Pool{
		New: func() interface{} {
			return rand.NewSource(seedGenerator.Int63())
		},
	}

	randomNumber := func() string {
		generator := pool.Get().(rand.Source)
		number := uint64(generator.Int63())
		pool.Put(generator)
		return fmt.Sprintf("%x", number)
	}

	return randomNumber
}

func ExtractTracerID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	sp := opentracing.SpanFromContext(ctx)
	if sp != nil {
		if jaegerSpanContext, ok := sp.Context().(jaeger.SpanContext); ok {
			return jaegerSpanContext.TraceID().String()
		}
	}

	val := ctx.Value(ActiveEsimKey)
	if tracerID, ok := val.(string); ok {
		return tracerID
	}

	return ""
}
