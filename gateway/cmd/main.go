package main

import (
	"github.com/maksroxx/DeliveryService/gateway/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	server := app.NewServer(logger)

	logger.Info("Starting API Gateway on :8228")
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
