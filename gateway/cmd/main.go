package main

import (
	"net/http"
	"time"

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

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
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
		{
			Prefix:      "/api/my/packages",
			TargetURL:   "http://localhost:8333",
			PathRewrite: "/my/packages",
		},
	}

	publicHandler := handlers.NewRouter(publicRoutes, logger)
	publicChain := middleware.NewLogMiddleware(
		enableCORS(publicHandler),
		logger,
		httpRequestsTotal,
		httpResponseTimeSeconds,
	)

	protectedHandler := handlers.NewRouter(protectedRoutes, logger)
	authProtected := middleware.NewAuthMiddleware(
		enableCORS(protectedHandler),
		logger,
	)
	fullProtectedChain := middleware.NewLogMiddleware(
		authProtected,
		logger,
		httpRequestsTotal,
		httpResponseTimeSeconds,
	)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/api/register", publicChain)
	http.Handle("/api/login", publicChain)
	http.Handle("/api/", fullProtectedChain)

	logger.Info("Starting API Gateway on :8228")
	server := &http.Server{
		Addr:              ":8228",
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
