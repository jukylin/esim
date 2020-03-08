package middle_ware

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

func Monitor(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start)
		requestTotal.With(prometheus.Labels{"method": r.Method, "endpoint": r.URL.Path}).Inc()
		requestDuration.With(prometheus.Labels{"method": r.Method, "endpoint": r.URL.Path}).Observe(duration.Seconds())
	}
}
