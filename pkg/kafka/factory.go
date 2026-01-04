package kafka

import (
	"fmt"
	"go1/config"
)

// Factory creates and configures Kafka consumers based on config
type Factory struct {
	config *config.Config
}

func NewFactory(cfg *config.Config) *Factory {
	return &Factory{config: cfg}
}

// TopicHandlerConfig represents a topic with an instantiated handler
type TopicHandlerConfig struct {
	Name    string
	Handler MessageHandler
	Retry   *TopicRetryConfig
}

// BuildFromTopics creates a consumer and handler map from instantiated topic configurations
func (f *Factory) BuildFromTopics(topicConfigs []TopicHandlerConfig) (*Consumer, map[string]MessageHandler, error) {
	// Read suffix config from YAML (empty is OK, consumer will use defaults)
	retrySuffix := f.config.Kafka.Retry.RetrySuffix
	dlqSuffix := f.config.Kafka.Retry.DLQSuffix

	// Build topics list (base + retry topics)
	topics := make([]string, 0, len(topicConfigs)*2)
	handlers := make(map[string]MessageHandler)
	topicRetryMap := make(map[string]TopicRetryConfig)

	for _, tc := range topicConfigs {
		if tc.Name == "" {
			continue
		}
		// Add base topic
		topics = append(topics, tc.Name)
		handlers[tc.Name] = tc.Handler

		// If retry is enabled, add retry topic and config
		if tc.Retry != nil {
			retryTopic := tc.Name + retrySuffix
			topics = append(topics, retryTopic)
			handlers[retryTopic] = tc.Handler // Same handler for retry

			// Retry config is already in time.Duration from caller
			retryConfig := *tc.Retry

			topicRetryMap[tc.Name] = retryConfig
			topicRetryMap[retryTopic] = retryConfig
		}
	}

	if len(topics) == 0 {
		return nil, nil, fmt.Errorf("no topics to consume")
	}

	// Create consumer (defaults will be applied in NewConsumerWithOptions if suffixes are empty)
	consumer, err := NewConsumerWithOptions(
		f.config.Kafka.Brokers,
		topics,
		ConsumerOptions{
			GroupID:          f.config.Kafka.GroupID,
			RetryTopicSuffix: retrySuffix,
			DLQTopicSuffix:   dlqSuffix,
			TopicRetryConfig: topicRetryMap,
		},
		f.config.Kafka.Producer, // Producer config for retry/DLQ publishing
		f.config.Kafka.Consumer, // Consumer config for session timeout, heartbeat, etc.
	)
	if err != nil {
		return nil, nil, err
	}

	return consumer, handlers, nil
}
