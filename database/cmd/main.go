package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/maksroxx/DeliveryService/database/configs"
	"github.com/maksroxx/DeliveryService/database/internal/clients"
	"github.com/maksroxx/DeliveryService/database/internal/handlers"
	"github.com/maksroxx/DeliveryService/database/internal/kafka"
	"github.com/maksroxx/DeliveryService/database/internal/middleware"
	"github.com/maksroxx/DeliveryService/database/internal/processor"
	"github.com/maksroxx/DeliveryService/database/internal/repository"
	"github.com/maksroxx/DeliveryService/database/internal/service"
	pb "github.com/maksroxx/DeliveryService/proto/database"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	logger := logrus.New()
	cfg := configs.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoCfg := cfg.Database.MongoDB
	clientOptions := options.Client().ApplyURI(mongoCfg.URI)

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Errorf("Error disconnecting MongoDB: %v", err)
		}
	}()
	db := mongoClient.Database(mongoCfg.Database)

	calcClient, err := clients.NewCalculatorClient(cfg.Calculator.GRPCAddress)
	if err != nil {
		logger.Fatal("Failed to connect to calculator:", err)
	}
	defer calcClient.Close()
	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic[1:])
	if err != nil {
		logger.Fatal("Failed to init Kafka producer:", err)
	}
	defer producer.Close()
	repo := repository.NewMongoRepository(db, "packages")
	service := service.NewPackageService(repo, calcClient, producer, logger)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCAuthInterceptor()),
	)
	pb.RegisterPackageServiceServer(grpcServer, handlers.NewGrpcPackageHandler(service, logger))

	go func() {
		listener, err := net.Listen("tcp", ":50054")
		if err != nil {
			logger.Fatalf("Failed to start gRPC server: %v", err)
		}
		logger.Infof("gRPC server listening on :50054")
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatalf("gRPC server failed: %v", err)
		}
	}()
	packageHandler := handlers.NewPackageHandler(repo, calcClient, producer, logger)

	mux := http.NewServeMux()
	protected := http.NewServeMux()
	packageHandler.RegisterDefaultRoutes(mux)
	packageHandler.RegisterUserRoutes(protected)
	protectedHandler := middleware.AuthMiddleware(protected)

	mux.Handle("/", protectedHandler)
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: mux,
	}

	kafkaCfg := kafka.Config{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic[1:],
		GroupID: cfg.Kafka.GroupID,
	}
	processor := processor.NewPackageProcessor(logger, repo)
	consumer, err := kafka.NewConsumer(kafkaCfg, processor, logger)
	if err != nil {
		logger.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	go consumer.Run(ctx)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		logger.Info("Received shutdown signal, shutting down gracefully...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("HTTP server shutdown error: %v", err)
		}
		if err := consumer.Close(); err != nil {
			logger.Errorf("Error closing Kafka consumer: %v", err)
		}
		grpcServer.GracefulStop()
	}()

	logger.Infof("Server starting on %s", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server error: %v", err)
	}

	logger.Info("Server gracefully stopped")
}
