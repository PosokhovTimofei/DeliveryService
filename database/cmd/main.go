package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/maksroxx/DeliveryService/database/internal/configs"
	"github.com/maksroxx/DeliveryService/database/internal/handlers"
	"github.com/maksroxx/DeliveryService/database/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := configs.Load()

	// MongoDB connection
	mongoCfg := cfg.Database.MongoDB
	clientOptions := options.Client().ApplyURI(mongoCfg.URI)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	// MongoDB ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}

	// Database setup
	db := client.Database(mongoCfg.Database)
	repo := repository.NewMongoRepository(db, "packages")

	// HTTP server setup
	packageHandler := handlers.NewPackageHandler(repo)
	router := mux.NewRouter()
	packageHandler.RegisterRoutes(router)

	// Добавлено логирование запуска
	serverAddr := cfg.Server.Address
	log.Printf("🚀 Server starting on %s", serverAddr)

	// Добавлена обработка ошибок запуска сервера
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
