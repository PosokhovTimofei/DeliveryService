package processor

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/consumer/types"
	"github.com/sirupsen/logrus"
)

type PackageProcessor struct {
	log *logrus.Logger
}

func NewPackageProcessor(logger *logrus.Logger) *PackageProcessor {
	return &PackageProcessor{log: logger}
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
		p.log.WithFields(logrus.Fields{
			"topic":     message.Topic,
			"partition": message.Partition,
			"offset":    message.Offset,
		}).Debug("Received message")

		var pkg types.Package
		if err := json.Unmarshal(message.Value, &pkg); err != nil {
			p.log.WithFields(logrus.Fields{
				"error": err,
				"value": string(message.Value),
			}).Error("Error decoding message")
			continue
		}

		p.log.WithFields(logrus.Fields{
			"package_id": pkg.ID,
			"status":     pkg.Status,
			"weight":     pkg.Weight,
		}).Info("Processing package")

		pkg.Status = "PROCESSED"
		p.log.WithField("new_status", pkg.Status).
			Info("Package processed")

		session.MarkMessage(message, "")
	}
	return nil
}
