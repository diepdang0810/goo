package kafka

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go1/pkg/logger"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	Client   *kgo.Client
	opts     ConsumerOptions
	producer *KafkaProducer
}

type MessageHandler func(context.Context, *kgo.Record) error

// TopicRetryConfig defines retry behavior for a specific topic.
type TopicRetryConfig struct {
	MaxAttempts int
	Backoff     time.Duration
}

// ConsumerOptions defines retry/DLQ behavior for the consumer.
type ConsumerOptions struct {
	// GroupID is used to initialize consumer group.
	GroupID string
	// RetryTopicSuffix will be appended to original topic for retry. Empty defaults to ".retry".
	RetryTopicSuffix string
	// DLQTopicSuffix will be appended to original topic for DLQ. Empty defaults to ".dlq".
	DLQTopicSuffix string
	// TopicRetryConfig maps topic names to their retry configs. If a topic is not in the map, retry is disabled.
	TopicRetryConfig map[string]TopicRetryConfig
}

func NewConsumer(brokers, groupID string, topics []string) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		Client: client,
		opts: ConsumerOptions{
			GroupID:          groupID,
			RetryTopicSuffix: ".retry",
			DLQTopicSuffix:   ".dlq",
			TopicRetryConfig: make(map[string]TopicRetryConfig),
		},
	}, nil
}

func NewConsumerWithOptions(brokers string, topics []string, options ConsumerOptions) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
		kgo.ConsumerGroup(options.GroupID),
		kgo.ConsumeTopics(topics...),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	prod, err := NewProducer(brokers)
	if err != nil {
		return nil, err
	}

	// Set defaults
	if options.RetryTopicSuffix == "" {
		options.RetryTopicSuffix = ".retry"
	}
	if options.DLQTopicSuffix == "" {
		options.DLQTopicSuffix = ".dlq"
	}
	if options.TopicRetryConfig == nil {
		options.TopicRetryConfig = make(map[string]TopicRetryConfig)
	}

	return &Consumer{Client: client, opts: options, producer: prod}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := c.Client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return nil
			}

			fetches.EachError(func(topic string, partition int32, err error) {
				logger.Log.Error("Fetch error",
					logger.Field{Key: "topic", Value: topic},
					logger.Field{Key: "partition", Value: partition},
					logger.Field{Key: "error", Value: err})
			})

			fetches.EachRecord(func(record *kgo.Record) {
				if err := handler(ctx, record); err != nil {
					c.handleError(ctx, record, err)
				}
			})
		}
	}
}

func (c *Consumer) handleError(ctx context.Context, record *kgo.Record, err error) {
	// Get retry config for this topic
	retryConfig, hasRetryConfig := c.opts.TopicRetryConfig[record.Topic]

	// Check if DLQ topic (never retry DLQ)
	if strings.HasSuffix(record.Topic, c.opts.DLQTopicSuffix) {
		logger.Log.Warn("DLQ topic error, not retrying",
			logger.Field{Key: "topic", Value: record.Topic},
			logger.Field{Key: "error", Value: err})
		return
	}

	// If no retry config for this topic, just log
	if !hasRetryConfig {
		logger.Log.Warn("No retry config for topic",
			logger.Field{Key: "topic", Value: record.Topic},
			logger.Field{Key: "error", Value: err})
		return
	}

	attempts := getAttempts(record.Headers)
	attempts++

	logger.Log.Error("Handler error",
		logger.Field{Key: "topic", Value: record.Topic},
		logger.Field{Key: "attempt", Value: attempts},
		logger.Field{Key: "maxAttempts", Value: retryConfig.MaxAttempts},
		logger.Field{Key: "error", Value: err})

	if c.producer == nil {
		return
	}

	// If max attempts reached, send to DLQ
	if attempts >= retryConfig.MaxAttempts {
		dlqTopic := getBaseTopic(record.Topic, c.opts.RetryTopicSuffix) + c.opts.DLQTopicSuffix
		headers := updateAttemptsHeader(record.Headers, attempts)
		if err := c.producer.Client.ProduceSync(ctx, &kgo.Record{
			Topic:   dlqTopic,
			Key:     record.Key,
			Value:   record.Value,
			Headers: headers,
		}).FirstErr(); err != nil {
			logger.Log.Error("Failed to publish to DLQ",
				logger.Field{Key: "dlqTopic", Value: dlqTopic},
				logger.Field{Key: "error", Value: err})
		} else {
			logger.Log.Info("Message sent to DLQ",
				logger.Field{Key: "dlqTopic", Value: dlqTopic},
				logger.Field{Key: "attempts", Value: attempts})
		}
		return
	}

	// Apply backoff before retry
	if retryConfig.Backoff > 0 {
		select {
		case <-time.After(retryConfig.Backoff):
		case <-ctx.Done():
			return
		}
	}

	// Send to retry topic
	retryTopic := getBaseTopic(record.Topic, c.opts.RetryTopicSuffix) + c.opts.RetryTopicSuffix
	headers := updateAttemptsHeader(record.Headers, attempts)
	if err := c.producer.Client.ProduceSync(ctx, &kgo.Record{
		Topic:   retryTopic,
		Key:     record.Key,
		Value:   record.Value,
		Headers: headers,
	}).FirstErr(); err != nil {
		logger.Log.Error("Failed to publish to retry topic",
			logger.Field{Key: "retryTopic", Value: retryTopic},
			logger.Field{Key: "error", Value: err})
	} else {
		logger.Log.Info("Message sent to retry topic",
			logger.Field{Key: "retryTopic", Value: retryTopic},
			logger.Field{Key: "attempts", Value: attempts})
	}
}

func (c *Consumer) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
	if c.producer != nil {
		c.producer.Close()
	}
}

// GetRetrySuffix returns the configured retry topic suffix
func (c *Consumer) GetRetrySuffix() string {
	if c.opts.RetryTopicSuffix == "" {
		return ".retry"
	}
	return c.opts.RetryTopicSuffix
}

// GetDLQSuffix returns the configured DLQ topic suffix
func (c *Consumer) GetDLQSuffix() string {
	if c.opts.DLQTopicSuffix == "" {
		return ".dlq"
	}
	return c.opts.DLQTopicSuffix
}

// getBaseTopic removes retry suffix if present to get the original topic name
func getBaseTopic(topic, retrySuffix string) string {
	if strings.HasSuffix(topic, retrySuffix) {
		return strings.TrimSuffix(topic, retrySuffix)
	}
	return topic
}

func getAttempts(hdrs []kgo.RecordHeader) int {
	for _, h := range hdrs {
		if strings.EqualFold(h.Key, "x-attempt") {
			if v, err := strconv.Atoi(string(h.Value)); err == nil {
				return v
			}
			break
		}
	}
	return 0
}

func updateAttemptsHeader(hdrs []kgo.RecordHeader, attempts int) []kgo.RecordHeader {
	updated := make([]kgo.RecordHeader, 0, len(hdrs)+1)
	// Copy all headers except x-attempt
	for _, h := range hdrs {
		if !strings.EqualFold(h.Key, "x-attempt") {
			updated = append(updated, h)
		}
	}
	// Add updated x-attempt header
	updated = append(updated, kgo.RecordHeader{Key: "x-attempt", Value: []byte(strconv.Itoa(attempts))})
	return updated
}
