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
			var userID string
			for _, header := range msg.Headers {
				if string(header.Key) == "User-ID" {
					userID = string(header.Value)
					break
				}
			}

			if userID == "" {
				p.log.Warn("Missing User-ID header in Kafka message")
				return
			}

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

			req, err := http.NewRequest(
				"POST",
				"http://localhost:8333/packages",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				p.log.WithError(err).Error("Failed to create request")
				return
			}

			req.Header.Set("X-User-ID", userID)
			req.Header.Set("Content-Type", "application/json")

			resp, err := p.client.Do(req)
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
				p.log.WithFields(logrus.Fields{
					"package_id": pkg.ID,
					"user_id":    userID,
				}).Info("Package processed successfully")
			} else {
				body, _ := io.ReadAll(resp.Body)
				p.log.WithFields(logrus.Fields{
					"status":   resp.StatusCode,
					"response": string(body),
					"user_id":  userID,
				}).Error("Unexpected API response")
			}
		}(message)
	}
	return nil
}
