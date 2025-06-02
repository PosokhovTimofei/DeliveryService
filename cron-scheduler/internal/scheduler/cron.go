package scheduler

import (
	"github.com/maksroxx/DeliveryService/cron-scheduler/internal/clients"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type cronDeps struct {
	Log     *logrus.Logger
	Clients interface {
		Close()
	}
}

func NewCronDeps(log *logrus.Logger, clients *clients.GRPCClients) *cronDeps {
	return &cronDeps{
		Log:     log,
		Clients: clients,
	}
}

func Start(deps cronDeps) {
	c := cron.New(cron.WithSeconds())

	addTransferExpiredPackages(c, deps)
	addStartAuctionFridays(c, deps)
	addRepeatAuctionSaturday(c, deps)

	c.Start()

	select {}
}
