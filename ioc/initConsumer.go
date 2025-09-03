package ioc

import (
	"github.com/IBM/sarama"
	"github.com/webook-project-go/webook-active/events"
	"github.com/webook-project-go/webook-active/service"
	"github.com/webook-project-go/webook-pkgs/logger"
)

func InitConsumer(client sarama.Client, l logger.Logger, svc service.Service) []events.Consumer {
	consumer := events.NewKafkaConsumer(client, l, svc)
	return []events.Consumer{consumer}
}
