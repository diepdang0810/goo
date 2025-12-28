package worker

import (
	"context"

	"go1/config"
	"go1/internal/modules/order/infrastructure/repository"
	"go1/internal/worker/handlers"
	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"go1/internal/modules/order/activity"
	"go1/internal/modules/order/workflow"

	"go.temporal.io/sdk/client"
	tWorker "go.temporal.io/sdk/worker"
)

// App encapsulates worker application dependencies and initialization
type App struct {
	worker         *Worker
	postgres       *postgres.Postgres
	redis          *redis.RedisClient
	config         *config.Config
	temporalClient client.Client
	temporalWorker tWorker.Worker
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
	if err := app.initTemporal(); err != nil {
		return nil, err
	}
	if err := app.initTemporalWorker(); err != nil {
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

func (a *App) initTemporal() error {
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233", // Should be from config
	})
	if err != nil {
		return err
	}
	a.temporalClient = c
	return nil
}

func (a *App) initTemporalWorker() error {
	w := tWorker.New(a.temporalClient, "ORDER_TASK_QUEUE", tWorker.Options{})

	w.RegisterWorkflow(workflow.CreateOrderWorkflow)

	repo := repository.NewPostgresOrderRepository(a.postgres.Pool)
	act := activity.NewOrderActivities(repo)
	w.RegisterActivity(act)

	a.temporalWorker = w
	return nil
}

func (a *App) initWorker() error {
	// Register handlers here with NEW simplified API (auto-unmarshal!)
	w, err := NewWorkerBuilder(a.config).
		WithPostgres(a.postgres).
		WithRedis(a.redis).
		WithTemporal(a.temporalClient).
		AddTopic("user_created", func(pg *postgres.Postgres, rds *redis.RedisClient, tc client.Client) kafka.MessageHandler {
			return handlers.NewUserCreatedHandler(pg, rds).Handle()
		}).
		AddTopic("dbserver1.public.users", func(pg *postgres.Postgres, rds *redis.RedisClient, tc client.Client) kafka.MessageHandler {
			return handlers.NewUserCDCHandler(pg, rds).Handle()
		}).
		AddTopic("test_success", func(pg *postgres.Postgres, rds *redis.RedisClient, tc client.Client) kafka.MessageHandler {
			return handlers.NewTestSuccessHandlerV2(pg, rds).Handle()
		}).
		AddTopic("test_retry", func(pg *postgres.Postgres, rds *redis.RedisClient, tc client.Client) kafka.MessageHandler {
			return handlers.NewTestRetryHandlerV2(pg, rds).Handle()
		}).
		WithShipmentEvents().
		Build()

	if err != nil {
		return err
	}

	a.worker = w
	return nil
}

// Run starts the worker and blocks until context is cancelled
func (a *App) Run(ctx context.Context) error {
	if a.temporalWorker != nil {
		if err := a.temporalWorker.Start(); err != nil {
			return err
		}
		defer a.temporalWorker.Stop()
	}
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
