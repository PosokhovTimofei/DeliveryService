package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/maksroxx/DeliveryService/auth/metrics"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next   http.Handler
	logger *logrus.Logger
}

func NewLogMiddleware(next http.Handler, logger *logrus.Logger) *LogMiddleware {
	return &LogMiddleware{
		next:   next,
		logger: logger,
	}
}

func (m *LogMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	lrw := &LoggingResponseWriter{ResponseWriter: w}

	defer func() {
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(lrw.Status)

		metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Inc()

		metrics.HTTPResponseTime.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Observe(duration)

		m.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": duration,
			"status":   status,
		}).Info("Request processed")
	}()

	m.next.ServeHTTP(lrw, r)
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.Status = code
	lrw.ResponseWriter.WriteHeader(code)
}
