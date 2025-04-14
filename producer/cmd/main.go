package main

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/producer/configs"
	"github.com/maksroxx/DeliveryService/producer/internal/calculator"
	"github.com/maksroxx/DeliveryService/producer/internal/delivery/kafka"
	"github.com/maksroxx/DeliveryService/producer/internal/handler"
	"github.com/maksroxx/DeliveryService/producer/internal/middleware"
	"github.com/maksroxx/DeliveryService/producer/internal/service"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	cfg := configs.Load()

	kafkaProducer, err := kafka.NewProducer(kafka.Config{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	})
	if err != nil {
		logger.Fatal(err)
	}
	defer kafkaProducer.Close()

	var (
		client         = calculator.NewClient(cfg.Calculator.URL)
		svc            = service.NewPackageService(kafkaProducer, client)
		packageHandler = handler.NewPackageHandler(svc)
		packageChain   = middleware.NewLogMiddleware(packageHandler, logger)
	)

	http.Handle("/producer", packageChain)

	logger.Infof("Starting API Gateway on %s", cfg.Server.Address)
	if err := http.ListenAndServe(cfg.Server.Address, nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
