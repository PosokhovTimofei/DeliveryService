package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/auction/internal/metrics"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/sirupsen/logrus"
)

type AucPublisher interface {
	PublishPayment(ctx context.Context, result *models.AuctionResult) error
	PublishNotification(ctx context.Context, note *models.Notification) error
	PublishDeliveryInit(ctx context.Context, init *models.DeliveryInit) error
	Close() error
}

type AuctionPublisher struct {
	producer sarama.SyncProducer
	topics   []string
	log      *logrus.Logger
}

func NewAuctionPublisher(brokers []string, topic []string, log *logrus.Logger) (AucPublisher, error) {
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
	start := time.Now()
	topic := p.topics[0]

	value, err := json.Marshal(result)
	if err != nil {
		metrics.KafkaMessagesError.WithLabelValues(topic, "marshal_error").Inc()
		p.log.WithError(err).Error("Failed to marshal auction result")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	_, _, err = p.producer.SendMessage(msg)
	metrics.KafkaPublishDuration.WithLabelValues(topic).Observe(time.Since(start).Seconds())

	if err != nil {
		metrics.KafkaMessagesError.WithLabelValues(topic, "send_error").Inc()
		p.log.WithError(err).Error("Failed to send auction result message")
		return err
	}

	metrics.KafkaMessagesSent.WithLabelValues(topic).Inc()
	p.log.Info("Auction result message sent successfully")
	return nil
}

func (p *AuctionPublisher) PublishNotification(ctx context.Context, note *models.Notification) error {
	if len(p.topics) < 2 {
		metrics.KafkaMessagesError.WithLabelValues("notification", "no_topic").Inc()
		p.log.Error("No topic configured for notifications")
		return fmt.Errorf("notification topic not configured")
	}

	start := time.Now()
	topic := p.topics[1]

	value, err := json.Marshal(note)
	if err != nil {
		metrics.KafkaMessagesError.WithLabelValues(topic, "marshal_error").Inc()
		p.log.WithError(err).Error("Failed to marshal notification")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	_, _, err = p.producer.SendMessage(msg)
	metrics.KafkaPublishDuration.WithLabelValues(topic).Observe(time.Since(start).Seconds())

	if err != nil {
		metrics.KafkaMessagesError.WithLabelValues(topic, "send_error").Inc()
		p.log.WithError(err).Error("Failed to send notification message")
		return err
	}

	metrics.KafkaMessagesSent.WithLabelValues(topic).Inc()
	p.log.WithField("user_id", note.UserID).Info("Notification message sent successfully")
	return nil
}

func (p *AuctionPublisher) PublishDeliveryInit(ctx context.Context, init *models.DeliveryInit) error {
	if len(p.topics) < 3 {
		metrics.KafkaMessagesError.WithLabelValues("delivery", "no_topic").Inc()
		return fmt.Errorf("missing delivery topic")
	}

	start := time.Now()
	topic := p.topics[2]

	data, err := json.Marshal(init)
	if err != nil {
		metrics.KafkaMessagesError.WithLabelValues(topic, "marshal_error").Inc()
		p.log.WithError(err).Error("Failed to marshal delivery init event")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	metrics.KafkaPublishDuration.WithLabelValues(topic).Observe(time.Since(start).Seconds())

	if err != nil {
		metrics.KafkaMessagesError.WithLabelValues(topic, "send_error").Inc()
		p.log.WithError(err).Error("Failed to publish delivery init")
		return err
	}

	metrics.KafkaMessagesSent.WithLabelValues(topic).Inc()
	return nil
}

func (p *AuctionPublisher) Close() error {
	return p.producer.Close()
}
