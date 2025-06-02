package scheduler

import (
	"context"
	"time"

	"github.com/maksroxx/DeliveryService/cron-scheduler/internal/clients"
	auctionpb "github.com/maksroxx/DeliveryService/proto/auction"
	databasepb "github.com/maksroxx/DeliveryService/proto/database"
	"github.com/robfig/cron/v3"
)

func addTransferExpiredPackages(c *cron.Cron, deps cronDeps) {
	c.AddFunc("0 0 1 * * *", func() {
		log := deps.Log.WithField("job", "TransferExpiredPackages")
		log.Info("Started")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		_, err := deps.Clients.(*clients.GRPCClients).Packages.TransferExpiredPackages(ctx, &databasepb.Empty{})
		if err != nil {
			log.WithError(err).Error("Failed")
		} else {
			log.Info("Completed")
		}
	})
}

func addStartAuctionFridays(c *cron.Cron, deps cronDeps) {
	c.AddFunc("0 0 19 * * 5", func() {
		log := deps.Log.WithField("job", "StartAuction")
		log.Info("Started")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		_, err := deps.Clients.(*clients.GRPCClients).Auction.StartAuction(ctx, &auctionpb.Empty{})
		if err != nil {
			log.WithError(err).Error("Failed")
		} else {
			log.Info("Completed")
		}
	})
}

func addRepeatAuctionSaturday(c *cron.Cron, deps cronDeps) {
	c.AddFunc("0 0 19 * * 6", func() {
		log := deps.Log.WithField("job", "RepeatAuction")
		log.Info("Started")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		_, err := deps.Clients.(*clients.GRPCClients).Auction.RepeateAuction(ctx, &auctionpb.Empty{})
		if err != nil {
			log.WithError(err).Error("Failed")
		} else {
			log.Info("Completed")
		}
	})
}
