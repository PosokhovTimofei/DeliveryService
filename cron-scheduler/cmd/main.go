package main

import (
	"github.com/maksroxx/DeliveryService/cron-scheduler/internal/clients"
	"github.com/maksroxx/DeliveryService/cron-scheduler/internal/scheduler"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	clients, err := clients.InitGRPCClients()
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize gRPC clients")
	}
	defer clients.Close()
	cronDeps := scheduler.NewCronDeps(log, clients)
	scheduler.Start(*cronDeps)
}
