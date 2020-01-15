package middle_ware

import (
	"net/http"
	"time"

	"github.com/jukylin/esim/opentracing"
	"github.com/jukylin/esim/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/gin-gonic/gin"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func GinMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		requestTotal.With(prometheus.Labels{"method": c.Request.Method, "endpoint": c.Request.Host}).Inc()
		requestDuration.With(prometheus.Labels{"method": c.Request.Method,
		"endpoint": c.Request.Host}).Observe(duration.Seconds())
	}
}

func GinTracer(serviceName string, logger log.Logger) gin.HandlerFunc {
	tracer := opentracing.NewTracer(serviceName, logger)
	return func(c *gin.Context) {
		spContext, _ := tracer.Extract(opentracing2.HTTPHeaders,
			opentracing2.HTTPHeadersCarrier(c.Request.Header))

		sp := tracer.StartSpan("HTTP " + c.Request.Method,
			ext.RPCServerOption(spContext))

		ext.HTTPMethod.Set(sp, c.Request.Method)
		ext.HTTPUrl.Set(sp, c.Request.URL.String())
		ext.Component.Set(sp, "net/http")

		c.Request = c.Request.WithContext(opentracing2.ContextWithSpan(c.Request.Context(), sp))
		c.Next()

		ext.HTTPStatusCode.Set(sp, uint16(c.Writer.Status()))
		if c.Writer.Status() >= http.StatusInternalServerError{
			ext.Error.Set(sp, true)
		}
		sp.Finish()
	}
}