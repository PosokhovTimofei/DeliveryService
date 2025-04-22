package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next          http.Handler
	logger        *logrus.Logger
	requestsTotal *prometheus.CounterVec
	responseTime  *prometheus.HistogramVec
}

func NewLogMiddleware(
	next http.Handler,
	logger *logrus.Logger,
	requestsTotal *prometheus.CounterVec,
	responseTime *prometheus.HistogramVec,
) *LogMiddleware {
	return &LogMiddleware{
		next:          next,
		logger:        logger,
		requestsTotal: requestsTotal,
		responseTime:  responseTime,
	}
}

func (m *LogMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	lrw := utils.NewLoggingResponseWriter(w)

	defer func() {
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(lrw.StatusCode)

		m.requestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			statusCode,
		).Inc()

		m.responseTime.WithLabelValues(
			r.Method,
			r.URL.Path,
			statusCode,
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
