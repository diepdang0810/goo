package worker

import (
	"go1/config"
	"go1/internal/modules/order/infrastructure/repository"
	workerHandlers "go1/internal/worker/handlers"
	"go1/pkg/kafka"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"go.temporal.io/sdk/client"
)

// HandlerFactory is a function that creates a handler with injected dependencies
type HandlerFactory func(pg *postgres.Postgres, rds *redis.RedisClient, temporalClient client.Client) kafka.MessageHandler

// TopicConfig represents configuration for a single topic with handler factory
type TopicConfig struct {
	Name           string
	HandlerFactory HandlerFactory    // Factory to create handler with dependencies
	Retry          *TopicRetryConfig // nil means no retry
}

// InstantiatedTopicConfig represents a topic with an instantiated handler
type InstantiatedTopicConfig struct {
	Name    string
	Handler kafka.MessageHandler
	Retry   *TopicRetryConfig
}

// TopicRetryConfig for a specific topic
type TopicRetryConfig struct {
	MaxAttempts int
	Backoff     int // in milliseconds
}

// WorkerBuilder helps build a worker with topics in a fluent way
type WorkerBuilder struct {
	config         *config.Config
	topicConfigs   []TopicConfig
	postgres       *postgres.Postgres
	redis          *redis.RedisClient
	temporalClient client.Client
}

func NewWorkerBuilder(cfg *config.Config) *WorkerBuilder {
	return &WorkerBuilder{
		config:       cfg,
		topicConfigs: make([]TopicConfig, 0),
	}
}

// WithPostgres sets the PostgreSQL connection for the worker
func (b *WorkerBuilder) WithPostgres(pg *postgres.Postgres) *WorkerBuilder {
	b.postgres = pg
	return b
}

// WithRedis sets the Redis connection for the worker
func (b *WorkerBuilder) WithRedis(rc *redis.RedisClient) *WorkerBuilder {
	b.redis = rc
	return b
}

// WithTemporal sets the Temporal client for the worker
func (b *WorkerBuilder) WithTemporal(c client.Client) *WorkerBuilder {
	b.temporalClient = c
	return b
}

// AddTopic adds a topic to consume with a handler factory
func (b *WorkerBuilder) AddTopic(name string, factory HandlerFactory) *WorkerBuilder {
	// Get retry config from YAML if exists
	var retryConfig *TopicRetryConfig
	if topicCfg, exists := b.config.Kafka.Retry.Topics[name]; exists && topicCfg.EnableRetry {
		retryConfig = &TopicRetryConfig{
			MaxAttempts: topicCfg.MaxAttempts,
			Backoff:     topicCfg.BackoffMs,
		}
	}

	b.topicConfigs = append(b.topicConfigs, TopicConfig{
		Name:           name,
		HandlerFactory: factory,
		Retry:          retryConfig,
	})
	return b
}

// Build creates the worker with all configured topics
func (b *WorkerBuilder) Build() (*Worker, error) {
	// Instantiate handlers with injected dependencies
	instantiatedConfigs := make([]InstantiatedTopicConfig, 0, len(b.topicConfigs))
	for _, tc := range b.topicConfigs {
		handler := tc.HandlerFactory(b.postgres, b.redis, b.temporalClient)
		instantiatedConfigs = append(instantiatedConfigs, InstantiatedTopicConfig{
			Name:    tc.Name,
			Handler: handler,
			Retry:   tc.Retry,
		})
	}

	factory := NewKafkaConsumerFactory(b.config)
	consumer, handlers, err := factory.BuildFromTopics(instantiatedConfigs)
	if err != nil {
		return nil, err
	}

	return &Worker{
		consumer: consumer,
		handlers: handlers,
		config:   b.config,
		postgres: b.postgres,
		redis:    b.redis,
	}, nil
}

func (b *WorkerBuilder) WithShipmentEvents() *WorkerBuilder {
	b.topicConfigs = append(b.topicConfigs, TopicConfig{
		Name: "shipment-events",
		HandlerFactory: func(pg *postgres.Postgres, rds *redis.RedisClient, temporalClient client.Client) kafka.MessageHandler {
			repo := repository.NewPostgresOrderRepository(pg.Pool)
			handler := workerHandlers.NewShipmentEventHandler(temporalClient, repo)
			return handler.Handle
		},
	})
	return b
}
