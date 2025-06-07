package kafka

import "github.com/IBM/sarama"

type ConfigProducer struct {
	Brokers []string
	Topic   []string
	Version string
}

func NewProducerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	return config
}

type ConfigConsumer struct {
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
