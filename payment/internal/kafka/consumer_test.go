package kafka_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/payment/internal/kafka"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockConsumerGroup struct {
	consumeFn func(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error
	closeFn   func() error
}

func (m *mockConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return m.consumeFn(ctx, topics, handler)
}

func (m *mockConsumerGroup) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func (m *mockConsumerGroup) Errors() <-chan error {
	return make(chan error)
}

func (m *mockConsumerGroup) Pause(map[string][]int32) {}

func (m *mockConsumerGroup) Resume(map[string][]int32) {}

func (m *mockConsumerGroup) PauseAll() {}

func (m *mockConsumerGroup) ResumeAll() {}

func (m *mockConsumerGroup) Claims() map[string][]int32 {
	return nil
}

func (m *mockConsumerGroup) MemberID() string {
	return "mock-member"
}

func (m *mockConsumerGroup) GenerationID() int32 {
	return 1
}

func TestConsumerConsumeSuccess(t *testing.T) {
	mockGroup := &mockConsumerGroup{
		consumeFn: func(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
			return nil
		},
		closeFn: func() error {
			return nil
		},
	}

	c := kafka.NewTestableConsumer(
		mockGroup,
		nil,
		"test-topic",
		logrus.New(),
	)

	err := c.RunOnce(context.Background())
	assert.NoError(t, err)
}

func TestConsumerConsumeFailure(t *testing.T) {
	mockGroup := &mockConsumerGroup{
		consumeFn: func(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
			return errors.New("consume failure")
		},
		closeFn: func() error {
			return nil
		},
	}

	c := kafka.NewTestableConsumer(
		mockGroup,
		nil,
		"test-topic",
		logrus.New(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := c.RunOnce(ctx)
	assert.EqualError(t, err, "consume failure")
}

func TestCalculateBackoff(t *testing.T) {
	for i := 1; i <= 10; i++ {
		delay := kafka.CalculateBackoff(i)
		assert.LessOrEqual(t, delay, time.Minute)
		assert.Greater(t, delay, 0*time.Second)
	}
}

func TestConsumerCloseSuccess(t *testing.T) {
	mockGroup := &mockConsumerGroup{
		closeFn: func() error {
			return nil
		},
	}

	c := kafka.NewTestableConsumer(mockGroup, nil, "topic", logrus.New())
	err := c.Close()
	assert.NoError(t, err)
}

func TestConsumerRunStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mockGroup := &mockConsumerGroup{
		consumeFn: func(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
			cancel()
			return nil
		},
	}

	c := kafka.NewTestableConsumer(mockGroup, nil, "topic", logrus.New())
	c.Run(ctx)
}
