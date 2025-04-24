package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/producer/pkg"
)

type Producer struct {
	syncProducer sarama.SyncProducer
	topic        string
}

func NewProducer(cfg Config) (*Producer, error) {
	saramaConfig := NewProducerConfig()

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}
	return &Producer{
		syncProducer: producer,
		topic:        cfg.Topic,
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
		Topic:   p.topic,
		Value:   sarama.ByteEncoder(pkgJson),
	}

	_, _, err = p.syncProducer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	return p.syncProducer.Close()
}
