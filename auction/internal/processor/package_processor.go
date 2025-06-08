package processor

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/auction/internal/kafka"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/sirupsen/logrus"
)

type PackageProcessor struct {
	log      *logrus.Logger
	repo     repository.Packager
	producer *kafka.AuctionPublisher
}

func NewPackageProcessor(logger *logrus.Logger, repo repository.Packager, producer *kafka.AuctionPublisher) *PackageProcessor {
	return &PackageProcessor{
		log:      logger,
		repo:     repo,
		producer: producer,
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
			p.handleExpiredPackages(session, message)
		case "paid-packages":
			p.handlePaidPackage(session, message)
		}
	}
	return nil
}

func (p *PackageProcessor) handleExpiredPackages(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	ctx := session.Context()
	var pkg models.Package
	if err := json.Unmarshal(msg.Value, &pkg); err != nil {
		p.log.WithError(err).Error("Failed to decode package event")
		return
	}

	pkg.Status = "Waiting"
	pkg.UserID = ""
	pkg.UpdatedAt = time.Now()

	savedPkg, err := p.repo.Create(ctx, &pkg)
	if err != nil {
		p.log.WithError(err).Error("Failed to save package to repository")
		return
	}

	p.log.WithField("package_id", savedPkg.PackageID).Info("Package saved for auction")
	session.MarkMessage(msg, "")
}

func (p *PackageProcessor) handlePaidPackage(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	ctx := session.Context()
	var event models.PaidPackageEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		p.log.WithError(err).Error("Failed to decode paid package event")
		return
	}

	if event.Status != "PAID" {
		p.log.Warn("Skipping non-paid status package")
		return
	}

	pkg, err := p.repo.FindByID(ctx, event.PackageID)
	if err != nil {
		p.log.WithError(err).Errorf("Failed to find package %s", event.PackageID)
		return
	}
	pkg.Status = event.Status
	if err := p.repo.Update(ctx, pkg); err != nil {
		p.log.WithError(err).Errorf("Failed to update package %s", pkg.PackageID)
		return
	}

	init := &models.DeliveryInit{
		PackageID: pkg.PackageID,
		UserID:    event.UserID,
		From:      pkg.From,
		Address:   pkg.Address,
		Cost:      pkg.Cost,
		Weight:    pkg.Weight,
		Width:     pkg.Width,
		Length:    pkg.Length,
		Height:    pkg.Height,
		Currency:  pkg.Currency,
	}

	if err := p.producer.PublishDeliveryInit(ctx, init); err != nil {
		p.log.WithError(err).Error("Failed to publish delivery init")
	} else {
		p.log.WithField("package_id", pkg.PackageID).Info("Delivery init published to register-delivery")
	}
	session.MarkMessage(msg, "")
}
