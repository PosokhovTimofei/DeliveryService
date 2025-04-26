package processor

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/payment/internal/db"
	"github.com/maksroxx/DeliveryService/payment/internal/models"
	"github.com/sirupsen/logrus"
)

type PaymentProcessor struct {
	log  *logrus.Logger
	repo db.Paymenter
}

func NewPaymentProcessor(log *logrus.Logger, repo db.Paymenter) *PaymentProcessor {
	return &PaymentProcessor{
		log:  log,
		repo: repo,
	}
}

func (p *PaymentProcessor) Setup(sarama.ConsumerGroupSession) error {
	p.log.Info("Payment processor setup")
	return nil
}

func (p *PaymentProcessor) Cleanup(sarama.ConsumerGroupSession) error {
	p.log.Info("Payment processor cleanup")
	return nil
}

func (p *PaymentProcessor) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		func(msg *sarama.ConsumerMessage) {
			var payment models.Payment
			if err := json.Unmarshal(msg.Value, &payment); err != nil {
				p.log.WithError(err).Error("Failed to decode payment event")
				return
			}

			err := p.repo.CreatePayment(context.Background(), payment)
			if err != nil {
				p.log.WithError(err).Error("Failed to save payment to DB")
				return
			}

			session.MarkMessage(msg, "")
			p.log.WithFields(logrus.Fields{
				"user_id":    payment.UserID,
				"package_id": payment.PackageID,
			}).Info("Payment saved to database successfully")
		}(message)
	}
	return nil
}
