package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/maksroxx/DeliveryService/gateway/internal/metrics"
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
) *LogMiddleware {
	return &LogMiddleware{
		next:          next,
		logger:        logger,
		requestsTotal: metrics.HttpRequestsTotal,
		responseTime:  metrics.HttpResponseTimeSeconds,
	}
}

func (m *LogMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	lrw := utils.NewLoggingResponseWriter(w, http.StatusOK)

	defer func() {
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(lrw.StatusCode)

		scheme := "http://"
		if r.TLS != nil {
			scheme = "https://"
		}
		fullURL := scheme + r.Host + r.URL.RequestURI()

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
			"link":   fullURL,
		}).Info(fullURL)
	}()

	m.next.ServeHTTP(lrw, r)
}
