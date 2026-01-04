package kafka

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	appConfig "go1/config"
	"go1/pkg/logger"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	opts          ConsumerOptions
	producer      *KafkaProducer
	topics        []string
}

type MessageHandler func(context.Context, *sarama.ConsumerMessage) error

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

func NewConsumerWithOptions(brokers string, topics []string, options ConsumerOptions, producerConfig appConfig.KafkaProducerConfig, consumerConfig appConfig.KafkaConsumerConfig) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}

	// Apply consumer config
	if consumerConfig.SessionTimeoutMs > 0 {
		config.Consumer.Group.Session.Timeout = time.Duration(consumerConfig.SessionTimeoutMs) * time.Millisecond
	}
	if consumerConfig.HeartbeatIntervalMs > 0 {
		config.Consumer.Group.Heartbeat.Interval = time.Duration(consumerConfig.HeartbeatIntervalMs) * time.Millisecond
	}
	if consumerConfig.MaxProcessingTimeMs > 0 {
		config.Consumer.MaxProcessingTime = time.Duration(consumerConfig.MaxProcessingTimeMs) * time.Millisecond
	}

	brokerList := strings.Split(brokers, ",")
	consumerGroup, err := sarama.NewConsumerGroup(brokerList, options.GroupID, config)
	if err != nil {
		return nil, err
	}

	prod, err := NewProducer(brokers, producerConfig)
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

	return &Consumer{
		consumerGroup: consumerGroup,
		topics:        topics,
		opts:          options,
		producer:      prod,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	consumerHandler := &consumerGroupHandler{
		consumer: c,
		handler:  handler,
		ready:    make(chan bool),
	}

	// Start error handling goroutine
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range c.consumerGroup.Errors() {
			logger.Log.Error("Consumer group error", logger.Field{Key: "error", Value: err})
		}
	}()

	// Consume runs in a loop to handle rebalancing
	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		if err := c.consumerGroup.Consume(ctx, c.topics, consumerHandler); err != nil {
			logger.Log.Error("Consumer group error", logger.Field{Key: "error", Value: err})
			return err
		}

		// Check if context was cancelled
		if ctx.Err() != nil {
			wg.Wait()
			return ctx.Err()
		}

		// Reset ready channel for next session
		consumerHandler.ready = make(chan bool)
	}
}

func (c *Consumer) handleError(ctx context.Context, message *sarama.ConsumerMessage, err error) {
	// Get retry config for this topic
	retryConfig, hasRetryConfig := c.opts.TopicRetryConfig[message.Topic]

	// Check if DLQ topic (never retry DLQ)
	if strings.HasSuffix(message.Topic, c.opts.DLQTopicSuffix) {
		logger.Log.Warn("DLQ topic error, not retrying",
			logger.Field{Key: "topic", Value: message.Topic},
			logger.Field{Key: "error", Value: err})
		return
	}

	// If no retry config for this topic, just log
	if !hasRetryConfig {
		logger.Log.Warn("No retry config for topic",
			logger.Field{Key: "topic", Value: message.Topic},
			logger.Field{Key: "error", Value: err})
		return
	}

	attempts := getAttempts(message.Headers)
	attempts++

	logger.Log.Error("Handler error",
		logger.Field{Key: "topic", Value: message.Topic},
		logger.Field{Key: "attempt", Value: attempts},
		logger.Field{Key: "maxAttempts", Value: retryConfig.MaxAttempts},
		logger.Field{Key: "error", Value: err})

	if c.producer == nil {
		return
	}

	// Update headers with new attempt count
	updatedHeaders := updateAttemptsHeader(message.Headers, attempts)

	// If max attempts reached, send to DLQ
	if attempts >= retryConfig.MaxAttempts {
		c.sendToDLQ(ctx, message, updatedHeaders, attempts)
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
	c.sendToRetry(ctx, message, updatedHeaders, attempts)
}

func (c *Consumer) sendToDLQ(ctx context.Context, message *sarama.ConsumerMessage, headers []sarama.RecordHeader, attempts int) {
	dlqTopic := getBaseTopic(message.Topic, c.opts.RetryTopicSuffix) + c.opts.DLQTopicSuffix

	if err := c.producer.PublishWithHeaders(ctx, dlqTopic, message.Key, message.Value, headers); err != nil {
		logger.Log.Error("Failed to publish to DLQ",
			logger.Field{Key: "dlqTopic", Value: dlqTopic},
			logger.Field{Key: "error", Value: err})
	} else {
		logger.Log.Info("Message sent to DLQ",
			logger.Field{Key: "dlqTopic", Value: dlqTopic},
			logger.Field{Key: "attempts", Value: attempts})
	}
}

func (c *Consumer) sendToRetry(ctx context.Context, message *sarama.ConsumerMessage, headers []sarama.RecordHeader, attempts int) {
	retryTopic := getBaseTopic(message.Topic, c.opts.RetryTopicSuffix) + c.opts.RetryTopicSuffix
	if err := c.producer.PublishWithHeaders(ctx, retryTopic, message.Key, message.Value, headers); err != nil {
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
	if c.consumerGroup != nil {
		c.consumerGroup.Close()
	}
	if c.producer != nil {
		c.producer.Close()
	}
}

// getBaseTopic removes retry suffix if present to get the original topic name
func getBaseTopic(topic, retrySuffix string) string {
	if strings.HasSuffix(topic, retrySuffix) {
		return strings.TrimSuffix(topic, retrySuffix)
	}
	return topic
}

func getAttempts(hdrs []*sarama.RecordHeader) int {
	for _, h := range hdrs {
		if strings.EqualFold(string(h.Key), "x-attempt") {
			if v, err := strconv.Atoi(string(h.Value)); err == nil {
				return v
			}
			break
		}
	}
	return 0
}

func updateAttemptsHeader(hdrs []*sarama.RecordHeader, attempts int) []sarama.RecordHeader {
	updated := make([]sarama.RecordHeader, 0, len(hdrs)+1)
	// Copy all headers except x-attempt
	for _, h := range hdrs {
		if !strings.EqualFold(string(h.Key), "x-attempt") {
			updated = append(updated, *h)
		}
	}
	// Add updated x-attempt header
	updated = append(updated, sarama.RecordHeader{
		Key:   []byte("x-attempt"),
		Value: []byte(strconv.Itoa(attempts)),
	})
	return updated
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *Consumer
	handler  MessageHandler
	ready    chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	// Close ready channel to signal that consumer is ready
	close(h.ready)
	logger.Log.Info("Consumer group session setup completed")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Log.Info("Consumer group session cleanup completed")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// NOTE: Do not move the code below to a goroutine.
// The `ConsumeClaim` itself is called within a goroutine by sarama for each partition, see:
// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
// This ensures each partition processes messages sequentially to maintain ordering.
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger.Log.Info("Starting to consume partition",
		logger.Field{Key: "topic", Value: claim.Topic()},
		logger.Field{Key: "partition", Value: claim.Partition()},
		logger.Field{Key: "initialOffset", Value: claim.InitialOffset()})

	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				logger.Log.Info("Message channel closed",
					logger.Field{Key: "topic", Value: claim.Topic()},
					logger.Field{Key: "partition", Value: claim.Partition()})
				return nil
			}

			// Process message sequentially to maintain ordering within partition
			if err := h.handler(session.Context(), message); err != nil {
				h.consumer.handleError(session.Context(), message, err)
			}

			// Mark message as processed (commit offset)
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			logger.Log.Info("Session context cancelled, exiting consume loop",
				logger.Field{Key: "topic", Value: claim.Topic()},
				logger.Field{Key: "partition", Value: claim.Partition()})
			return nil
		}
	}
}
