package middle_ware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
)

func GinMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		requestTotal.With(prometheus.Labels{"method": c.Request.Method,
		"endpoint": c.Request.Host}).Inc()
		requestDuration.With(prometheus.Labels{"method": c.Request.Method,
			"endpoint": c.Request.Host}).Observe(duration.Seconds())
	}
}

func GinTracer(tracer opentracing2.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		spContext, _ := tracer.Extract(opentracing2.HTTPHeaders,
			opentracing2.HTTPHeadersCarrier(c.Request.Header))

		sp := tracer.StartSpan("HTTP "+c.Request.Method,
			ext.RPCServerOption(spContext))

		ext.HTTPMethod.Set(sp, c.Request.Method)
		ext.HTTPUrl.Set(sp, c.Request.URL.String())
		ext.Component.Set(sp, "net/http")

		c.Request = c.Request.WithContext(opentracing2.ContextWithSpan(c.Request.Context(), sp))
		c.Next()

		ext.HTTPStatusCode.Set(sp, uint16(c.Writer.Status()))
		if c.Writer.Status() >= http.StatusInternalServerError {
			ext.Error.Set(sp, true)
		}
		sp.Finish()
	}
}
