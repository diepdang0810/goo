package kafka

import (
	"context"
	"strings"

	"go1/pkg/logger"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	Client *kgo.Client
}

type MessageHandler func(context.Context, *kgo.Record) error

func NewConsumer(brokers, groupID string, topics []string) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &Consumer{Client: client}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := c.Client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return nil
			}

			fetches.EachError(func(topic string, partition int32, err error) {
				logger.Log.Error("Fetch error", 
					logger.Field{Key: "topic", Value: topic}, 
					logger.Field{Key: "partition", Value: partition},
					logger.Field{Key: "error", Value: err})
			})

			fetches.EachRecord(func(record *kgo.Record) {
				if err := handler(ctx, record); err != nil {
					logger.Log.Error("Handler error", 
						logger.Field{Key: "topic", Value: record.Topic}, 
						logger.Field{Key: "error", Value: err})
				}
			})
		}
	}
}

func (c *Consumer) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
}
