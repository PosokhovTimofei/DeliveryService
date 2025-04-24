package middleware

import (
	"net/http"
	"time"

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

	lrw := &loggingResponseWriter{ResponseWriter: w}

	m.next.ServeHTTP(lrw, r)

	duration := time.Since(start)
	m.logger.WithFields(logrus.Fields{
		"status":   lrw.status,
		"duration": duration,
		"method":   r.Method,
		"path":     r.URL.Path,
	}).Info("Request completed")
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}
