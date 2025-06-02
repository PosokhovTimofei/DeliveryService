package processor

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/sirupsen/logrus"
)

type PackageProcessor struct {
	log  *logrus.Logger
	repo repository.Packager
}

func NewPackageProcessor(logger *logrus.Logger, repo repository.Packager) *PackageProcessor {
	return &PackageProcessor{
		log:  logger,
		repo: repo,
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

	pkg.Status = "Auctioning"
	pkg.UpdatedAt = time.Now()

	savedPkg, err := p.repo.Create(ctx, &pkg)
	if err != nil {
		p.log.WithError(err).Error("Failed to save package to repository")
		return
	}

	p.log.WithField("package_id", savedPkg.PackageID).Info("Package saved for auction")
	session.MarkMessage(msg, "")
}
