package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksroxx/DeliveryService/auction/configs"
	"github.com/maksroxx/DeliveryService/auction/internal/handlers"
	"github.com/maksroxx/DeliveryService/auction/internal/kafka"
	"github.com/maksroxx/DeliveryService/auction/internal/middleware"
	"github.com/maksroxx/DeliveryService/auction/internal/processor"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/maksroxx/DeliveryService/auction/internal/service"
	auctionpb "github.com/maksroxx/DeliveryService/proto/auction"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	cfg := configs.Load()
	if err := configs.LoadAuthConfig(); err != nil {
		log.WithError(err).Fatal("Failed to load auth config")
	}
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

	producer, err := kafka.NewAuctionPublisher(cfg.Kafka.Brokers, cfg.Kafka.ProduceTopic, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create Kafka publisher")
	}
	defer producer.Close()

	processor := processor.NewPackageProcessor(log, packageRepo, producer)

	consumer, err := kafka.NewConsumer(kafka.Config{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.ConsumeTopic,
		GroupID: cfg.Kafka.GroupID,
	}, processor, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	bidHandler := handlers.NewBidGRPCHandler(bidRepo, packageRepo, auctionService, producer, log)
	go func() {
		lis, err := net.Listen("tcp", cfg.Server.GRPCAddress)
		if err != nil {
			log.WithError(err).Fatal("Failed to listen for gRPC")
		}

		grpcServer := grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				middleware.AuthInterceptor(),
			),
			grpc.ChainStreamInterceptor(
				middleware.StreamAuthInterceptor(),
			),
		)
		auctionpb.RegisterAuctionServiceServer(grpcServer, bidHandler)

		log.Infof("gRPC server started on %s", cfg.Server.GRPCAddress)
		if err := grpcServer.Serve(lis); err != nil {
			log.WithError(err).Fatal("Failed to serve gRPC")
		}
	}()

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
