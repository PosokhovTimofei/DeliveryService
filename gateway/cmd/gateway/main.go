package main

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/handlers"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		logger         = logrus.New()
		packageHandler = handlers.NewPackageHandler("http://localhost:1234/packages", logger)
		statusHandler  = handlers.NewStatusHandler("http://localhost:2345/status", logger)
		packageChain   = middleware.NewLogMiddleware(
			packageHandler,
			logger,
		)
	)

	http.Handle("/api/packages", packageChain)
	http.Handle("/api/status/", statusHandler)

	logger.Info("Starting API Gateway on :8228")
	if err := http.ListenAndServe(":8228", nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
