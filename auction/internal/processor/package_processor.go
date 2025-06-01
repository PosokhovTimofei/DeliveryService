package processor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/auction/internal/kafka"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/maksroxx/DeliveryService/auction/internal/service"
	"github.com/sirupsen/logrus"
)

type PackageProcessor struct {
	log        *logrus.Logger
	repo       repository.Packager
	auctionSvc *service.AuctionService
	publisher  *kafka.AuctionPublisher
}

func NewPackageProcessor(logger *logrus.Logger, repo repository.Packager, auctionSvc *service.AuctionService, publisher *kafka.AuctionPublisher) *PackageProcessor {
	return &PackageProcessor{
		log:        logger,
		repo:       repo,
		auctionSvc: auctionSvc,
		publisher:  publisher,
	}
}

func (p *PackageProcessor) Setup(sarama.ConsumerGroupSession) error {
	p.log.Info("Consumer group setup")
	return nil
}

func (p *PackageProcessor) Cleanup(sarama.ConsumerGroupSession) error {
	p.log.Info("Consumer group cleanup")
	return nil
}

func (p *PackageProcessor) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		switch string(message.Topic) {
		case "expired-packages":
			p.handleExpiredPackages(session.Context(), message)
		}
	}
	return nil
}

func (p *PackageProcessor) handleExpiredPackages(ctx context.Context, msg *sarama.ConsumerMessage) {
	var pkg models.Package
	if err := json.Unmarshal(msg.Value, &pkg); err != nil {
		p.log.WithError(err).Error("Failed to decode package event")
		return
	}

	pkg.Status = "Auctioning"
	pkg.UpdatedAt = time.Now()

	savedPkg, err := p.repo.Create(ctx, &pkg)
	if err != nil {
		p.log.WithError(err).Error("Failed to save package to repository")
		return
	}

	p.log.WithField("package_id", savedPkg.PackageID).Info("Package saved for auction")

	service.StartAuction(ctx, savedPkg, p.auctionSvc, p.publisher, p.repo, p.log)
}
