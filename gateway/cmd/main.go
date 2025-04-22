package main

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/handlers"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpResponseTimeSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Response time of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpResponseTimeSeconds)
}

func main() {
	logger := logrus.New()
	routes := []handlers.RouteConfig{
		{
			Prefix:      "/api/packages",
			TargetURL:   "http://localhost:8333",
			PathRewrite: "/packages",
		},
		{
			Prefix:      "/api/calculate",
			TargetURL:   "http://localhost:8121",
			PathRewrite: "/calculate",
		},
		{
			Prefix:      "/api/create",
			TargetURL:   "http://localhost:1234",
			PathRewrite: "/producer",
		},
	}

	http.Handle("/metrics", promhttp.Handler())

	mainHandler := handlers.NewRouter(routes, logger)
	chain := middleware.NewLogMiddleware(mainHandler, logger, httpRequestsTotal, httpResponseTimeSeconds)

	http.Handle("/", chain)

	logger.Info("Starting API Gateway on :8228")
	if err := http.ListenAndServe(":8228", nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
