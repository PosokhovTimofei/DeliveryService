package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maksroxx/DeliveryService/database/internal/configs"
	"github.com/maksroxx/DeliveryService/database/internal/handlers"
	"github.com/maksroxx/DeliveryService/database/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := configs.Load()

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

	db := client.Database(mongoCfg.Database)
	repo := repository.NewMongoRepository(db, "packages")

	packageHandler := handlers.NewPackageHandler(repo)
	router := mux.NewRouter()
	packageHandler.RegisterRoutes(router)

	log.Printf("Server starting on %s", cfg.Server.Address)

	if err := http.ListenAndServe(cfg.Server.Address, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
