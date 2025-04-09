package main

import (
	"fmt"
	"net/http"

	"github.com/maksroxx/DeliveryService/producer/configs"
	"github.com/maksroxx/DeliveryService/producer/internal/delivery/kafka"
	"github.com/maksroxx/DeliveryService/producer/internal/handler"
	"github.com/maksroxx/DeliveryService/producer/internal/middleware"
	"github.com/maksroxx/DeliveryService/producer/internal/service"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	cfg := configs.Load()
	fmt.Println("Kafka topic:", cfg.Kafka.Topic)
	fmt.Println("Server port:", cfg.Server.Address)

	kafkaProducer, err := kafka.NewProducer(kafka.Config{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	})
	if err != nil {
		logger.Fatal(err)
	}
	defer kafkaProducer.Close()

	var (
		svc            = service.NewPackageService(kafkaProducer)
		packageHandler = handler.NewPackageHandler(svc)
		packageChain   = middleware.NewLogMiddleware(packageHandler, logger)
	)

	http.Handle("/packages", packageChain)

	logger.Infof("Starting API Gateway on %s", cfg.Server.Address)
	if err := http.ListenAndServe(cfg.Server.Address, nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
