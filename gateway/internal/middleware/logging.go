package middleware

import (
	"net/http"
	"time"

	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
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
	lrw := utils.NewLoggingResponseWriter(w)

	defer func() {
		m.logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"time":   time.Since(start).String(),
			"status": lrw.StatusCode,
		}).Info("--- Request processed")
	}()

	m.next.ServeHTTP(lrw, r)
}
