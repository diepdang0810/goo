package worker

import (
	"go1/internal/shared/order/activity"
	"go1/internal/shared/order/infrastructure/repository"
	"go1/internal/shared/order/workflow"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"go.temporal.io/sdk/client"
	tWorker "go.temporal.io/sdk/worker"
)

func (w *Worker) initLogger() error {
	log := logger.NewZapLogger(w.config.App.Env)
	logger.SetLogger(log)
	return nil
}

func (w *Worker) initPostgres() error {
	pg, err := postgres.New(
		w.config.Postgres.Host,
		w.config.Postgres.Port,
		w.config.Postgres.User,
		w.config.Postgres.Password,
		w.config.Postgres.DBName,
		w.config.Postgres.MaxConns,
		w.config.Postgres.MinConns,
	)
	if err != nil {
		return err
	}
	w.postgres = pg
	logger.Log.Info("Connected to PostgreSQL", logger.Field{Key: "database", Value: w.config.Postgres.DBName})
	return nil
}

func (w *Worker) initRedis() error {
	redisClient, err := redis.New(w.config.Redis.Addr)
	if err != nil {
		return err
	}
	w.redis = redisClient
	logger.Log.Info("Connected to Redis", logger.Field{Key: "addr", Value: w.config.Redis.Addr})
	return nil
}

func (w *Worker) initTemporal() error {
	c, err := client.Dial(client.Options{
		HostPort: w.config.Temporal.HostPort,
	})
	if err != nil {
		return err
	}
	w.temporalClient = c
	return nil
}

func (w *Worker) initTemporalWorker() error {
	tw := tWorker.New(w.temporalClient, w.config.Temporal.TaskQueue, tWorker.Options{})

	tw.RegisterWorkflow(workflow.CreateOrderWorkflow)

	repo := repository.NewPostgresOrderRepository(w.postgres.Pool)
	act := activity.NewOrderActivities(repo)
	tw.RegisterActivity(act)

	w.temporalWorker = tw
	return nil
}

func (w *Worker) initKafkaManager() error {
	// Register handlers here with explicit dependencies
	// Note: Retry configuration for topics is automatically loaded by AddTopic
	manager, err := NewWorkerBuilder(w.config).
		WithShipmentEvents(w.postgres, w.temporalClient).
		WithDispatchEvents(w.postgres, w.temporalClient).
		WithOrderEvents(w.temporalClient).
		Build()

	if err != nil {
		return err
	}

	w.kafkaManager = manager
	return nil
}
