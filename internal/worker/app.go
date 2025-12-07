package worker

import (
	"context"

	"go1/config"
	"go1/internal/worker/handlers"
	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"
)

// App encapsulates worker application dependencies and initialization
type App struct {
	worker   *Worker
	postgres *postgres.Postgres
	redis    *redis.RedisClient
	config   *config.Config
}

// NewApp creates and initializes a new worker application
func NewApp(cfg *config.Config) (*App, error) {
	app := &App{config: cfg}

	if err := app.initLogger(); err != nil {
		return nil, err
	}
	if err := app.initPostgres(); err != nil {
		return nil, err
	}
	if err := app.initRedis(); err != nil {
		return nil, err
	}
	if err := app.initWorker(); err != nil {
		return nil, err
	}

	return app, nil
}

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

func (a *App) initWorker() error {
	// Register handlers here
	w, err := NewWorkerBuilder(a.config).
		WithPostgres(a.postgres).
		WithRedis(a.redis).
		AddTopic("user_created", func(pg *postgres.Postgres, rds *redis.RedisClient) kafka.MessageHandler {
			return handlers.NewUserCreatedHandler(pg, rds).Handle
		}).
		// Easy to add more topics:
		// AddTopic("order_created", func(pg *postgres.Postgres, rds *redis.RedisClient) kafka.MessageHandler {
		// 	return handlers.NewOrderCreatedHandler(pg, rds).Handle
		// }).
		Build()

	if err != nil {
		return err
	}

	a.worker = w
	return nil
}

// Run starts the worker and blocks until context is cancelled
func (a *App) Run(ctx context.Context) error {
	return a.worker.Run(ctx)
}

// Close gracefully shuts down the worker application
func (a *App) Close() {
	if a.postgres != nil {
		a.postgres.Close()
	}
	if a.redis != nil {
		a.redis.Close()
	}
	logger.Log.Info("Worker application closed")
}
