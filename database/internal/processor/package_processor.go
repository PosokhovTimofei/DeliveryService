package processor

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/database/internal/models"
	"github.com/maksroxx/DeliveryService/database/internal/repository"
	"github.com/sirupsen/logrus"
)

type PackageProcessor struct {
	log  *logrus.Logger
	repo repository.RouteRepository
}

func NewPackageProcessor(logger *logrus.Logger, repo repository.RouteRepository) *PackageProcessor {
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
		case "package-events":
			p.handlePackageEvent(session, message)
		case "pay-events":
			p.handlePayEvent(session, message)
		default:
			p.log.Warnf("Unknown topic: %s", message.Topic)
		}
	}
	return nil
}

func (p *PackageProcessor) handlePackageEvent(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	userID := extractUserID(msg.Headers)
	if userID == "" {
		p.log.Warn("Missing User-ID header in package message")
		return
	}

	var pack models.Package
	if err := json.Unmarshal(msg.Value, &pack); err != nil {
		p.log.WithError(err).Error("Failed to decode package event")
		return
	}

	pack.UserID = userID
	pack.Status = "PROCESSED"

	if _, err := p.repo.Create(session.Context(), &pack); err != nil {
		p.log.WithError(err).Error("Failed to create package in DB")
		return
	}

	session.MarkMessage(msg, "")
	p.log.WithFields(logrus.Fields{
		"package_id": pack.PackageID,
		"user_id":    userID,
	}).Info("Package created successfully (DB)")
}

func (p *PackageProcessor) handlePayEvent(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	userID := extractUserID(msg.Headers)
	if userID == "" {
		p.log.Warn("Missing User-ID header in payment message")
		return
	}

	var payment models.Payment
	if err := json.Unmarshal(msg.Value, &payment); err != nil {
		p.log.WithError(err).Error("Failed to decode payment event")
		return
	}

	if payment.Status != "PAID" {
		p.log.WithFields(logrus.Fields{
			"user_id":        payment.UserID,
			"package_id":     payment.PackageID,
			"payment_status": payment.Status,
		}).Warn("Payment not completed, skipping package update")
		return
	}

	update := models.PackageUpdate{
		PaymentStatus: "PAID",
	}

	if _, err := p.repo.UpdateRoute(session.Context(), payment.PackageID, update); err != nil {
		p.log.WithError(err).Error("Failed to update package in DB")
		return
	}

	session.MarkMessage(msg, "")
	p.log.WithFields(logrus.Fields{
		"package_id": payment.PackageID,
		"user_id":    userID,
	}).Info("Package updated successfully after payment (DB)")
}

func extractUserID(headers []*sarama.RecordHeader) string {
	for _, header := range headers {
		if string(header.Key) == "User-ID" {
			return string(header.Value)
		}
	}
	return ""
}
