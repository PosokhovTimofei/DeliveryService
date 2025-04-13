package main

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/handlers"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	packageHandler := handlers.NewPackageHandler(
		"http://localhost:8333",
		logger,
	)

	chain := middleware.NewLogMiddleware(packageHandler, logger)

	http.Handle("/api/packages/", chain)
	http.Handle("/api/packages", chain) // post

	logger.Info("Starting API Gateway on :8228")
	if err := http.ListenAndServe(":8228", nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
