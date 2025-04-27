package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
)

type Config struct {
	Brokers      []string
	PackageTopic string
	PaymentTopic string
	Version      string
}

func NewTransactionalConfig(transactionalID, versionStr string) (*sarama.Config, error) {
	config := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Kafka version: %w", err)
	}
	config.Version = version

	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1
	config.Producer.Transaction.ID = transactionalID

	return config, nil
}
