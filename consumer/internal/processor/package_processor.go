package processor

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/consumer/types"
	"github.com/sirupsen/logrus"
)

type PackageProcessor struct {
	log    *logrus.Logger
	client *http.Client
}

func NewPackageProcessor(logger *logrus.Logger) *PackageProcessor {
	return &PackageProcessor{
		log:    logger,
		client: &http.Client{Timeout: 5 * time.Second},
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
		func(msg *sarama.ConsumerMessage) {
			var pkg types.Package
			if err := json.Unmarshal(msg.Value, &pkg); err != nil {
				p.log.WithError(err).Error("Error decoding message")
				return
			}

			pkg.Status = "PROCESSED"
			jsonData, err := json.Marshal(pkg)
			if err != nil {
				p.log.WithError(err).Error("Error marshaling package")
				return
			}

			resp, err := p.client.Post(
				"http://localhost:8333/packages",
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				p.log.WithError(err).Error("Failed to send package")
				return
			}

			defer func() {
				if err := resp.Body.Close(); err != nil {
					p.log.WithError(err).Warn("Failed to close response body")
				}
			}()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				session.MarkMessage(msg, "")
				p.log.WithField("package_id", pkg.ID).Info("Package processed successfully")
			} else {
				body, _ := io.ReadAll(resp.Body)
				p.log.WithFields(logrus.Fields{
					"status":   resp.StatusCode,
					"response": string(body),
				}).Error("Unexpected API response")
			}
		}(message)
	}
	return nil
}
