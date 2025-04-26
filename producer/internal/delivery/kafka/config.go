package kafka

import "github.com/IBM/sarama"

type Config struct {
	Brokers      []string
	Topic        string
	PaymentTopic string
	Version      string
}

func NewProducerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	return config
}
