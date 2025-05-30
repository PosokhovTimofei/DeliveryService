package main

import (
	"time"

	"github.com/maksroxx/DeliveryService/cron-transfer/internal/config"
	"github.com/maksroxx/DeliveryService/cron-transfer/internal/job"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	cfg := config.Load()

	conn, err := grpc.NewClient(
		cfg.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 5 * time.Second}),
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to gRPC server")
	}
	defer conn.Close()

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	transfer := job.NewTransferJob(conn, log)

	log.Infof("Cron expired transfer started, interval: %v", cfg.Interval)

	transfer.Run()

	for range ticker.C {
		transfer.Run()
	}
}
