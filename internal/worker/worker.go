package worker

import (
	"context"
	"fmt"

	"go1/config"
	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Worker struct {
	consumer *kafka.Consumer
	handlers map[string]kafka.MessageHandler
	config   *config.Config
	postgres *postgres.Postgres
	redis    *redis.RedisClient
}

func NewWorker(cfg *config.Config) *Worker {
	return &Worker{
		config:   cfg,
		handlers: make(map[string]kafka.MessageHandler),
	}
}

func (w *Worker) Run(ctx context.Context) error {
	if w.consumer == nil {
		return fmt.Errorf("consumer not initialized. Use NewWorkerBuilder to build worker")
	}

	defer w.consumer.Close()

	logger.Log.Info("Worker started",
		logger.Field{Key: "groupId", Value: w.config.Kafka.GroupID},
		logger.Field{Key: "topicCount", Value: len(w.handlers)})

	return w.consumer.Consume(ctx, w.routeMessage)
}

func (w *Worker) routeMessage(ctx context.Context, record *kgo.Record) error {
	handler, ok := w.handlers[record.Topic]
	if !ok {
		logger.Log.Warn("No handler for topic", logger.Field{Key: "topic", Value: record.Topic})
		return nil
	}
	return handler(ctx, record)
}

// GetPostgres returns the PostgreSQL connection (can be nil if not initialized)
func (w *Worker) GetPostgres() *postgres.Postgres {
	return w.postgres
}

// GetRedis returns the Redis connection (can be nil if not initialized)
func (w *Worker) GetRedis() *redis.RedisClient {
	return w.redis
}
