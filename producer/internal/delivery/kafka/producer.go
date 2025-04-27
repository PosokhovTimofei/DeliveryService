package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/maksroxx/DeliveryService/producer/pkg"
)

const (
	maxCommitRetries = 3
	retryDelay       = 2 * time.Second
)

type Producer struct {
	syncProducer sarama.SyncProducer
	packageTopic string
	paymentTopic string
}

func NewProducer(cfg Config) (*Producer, error) {
	transactionalID := "producer-" + uuid.NewString()
	saramaConfig, err := NewTransactionalConfig(transactionalID, cfg.Version)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{
		syncProducer: producer,
		packageTopic: cfg.PackageTopic,
		paymentTopic: cfg.PaymentTopic,
	}, nil
}

func (p *Producer) BeginTransaction() error {
	return p.syncProducer.BeginTxn()
}

func (p *Producer) CommitTransaction() error {
	var lastErr error

	for attempt := 1; attempt <= maxCommitRetries; attempt++ {
		err := p.syncProducer.CommitTxn()
		if err == nil {
			return nil
		}

		lastErr = err

		if errors.Is(err, sarama.ErrUnknownProducerID) ||
			errors.Is(err, sarama.ErrInvalidProducerEpoch) ||
			errors.Is(err, sarama.ErrTransactionCoordinatorFenced) {
			_ = p.syncProducer.AbortTxn()
			return err
		}

		log.Printf("CommitTxn failed: %v. Retrying after %s...", err, retryDelay)
		time.Sleep(retryDelay)
	}

	log.Printf("All commit attempts failed after %d tries. Aborting transaction...", maxCommitRetries)
	_ = p.syncProducer.AbortTxn()
	return lastErr
}

func (p *Producer) AbortTransaction() error {
	return p.syncProducer.AbortTxn()
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
		Topic:   p.packageTopic,
		Key:     sarama.StringEncoder(pkg.ID),
		Value:   sarama.ByteEncoder(pkgJson),
		Headers: headers,
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
		Key:   sarama.StringEncoder(event.PackageID),
		Value: sarama.ByteEncoder(eventJson),
	}
	_, _, err = p.syncProducer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	return p.syncProducer.Close()
}
