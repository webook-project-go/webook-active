package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/kisara71/GoTemplate/slice"
	"github.com/webook-project-go/webook-active/domain"
	"github.com/webook-project-go/webook-active/service"
	"github.com/webook-project-go/webook-pkgs/logger"
	"github.com/webook-project-go/webook-pkgs/saramax"
	"time"
)

type ActiveEvent struct {
	UID        int64
	LastActive int64
}

type consumer struct {
	client sarama.Client
	l      logger.Logger
	svc    service.Service
}

func NewKafkaConsumer(client sarama.Client, l logger.Logger, svc service.Service) *consumer {
	return &consumer{client: client, l: l, svc: svc}
}

func (c *consumer) Start() error {
	group, err := sarama.NewConsumerGroupFromClient("user_active", c.client)
	if err != nil {
		return err
	}

	hdl := saramax.NewBatchHandler(1000, time.Millisecond*100, c.Consume, c.l)
	go func() {
		err := group.Consume(context.Background(), []string{"user_active"}, hdl)
		if err != nil {
			c.l.Fatal("start consumer failed", logger.Error(err))
		}
	}()
	return nil
}
func (c *consumer) toDomain(event ActiveEvent) domain.User {
	return domain.User{
		UID:        event.UID,
		LastActive: event.LastActive,
	}
}
func (c *consumer) Consume(messages []*sarama.ConsumerMessage, ts []ActiveEvent) error {
	events, err := slice.Map(0, len(ts), ts, c.toDomain)
	if err != nil {
		c.l.Error("convert to domain failed", logger.Error(err))
		return nil
	}
	err = c.svc.MarkActive(context.Background(), events)
	if err != nil {
		c.l.Error("consume msg failed", logger.Error(err))
		return err
	}
	return nil
}
