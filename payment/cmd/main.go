package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
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
	pgPool      *pgxpool.Pool
	repo        db.Paymenter
	producer    kafka.Producerer
	consumer    kafka.Consumerer
	httpServer  *http.Server
)

func main() {
	initializeLogger()
	loadConfig()

	switch strings.ToLower(cfg.Database.Driver) {
	case "postgres":
		connectPostgreSQL()
	case "mongo":
		connectMongoDB()
	default:
		logger.Fatalf("Unsupported database driver: %s", cfg.Database.Driver)
	}

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

func connectPostgreSQL() {
	cfg := cfg.Postgres
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	var err error
	pgPool, err = pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		logger.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	repo = db.NewPostgresPaymenter(pgPool)

	logger.Info("Connected to PostgreSQL")
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
			handler.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/payment/"), "/")
		if len(pathParts) == 0 || pathParts[0] == "" {
			handler.RespondError(w, http.StatusBadRequest, "Missing package ID")
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

	if mongoClient != nil {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			logger.Errorf("Error closing MongoDB connection: %v", err)
		}
	}

	if pgPool != nil {
		pgPool.Close()
	}

	logger.Info("Server gracefully stopped")
}
