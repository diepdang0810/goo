package worker

import (
	"context"

	"go1/config"
	"go1/internal/worker/handlers"
	"go1/pkg/kafka"
	"go1/pkg/logger"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Worker struct {
	consumer *kafka.Consumer
	handlers map[string]kafka.MessageHandler
	config   *config.Config
}

func NewWorker(cfg *config.Config) *Worker {
	return &Worker{
		config:   cfg,
		handlers: make(map[string]kafka.MessageHandler),
	}
}

func (w *Worker) Run(ctx context.Context) error {
	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(
		w.config.Kafka.Brokers,
		"user-worker-group",
		[]string{"user_created"},
	)
	if err != nil {
		return err
	}
	w.consumer = consumer
	defer w.consumer.Close()

	// Setup handlers
	w.setupHandlers()

	// Start consuming
	logger.Log.Info("Worker started", logger.Field{Key: "topics", Value: "user_created"})

	return w.consumer.Consume(ctx, w.routeMessage)
}

func (w *Worker) setupHandlers() {
	// Register handler cho từng topic
	userCreatedHandler := handlers.NewUserCreatedHandler()
	w.handlers["user_created"] = userCreatedHandler.Handle
}

func (w *Worker) routeMessage(ctx context.Context, record *kgo.Record) error {
	handler, ok := w.handlers[record.Topic]
	if !ok {
		logger.Log.Warn("No handler for topic", logger.Field{Key: "topic", Value: record.Topic})
		return nil
	}

	return handler(ctx, record)
}
