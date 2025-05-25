package main

import (
	"context"
	"log"
	"net/http"

	"github.com/maksroxx/DeliveryService/calculator/configs"
	"github.com/maksroxx/DeliveryService/calculator/internal/middleware"
	"github.com/maksroxx/DeliveryService/calculator/internal/repository"
	"github.com/maksroxx/DeliveryService/calculator/internal/service"
	"github.com/maksroxx/DeliveryService/calculator/internal/transport"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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
	repo := repository.NewCityMongoRepository(db, "countries")
	tariffRepo := repository.NewTariffMongoRepository(db, "tariffs")
	log := logrus.New()
	chain := middleware.NewChain(
		middleware.NewMetricsMiddleware(),
		middleware.NewLogMiddleware(log),
	)

	svc := service.NewExtendedCalculator(repo, tariffRepo)
	go func() {
		if err := transport.StartGRPCServer(cfg.GRPCPort, tariffRepo, *svc, log); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	startHTTPServer(cfg.HTTPPort, svc, chain, log, tariffRepo)
}

func startHTTPServer(port string, calc service.Calculator, chain *middleware.Chain, log *logrus.Logger, rep repository.TariffRepository) {
	handler := transport.NewHTTPHandler(calc, rep)

	http.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
		chain.Then(http.HandlerFunc(handler.HandleCalculate)).ServeHTTP(w, r)
	})
	http.HandleFunc("/calculate-by-tariff", func(w http.ResponseWriter, r *http.Request) {
		chain.Then(http.HandlerFunc(handler.HandleCalculateByTariff)).ServeHTTP(w, r)
	})
	http.HandleFunc("/tariffs", func(w http.ResponseWriter, r *http.Request) {
		chain.Then(http.HandlerFunc(handler.HandleTariffList)).ServeHTTP(w, r)
	})
	http.HandleFunc("/tariff", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			chain.Then(http.HandlerFunc(handler.CreateTariff)).ServeHTTP(w, r)
		case http.MethodDelete:
			chain.Then(http.HandlerFunc(handler.DeleteTariff)).ServeHTTP(w, r)
		}
	})
	log.Infof("HTTP server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
