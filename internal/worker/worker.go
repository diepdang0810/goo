package worker

import (
	"context"
	"fmt"

	"go1/config"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"go.temporal.io/sdk/client"
	tWorker "go.temporal.io/sdk/worker"
)

// Worker encapsulates worker application dependencies and initialization
type Worker struct {
	kafkaManager   *KafkaManager
	postgres       *postgres.Postgres
	redis          *redis.RedisClient
	config         *config.Config
	temporalClient client.Client
	temporalWorker tWorker.Worker
}

// New creates and initializes a new worker application
func New(cfg *config.Config) (*Worker, error) {
	worker := &Worker{config: cfg}

	if err := worker.initLogger(); err != nil {
		return nil, err
	}
	if err := worker.initPostgres(); err != nil {
		return nil, err
	}
	if err := worker.initRedis(); err != nil {
		return nil, err
	}
	if err := worker.initTemporal(); err != nil {
		return nil, err
	}
	if err := worker.initTemporalWorker(); err != nil {
		return nil, err
	}
	if err := worker.initKafkaManager(); err != nil {
		return nil, err
	}

	return worker, nil
}

// Run starts the worker and blocks until context is cancelled
func (w *Worker) Run(ctx context.Context) error {
	if w.temporalWorker != nil {
		if err := w.temporalWorker.Start(); err != nil {
			return fmt.Errorf("failed to start temporal worker: %w", err)
		}
		defer w.temporalWorker.Stop()
	}
	return w.kafkaManager.Run(ctx)
}

// Close gracefully shuts down the worker application
func (w *Worker) Close() {
	if w.temporalClient != nil {
		w.temporalClient.Close()
	}
	if w.postgres != nil {
		w.postgres.Close()
	}
	if w.redis != nil {
		w.redis.Close()
	}
	logger.Log.Info("Worker application closed")
}
