package worker

import (
	"time"

	"go1/config"
	"go1/pkg/kafka"
)

// KafkaConsumerFactory creates and configures Kafka consumers based on config
type KafkaConsumerFactory struct {
	config *config.Config
}

func NewKafkaConsumerFactory(cfg *config.Config) *KafkaConsumerFactory {
	return &KafkaConsumerFactory{config: cfg}
}

// BuildFromTopics creates a consumer and handler map from instantiated topic configurations
func (f *KafkaConsumerFactory) BuildFromTopics(topicConfigs []InstantiatedTopicConfig) (*kafka.Consumer, map[string]kafka.MessageHandler, error) {
	// Read suffix config from YAML (empty is OK, consumer will use defaults)
	retrySuffix := f.config.Kafka.Retry.RetrySuffix
	dlqSuffix := f.config.Kafka.Retry.DLQSuffix

	// Build topics list (base + retry topics)
	topics := make([]string, 0, len(topicConfigs)*2)
	handlers := make(map[string]kafka.MessageHandler)
	topicRetryMap := make(map[string]kafka.TopicRetryConfig)

	for _, tc := range topicConfigs {
		// Add base topic
		topics = append(topics, tc.Name)
		handlers[tc.Name] = tc.Handler

		// If retry is enabled, add retry topic and config
		if tc.Retry != nil {
			retryTopic := tc.Name + retrySuffix
			topics = append(topics, retryTopic)
			handlers[retryTopic] = tc.Handler // Same handler for retry

			// Convert backoff from ms to time.Duration
			retryConfig := kafka.TopicRetryConfig{
				MaxAttempts: tc.Retry.MaxAttempts,
				Backoff:     time.Duration(tc.Retry.Backoff) * time.Millisecond,
			}
			topicRetryMap[tc.Name] = retryConfig
			topicRetryMap[retryTopic] = retryConfig
		}
	}

	// Create consumer (defaults will be applied in NewConsumerWithOptions if suffixes are empty)
	consumer, err := kafka.NewConsumerWithOptions(
		f.config.Kafka.Brokers,
		topics,
		kafka.ConsumerOptions{
			GroupID:          f.config.Kafka.GroupID,
			RetryTopicSuffix: retrySuffix,
			DLQTopicSuffix:   dlqSuffix,
			TopicRetryConfig: topicRetryMap,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return consumer, handlers, nil
}
