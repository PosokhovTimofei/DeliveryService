package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/maksroxx/DeliveryService/telegram/configs"
	"github.com/maksroxx/DeliveryService/telegram/internal/bot"
	"github.com/maksroxx/DeliveryService/telegram/internal/clients"
	"github.com/maksroxx/DeliveryService/telegram/internal/handlers"
	"github.com/maksroxx/DeliveryService/telegram/internal/kafka"
	"github.com/maksroxx/DeliveryService/telegram/internal/processor"
	"github.com/maksroxx/DeliveryService/telegram/internal/repository"
	"github.com/maksroxx/DeliveryService/telegram/internal/service"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := configs.Load()
	log := logrus.New()
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.Database.MongoDB.URI))
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}
	db := mongoClient.Database(cfg.Database.MongoDB.Database)

	authClient, err := clients.NewAuthClient(cfg.GrpcConfig.Auth)
	if err != nil {
		log.Fatalf("Auth gRPC connection error: %v", err)
	}
	defer authClient.Close()

	packageClient, err := clients.NewPackageClient(cfg.GrpcConfig.Package)
	if err != nil {
		log.Fatalf("Package gRPC connection error: %v", err)
	}
	defer packageClient.Close()

	repo := repository.NewUserLinkRepository(db, "telegram_user_links")
	authService := service.NewAuthService(repo, authClient)
	packageService := service.NewPackageService(repo, packageClient)
	handler := handlers.NewHandler(authService, packageService)
	botAPI := bot.NewBot(cfg.Telegram.TelegramToken)

	notificationProcessor := processor.NewNotificationProcessor(log, repo, botAPI.API)
	consumer, err := kafka.NewConsumer(kafka.ConfigConsumer{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic[0],
		GroupID: cfg.Kafka.GroupID,
	}, notificationProcessor, log)
	if err != nil {
		log.Fatalf("Kafka consumer init error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go consumer.Run(ctx)
	go botAPI.Run(handler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("Shutting down Telegram service...")

	cancel()
	if err := consumer.Close(); err != nil {
		log.Errorf("Error while closing consumer: %v", err)
	}
}
