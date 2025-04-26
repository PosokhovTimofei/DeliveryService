package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
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

const (
	baseAPIURL = "http://localhost:8333"
)

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
			switch string(msg.Topic) {
			case "package-events":
				p.handlePackageEvent(session, msg)
			case "pay-events":
				p.handlePayEvent(session, msg)
			default:
				p.log.Warnf("Unknown topic: %s", msg.Topic)
			}
		}(message)
	}
	return nil
}

func (p *PackageProcessor) handlePackageEvent(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	userID := extractUserID(msg.Headers)
	if userID == "" {
		p.log.Warn("Missing User-ID header in package message")
		return
	}

	var pack types.Package
	if err := json.Unmarshal(msg.Value, &pack); err != nil {
		p.log.WithError(err).Error("Failed to decode package event")
		return
	}

	pack.Status = "PROCESSED"

	if err := p.sendCreatePackage(pack, userID); err != nil {
		p.log.WithError(err).Error("Failed to send create package request")
		return
	}

	session.MarkMessage(msg, "")
	p.log.WithFields(logrus.Fields{
		"package_id": pack,
		"user_id":    userID,
	}).Info("Package created successfully")
}

func (p *PackageProcessor) sendCreatePackage(pkg types.Package, userID string) error {
	jsonData, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/packages", baseAPIURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	req.Header.Set("X-User-ID", userID)
	req.Header.Set("Content-Type", "application/json")

	return p.doRequest(req)
}

func (p *PackageProcessor) handlePayEvent(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	userID := extractUserID(msg.Headers)
	if userID == "" {
		p.log.Warn("Missing User-ID header in payment message")
		return
	}

	var payment types.Payment
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

	if err := p.sendUpdatePackage(payment.PackageID, userID); err != nil {
		p.log.WithError(err).Error("Failed to send update package request")
		return
	}

	session.MarkMessage(msg, "")
	p.log.WithFields(logrus.Fields{
		"package_id": payment.PackageID,
		"user_id":    userID,
	}).Info("Package updated successfully after payment")
}

func (p *PackageProcessor) sendUpdatePackage(packageID, userID string) error {
	payload := map[string]string{
		"payment_status": "PAID",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/packages/%s", baseAPIURL, packageID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	req.Header.Set("X-User-ID", userID)
	req.Header.Set("Content-Type", "application/json")

	return p.doRequest(req)
}

func (p *PackageProcessor) doRequest(req *http.Request) error {
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			p.log.WithError(cerr).Warn("Failed to close response body")
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	p.log.WithFields(logrus.Fields{
		"status":   resp.StatusCode,
		"response": string(body),
	}).Error("Unexpected API response")

	return fmt.Errorf("non-2xx response received")
}

func extractUserID(headers []*sarama.RecordHeader) string {
	for _, header := range headers {
		if string(header.Key) == "User-ID" {
			return string(header.Value)
		}
	}
	return ""
}
