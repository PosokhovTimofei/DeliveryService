package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	handler  sarama.ConsumerGroupHandler
	topic    string
	log      *logrus.Logger
}

func NewConsumer(cfg Config, handler sarama.ConsumerGroupHandler, log *logrus.Logger) (*Consumer, error) {
	config := NewConsumerConfig()

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		handler:  handler,
		topic:    cfg.Topic,
		log:      log,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.log.Info("Stopping consumer")
			return
		default:
			if err := c.consumer.Consume(ctx, []string{c.topic}, c.handler); err != nil {
				c.log.WithError(err).Error("Consume error")
			}
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		c.log.WithError(err).Error("Error closing Kafka consumer")
		return err
	}
	c.log.Info("Kafka consumer closed successfully")
	return nil
}
