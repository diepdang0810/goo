package worker

import (
	"context"
	"fmt"

	"go1/config"
	"go1/pkg/kafka"
	"go1/pkg/logger"

	"github.com/IBM/sarama"
)

type KafkaManager struct {
	consumer *kafka.Consumer
	handlers map[string]kafka.MessageHandler
	config   *config.Config
}

func NewKafkaManager(cfg *config.Config) *KafkaManager {
	return &KafkaManager{
		config:   cfg,
		handlers: make(map[string]kafka.MessageHandler),
	}
}

func (m *KafkaManager) Run(ctx context.Context) error {
	if m.consumer == nil {
		return fmt.Errorf("consumer not initialized. Use NewWorkerBuilder to build worker")
	}

	defer m.consumer.Close()

	logger.Log.Info("KafkaManager started",
		logger.Field{Key: "groupId", Value: m.config.Kafka.GroupID},
		logger.Field{Key: "topicCount", Value: len(m.handlers)})

	return m.consumer.Consume(ctx, m.routeMessage)
}

func (m *KafkaManager) routeMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	handler, ok := m.handlers[message.Topic]
	if !ok {
		logger.Log.Warn("No handler for topic", logger.Field{Key: "topic", Value: message.Topic})
		return nil
	}
	return handler(ctx, message)
}
