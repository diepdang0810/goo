package worker

import (
	"go1/config"
	"go1/pkg/kafka"
	"strings"
	"time"
)

// TopicConfig represents configuration for a single topic with pre-configured handler
type TopicConfig struct {
	Name    string
	Handler kafka.MessageHandler // Handler already has dependencies injected by closure
	Retry   *TopicRetryConfig    // nil means no retry
}

// TopicRetryConfig for a specific topic
type TopicRetryConfig struct {
	MaxAttempts int
	Backoff     int // in milliseconds
}

// WorkerBuilder helps build a worker with topics in a fluent way
type WorkerBuilder struct {
	config       *config.Config
	topicConfigs []TopicConfig
}

func NewWorkerBuilder(cfg *config.Config) *WorkerBuilder {
	return &WorkerBuilder{
		config:       cfg,
		topicConfigs: make([]TopicConfig, 0),
	}
}

// AddTopic adds a topic to consume with a pre-configured handler
func (b *WorkerBuilder) AddTopic(name string, handler kafka.MessageHandler) *WorkerBuilder {
	var retryConfig *TopicRetryConfig
	topicCfg, exists := b.config.Kafka.Retry.Topics[name]
	if !exists {
		// Fallback for topics with dots (Viper issue): try replacing dots with underscores
		sanitized := strings.ReplaceAll(name, ".", "_")
		topicCfg, exists = b.config.Kafka.Retry.Topics[sanitized]
	}

	if exists && topicCfg.EnableRetry {
		retryConfig = &TopicRetryConfig{
			MaxAttempts: topicCfg.MaxAttempts,
			Backoff:     topicCfg.BackoffMs,
		}
	}

	b.topicConfigs = append(b.topicConfigs, TopicConfig{
		Name:    name,
		Handler: handler,
		Retry:   retryConfig,
	})
	return b
}

// Build creates the worker with all configured topics
func (b *WorkerBuilder) Build() (*KafkaManager, error) {
	// Instantiate handlers
	topicHandlerConfigs := make([]kafka.TopicHandlerConfig, 0, len(b.topicConfigs))
	for _, tc := range b.topicConfigs {
		var retryConfig *kafka.TopicRetryConfig
		if tc.Retry != nil {
			retryConfig = &kafka.TopicRetryConfig{
				MaxAttempts: tc.Retry.MaxAttempts,
				Backoff:     time.Duration(tc.Retry.Backoff) * time.Millisecond,
			}
		}

		topicHandlerConfigs = append(topicHandlerConfigs, kafka.TopicHandlerConfig{
			Name:    tc.Name,
			Handler: tc.Handler,
			Retry:   retryConfig,
		})
	}

	factory := kafka.NewFactory(b.config)
	consumer, handlers, err := factory.BuildFromTopics(topicHandlerConfigs)
	if err != nil {
		return nil, err
	}

	return &KafkaManager{
		consumer: consumer,
		handlers: handlers,
		config:   b.config,
	}, nil
}
