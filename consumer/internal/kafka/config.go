package kafka

import "github.com/IBM/sarama"

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	return config
}
