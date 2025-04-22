package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/maksroxx/DeliveryService/producer/internal/metrics"
	"github.com/sirupsen/logrus"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

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
	lrw := &LoggingResponseWriter{w, http.StatusOK}

	defer func() {
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(lrw.StatusCode)

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
			"method": r.Method,
			"path":   r.URL.Path,
			"time":   time.Since(start).String(),
			"status": lrw.StatusCode,
		}).Info("--- Request processed")
	}()

	m.next.ServeHTTP(lrw, r)
}
