package kafka

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	handler  sarama.ConsumerGroupHandler
	topic    string
	log      *logrus.Logger
}

func NewConsumer(cfg ConfigConsumer, handler sarama.ConsumerGroupHandler, log *logrus.Logger) (*Consumer, error) {
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

func NewTestableConsumer(
	group sarama.ConsumerGroup,
	handler sarama.ConsumerGroupHandler,
	topic string,
	log *logrus.Logger,
) *Consumer {
	return &Consumer{
		consumer: group,
		handler:  handler,
		topic:    topic,
		log:      log,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	var retryCount int
	for {
		select {
		case <-ctx.Done():
			c.log.Info("Stopping consumer")
			return
		default:
			err := c.consume(ctx)
			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					c.log.Warn("Consumer group closed, exiting")
					return
				}

				retryCount++
				backoffDuration := CalculateBackoff(retryCount)
				c.log.WithFields(logrus.Fields{
					"error":            err,
					"retry_in_seconds": backoffDuration.Seconds(),
					"retry_attempt":    retryCount,
				}).Error("Consume error, will retry")

				time.Sleep(backoffDuration)
			} else {
				retryCount = 0
			}
		}
	}
}

func (c *Consumer) consume(ctx context.Context) error {
	c.log.Infof("Starting consumption on topic: %s", c.topic)
	err := c.consumer.Consume(ctx, []string{c.topic}, c.handler)
	if err != nil {
		c.log.WithError(err).Error("Error during consumption")
	}
	return err
}

func (c *Consumer) Close() error {
	c.log.Info("Closing Kafka consumer group")
	if err := c.consumer.Close(); err != nil {
		c.log.WithError(err).Error("Error closing Kafka consumer")
		return err
	}
	c.log.Info("Kafka consumer closed successfully")
	return nil
}

func (c *Consumer) RunOnce(ctx context.Context) error {
	return c.consume(ctx)
}

func CalculateBackoff(retryCount int) time.Duration {
	const (
		baseDelay = 2 * time.Second
		maxDelay  = 1 * time.Minute
	)

	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(retryCount-1)))

	if delay > maxDelay {
		delay = maxDelay
	}

	jitterFactor := 0.5 + rand.Float64()
	jitteredDelay := time.Duration(float64(delay) * jitterFactor)

	if jitteredDelay > maxDelay {
		return maxDelay
	}
	return jitteredDelay
}
