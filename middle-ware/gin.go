package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jukylin/esim/log"
	tracerid "github.com/jukylin/esim/pkg/tracer-id"
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

// GinTracerId If not found opentracing's tracer_id then generate a new tracer_id.
func GinTracerID() gin.HandlerFunc {
	tracerID := tracerid.TracerID()
	return func(c *gin.Context) {
		if tracerid.ExtractTracerID(c.Request.Context()) == "" {
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(),
				tracerid.ActiveEsimKey, tracerID()))
		}

		c.Next()
	}
}

// GinLogFormatter is the log format function Logger middleware uses.
var GinLogFormatter = func(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}
	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v %s \n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		tracerid.ExtractTracerID(param.Request.Context()),
		param.ErrorMessage,
	)
}

// GinRecovery returns a middleware for a given logger that recovers from any panics.
func GinRecovery(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if logger != nil {
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if current[0] == "Authorization" {
							headers[idx] = current[0] + ": *"
						}
					}

					logger.Errorc(c.Request.Context(), "[Recovery] panic recovered:\n%s\n [%s]",
						strings.Join(headers, "\r\n"), err)
				}

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}
}
