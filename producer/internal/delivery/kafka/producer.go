package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/producer/pkg"
)

type Producer struct {
	syncProducer sarama.SyncProducer
	packageTopic string
	paymentTopic string
}

func NewProducer(cfg Config) (*Producer, error) {
	saramaConfig := NewProducerConfig()

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}
	return &Producer{
		syncProducer: producer,
		packageTopic: cfg.Topic,
		paymentTopic: cfg.PaymentTopic,
	}, nil
}

func (p *Producer) SendPackage(pkg pkg.Package, userID string) error {
	pkgJson, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	headers := []sarama.RecordHeader{
		{
			Key:   []byte("User-ID"),
			Value: []byte(userID),
		},
	}

	msg := &sarama.ProducerMessage{
		Headers: headers,
		Topic:   p.packageTopic,
		Value:   sarama.ByteEncoder(pkgJson),
	}

	_, _, err = p.syncProducer.SendMessage(msg)
	return err
}

func (p *Producer) SendPaymentEvent(event pkg.PaymentEvent) error {
	eventJson, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.paymentTopic,
		Value: sarama.ByteEncoder(eventJson),
	}

	_, _, err = p.syncProducer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	return p.syncProducer.Close()
}
