package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/database/internal/models"
)

type Producer struct {
	syncProducer sarama.SyncProducer
	topic        string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &Producer{
		syncProducer: producer,
		topic:        topic,
	}, nil
}

func (p *Producer) SendPaymentEvent(payment models.Payment) error {
	msgBytes, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(payment.PackageID),
		Value: sarama.ByteEncoder(msgBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("User-ID"),
				Value: []byte(payment.UserID),
			},
		},
	}

	_, _, err = p.syncProducer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	return p.syncProducer.Close()
}
