package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/maksroxx/DeliveryService/gateway/internal/metrics"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := utils.NewLoggingResponseWriter(w, http.StatusOK)
		next.ServeHTTP(rec, r)
		duration := time.Since(start).Seconds()

		metrics.HttpRequestsTotal.
			WithLabelValues(r.Method, r.URL.Path, fmt.Sprint(rec.StatusCode)).
			Inc()
		metrics.HttpResponseTimeSeconds.
			WithLabelValues(r.Method, r.URL.Path, fmt.Sprint(rec.StatusCode)).
			Observe(duration)
	})
}
