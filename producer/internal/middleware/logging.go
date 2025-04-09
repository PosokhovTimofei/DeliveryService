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

	defer func() {
		m.logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"time":   time.Since(start).String(),
		}).Info("--- Request processed")
	}()

	m.next.ServeHTTP(w, r)
}
