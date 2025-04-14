package main

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/handlers"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/sirupsen/logrus"
)

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

	mainHandler := handlers.NewRouter(routes, logger)
	chain := middleware.NewLogMiddleware(mainHandler, logger)

	http.Handle("/", chain)

	logger.Info("Starting API Gateway on :8228")
	if err := http.ListenAndServe(":8228", nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
