package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	addrs := viper.GetStringSlice("kafka.addrs")
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	client, err := sarama.NewClient(addrs, cfg)
	if err != nil {
		panic(err)
	}
	return client
}
