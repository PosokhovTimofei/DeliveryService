package kafka

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/sirupsen/logrus"
)

type AuctionPublisher struct {
	producer sarama.SyncProducer
	topic    string
	log      *logrus.Logger
}

func NewAuctionPublisher(brokers []string, topic string, log *logrus.Logger) (*AuctionPublisher, error) {
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
		topic:    topic,
		log:      log,
	}, nil
}

func (p *AuctionPublisher) Publish(ctx context.Context, result *models.AuctionResult) error {
	value, err := json.Marshal(result)
	if err != nil {
		p.log.WithError(err).Error("Failed to marshal auction result")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.log.WithError(err).Error("Failed to send auction result message")
		return err
	}

	p.log.WithFields(logrus.Fields{
		"partition": partition,
		"offset":    offset,
	}).Info("Auction result message sent successfully")
	return nil
}

func (p *AuctionPublisher) Close() error {
	return p.producer.Close()
}
