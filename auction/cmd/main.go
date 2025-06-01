package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksroxx/DeliveryService/auction/configs"
	"github.com/maksroxx/DeliveryService/auction/internal/kafka"
	"github.com/maksroxx/DeliveryService/auction/internal/processor"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/maksroxx/DeliveryService/auction/internal/service"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log := logrus.New()
	cfg := configs.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleShutdown(cancel, cfg.Server.ShutdownTimeout, log)

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.Database.URI))
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to MongoDB")
	}
	db := mongoClient.Database(cfg.Database.Database.Database)
	packageRepo := repository.NewPackageRepository(db, "auctioned")
	bidRepo := repository.NewBidRepository(db, "bids")
	auctionService := service.NewAuctionService(bidRepo)

	producer, err := kafka.NewAuctionPublisher(cfg.Kafka.Brokers, cfg.Kafka.ProduceTopic[0], log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create Kafka publisher")
	}
	defer producer.Close()

	processor := processor.NewPackageProcessor(log, packageRepo, auctionService, producer)

	consumer, err := kafka.NewConsumer(kafka.Config{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.ConsumeTopic,
		GroupID: cfg.Kafka.GroupID,
	}, processor, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	log.Info("Starting Kafka consumer...")
	consumer.Run(ctx)
	log.Info("Kafka consumer stopped.")
}

func handleShutdown(cancel context.CancelFunc, timeout time.Duration, log *logrus.Logger) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	log.Warn("Shutdown signal received")
	ctx, cancelTimeout := context.WithTimeout(context.Background(), timeout)
	defer cancelTimeout()

	cancel()
	<-ctx.Done()
	log.Info("Shutdown complete")
}
