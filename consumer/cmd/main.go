package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/maksroxx/DeliveryService/consumer/configs"
	"github.com/maksroxx/DeliveryService/consumer/internal/calculator"
	"github.com/maksroxx/DeliveryService/consumer/internal/kafka"
	"github.com/maksroxx/DeliveryService/consumer/internal/processor"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		logger = logrus.New()
		cfg    = configs.LoadConfig()
		config = kafka.Config{
			Brokers: cfg.Kafka.Brokers,
			Topic:   cfg.Kafka.Topic,
			GroupID: cfg.Kafka.GroupID,
		}
		calculator, err = calculator.NewClient(calculator.ClientType(cfg.Calculator.ClientType), cfg.Calculator.Address)
		// groupHandler
		processor = processor.NewPackageProcessor(logger, calculator)
	)
	consumer, err := kafka.NewConsumer(config, processor, logger)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGALRM)

	go func() {
		<-signals
		logger.Info("Receive shutdown signal")
		cancel()
		if err := consumer.Close(); err != nil {
			logger.Errorf("Error closing consumer: %v", err)
		}
	}()

	logger.Info("Starting consumer...")
	consumer.Run(ctx)
	logger.Info("Consumer stopped")
}
