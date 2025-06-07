package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/sirupsen/logrus"
)

type AuctionPublisher struct {
	producer sarama.SyncProducer
	topics   []string
	log      *logrus.Logger
}

func NewAuctionPublisher(brokers []string, topic []string, log *logrus.Logger) (*AuctionPublisher, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &AuctionPublisher{
		producer: producer,
		topics:   topic,
		log:      log,
	}, nil
}

func (p *AuctionPublisher) PublishPayment(ctx context.Context, result *models.AuctionResult) error {
	value, err := json.Marshal(result)
	if err != nil {
		p.log.WithError(err).Error("Failed to marshal auction result")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topics[0],
		Value: sarama.ByteEncoder(value),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		p.log.WithError(err).Error("Failed to send auction result message")
		return err
	}

	p.log.Info("Auction result message sent successfully")
	return nil
}

func (p *AuctionPublisher) PublishNotification(ctx context.Context, note *models.Notification) error {
	if len(p.topics) < 2 {
		p.log.Error("No topic configured for notifications")
		return fmt.Errorf("notification topic not configured")
	}

	value, err := json.Marshal(note)
	if err != nil {
		p.log.WithError(err).Error("Failed to marshal notification")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topics[1],
		Value: sarama.ByteEncoder(value),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		p.log.WithError(err).Error("Failed to send notification message")
		return err
	}

	p.log.WithField("user_id", note.UserID).Info("Notification message sent successfully")
	return nil
}

func (p *AuctionPublisher) PublishDeliveryInit(ctx context.Context, init *models.DeliveryInit) error {
	if len(p.topics) < 3 {
		return fmt.Errorf("missing delivery topic")
	}

	data, err := json.Marshal(init)
	if err != nil {
		p.log.WithError(err).Error("Failed to marshal delivery init event")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topics[2],
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		p.log.WithError(err).Error("Failed to publish delivery init")
	}
	return err
}

func (p *AuctionPublisher) Close() error {
	return p.producer.Close()
}
