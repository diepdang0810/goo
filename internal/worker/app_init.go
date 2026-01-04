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

func (a *App) initLogger() error {
	log := logger.NewZapLogger(a.config.App.Env)
	logger.SetLogger(log)
	return nil
}

func (a *App) initPostgres() error {
	pg, err := postgres.New(
		a.config.Postgres.Host,
		a.config.Postgres.Port,
		a.config.Postgres.User,
		a.config.Postgres.Password,
		a.config.Postgres.DBName,
		a.config.Postgres.MaxConns,
		a.config.Postgres.MinConns,
	)
	if err != nil {
		return err
	}
	a.postgres = pg
	logger.Log.Info("Connected to PostgreSQL", logger.Field{Key: "database", Value: a.config.Postgres.DBName})
	return nil
}

func (a *App) initRedis() error {
	redisClient, err := redis.New(a.config.Redis.Addr)
	if err != nil {
		return err
	}
	a.redis = redisClient
	logger.Log.Info("Connected to Redis", logger.Field{Key: "addr", Value: a.config.Redis.Addr})
	return nil
}

func (a *App) initTemporal() error {
	c, err := client.Dial(client.Options{
		HostPort: a.config.Temporal.HostPort,
	})
	if err != nil {
		return err
	}
	a.temporalClient = c
	return nil
}

func (a *App) initTemporalWorker() error {
	w := tWorker.New(a.temporalClient, a.config.Temporal.TaskQueue, tWorker.Options{})

	w.RegisterWorkflow(workflow.CreateOrderWorkflow)

	repo := repository.NewPostgresOrderRepository(a.postgres.Pool)
	act := activity.NewOrderActivities(repo)
	w.RegisterActivity(act)

	a.temporalWorker = w
	return nil
}

func (a *App) initWorker() error {
	// Register handlers here with NEW simplified API (auto-unmarshal!)
	// Note: Retry configuration for topics is automatically loaded by AddTopic
	w, err := NewWorkerBuilder(a.config).
		WithPostgres(a.postgres).
		WithRedis(a.redis).
		WithTemporal(a.temporalClient).
		WithShipmentEvents().
		WithDispatchEvents().
		WithOrderEvents().
		Build()

	if err != nil {
		return err
	}

	a.worker = w
	return nil
}
