package main

import (
	"context"
	"log"
	"net/http"

	"github.com/maksroxx/DeliveryService/producer/configs"
	"github.com/maksroxx/DeliveryService/producer/internal/calculator"
	"github.com/maksroxx/DeliveryService/producer/internal/delivery/kafka"
	"github.com/maksroxx/DeliveryService/producer/internal/handler"
	"github.com/maksroxx/DeliveryService/producer/internal/middleware"
	"github.com/maksroxx/DeliveryService/producer/internal/repository"
	"github.com/maksroxx/DeliveryService/producer/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	logger := logrus.New()
	cfg := configs.Load()

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Database.Uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database(cfg.Database.Name)

	http.Handle("/metrics", promhttp.Handler())

	kafkaProducer, err := kafka.NewProducer(kafka.Config{
		Brokers:      cfg.Kafka.Brokers,
		PackageTopic: cfg.Kafka.PackageTopic,
		PaymentTopic: cfg.Kafka.PaymentTopic,
		Version:      cfg.Kafka.Version,
	})
	if err != nil {
		logger.Fatal(err)
	}
	defer kafkaProducer.Close()

	client := calculator.NewClient(cfg.Calculator.URL)
	repo := repository.NewPackageRepository(db, "producer")
	svc := service.NewPackageService(kafkaProducer, client, repo, logger)
	handler := handler.NewPackageHandler(svc)

	http.Handle("/producer", middleware.NewLogMiddleware(handler, logger))

	logger.Infof("Starting Producer service on %s", cfg.Server.Address)
	if err := http.ListenAndServe(cfg.Server.Address, nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
