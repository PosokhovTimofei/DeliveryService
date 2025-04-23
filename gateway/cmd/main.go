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

	publicRoutes := []handlers.RouteConfig{
		{
			Prefix:      "/api/register",
			TargetURL:   "http://localhost:1703",
			PathRewrite: "/register",
		},
		{
			Prefix:      "/api/login",
			TargetURL:   "http://localhost:1703",
			PathRewrite: "/login",
		},
	}

	protectedRoutes := []handlers.RouteConfig{
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
		{
			Prefix:      "/api/profile",
			TargetURL:   "http://localhost:1704",
			PathRewrite: "/profile",
		},
	}

	publicHandler := handlers.NewRouter(publicRoutes, logger)
	protectedHandler := handlers.NewRouter(protectedRoutes, logger)

	authProtected := middleware.NewAuthMiddleware(protectedHandler, logger)
	fullProtectedChain := middleware.NewLogMiddleware(authProtected, logger, httpRequestsTotal, httpResponseTimeSeconds)

	publicChain := middleware.NewLogMiddleware(publicHandler, logger, httpRequestsTotal, httpResponseTimeSeconds)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/api/", fullProtectedChain)
	http.Handle("/api/register", publicChain)
	http.Handle("/api/login", publicChain)

	logger.Info("Starting API Gateway on :8228")
	if err := http.ListenAndServe(":8228", nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
