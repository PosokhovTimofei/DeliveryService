package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/maksroxx/DeliveryService/payment/configs"
	"github.com/maksroxx/DeliveryService/payment/internal/db"
	"github.com/maksroxx/DeliveryService/payment/internal/handler"
	"github.com/maksroxx/DeliveryService/payment/internal/kafka"
	"github.com/maksroxx/DeliveryService/payment/internal/processor"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	logger      = logrus.New()
	cfg         *configs.Config
	mongoClient *mongo.Client
	repo        db.Paymenter
	producer    *kafka.Producer
	consumer    *kafka.Consumer
	httpServer  *http.Server
)

func main() {
	initializeLogger()
	loadConfig()
	connectMongoDB()
	setupKafka()
	startConsumer()
	setupHTTPServer()

	gracefulShutdown()
}

func initializeLogger() {
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
}

func loadConfig() {
	cfg = configs.Load()
}

func connectMongoDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.Uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	mongoClient = client
	dbInstance := mongoClient.Database(cfg.Database.Name)
	repo = db.NewPaymentMongoRepository(dbInstance, "payments")
}

func setupKafka() {
	producerCfg := kafka.ConfigProducer{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.ProducerTopic,
		Version: cfg.Kafka.Version,
	}
	var err error
	producer, err = kafka.NewProducer(producerCfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	consumerCfg := kafka.ConfigConsumer{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.ConsumerTopic,
		GroupID: cfg.Kafka.GroupID,
	}

	processor := processor.NewPaymentProcessor(logger, repo)
	consumer, err = kafka.NewConsumer(consumerCfg, processor, logger)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
}

func startConsumer() {
	go func() {
		logger.Info("Starting Kafka consumer...")
		consumer.Run(context.Background())
	}()
}

func setupHTTPServer() {
	paymentHandler := handler.NewPaymentHandler(repo, producer)

	http.HandleFunc("/payment/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/payment/"), "/")
		if len(pathParts) == 0 || pathParts[0] == "" {
			http.Error(w, "Missing package ID", http.StatusBadRequest)
			return
		}

		r.URL.Path = "/payment/" + pathParts[0]
		paymentHandler.ServeHTTP(w, r)
	})

	httpServer = &http.Server{
		Addr:    cfg.Server.Address,
		Handler: http.DefaultServeMux,
	}

	go func() {
		logger.Infof("HTTP server started at %s", cfg.Server.Address)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()
}

func gracefulShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		logger.Fatalf("HTTP shutdown error: %v", err)
	}

	if err := consumer.Close(); err != nil {
		logger.Errorf("Error closing Kafka consumer: %v", err)
	}

	if err := producer.Close(); err != nil {
		logger.Errorf("Error closing Kafka producer: %v", err)
	}

	if err := mongoClient.Disconnect(context.Background()); err != nil {
		logger.Errorf("Error closing MongoDB connection: %v", err)
	}

	logger.Info("Server gracefully stopped")
}
