package main

import (
	"context"
	"log"

	"github.com/maksroxx/DeliveryService/telegram/configs"
	"github.com/maksroxx/DeliveryService/telegram/internal/bot"
	"github.com/maksroxx/DeliveryService/telegram/internal/clients"
	"github.com/maksroxx/DeliveryService/telegram/internal/handlers"
	"github.com/maksroxx/DeliveryService/telegram/internal/repository"
	"github.com/maksroxx/DeliveryService/telegram/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := configs.Load()

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.Database.MongoDB.URI))
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к MongoDB: %v", err)
	}
	db := mongoClient.Database(cfg.Database.MongoDB.Database)

	authClient, err := clients.NewAuthClient(cfg.GrpcConfig.Auth)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к Auth GRPC: %v", err)
	}
	defer authClient.Close()

	packageClient, err := clients.NewPackageClient(cfg.GrpcConfig.Package)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к Package GRPC: %v", err)
	}
	defer packageClient.Close()

	repo := repository.NewUserLinkRepository(db, "telegram_user_links")
	authService := service.NewAuthService(repo, authClient)
	packageService := service.NewPackageService(repo, packageClient)
	handler := handlers.NewHandler(authService, packageService)
	botAPI := bot.NewBot(cfg.Telegram.TelegramToken)
	botAPI.Run(handler)
}
