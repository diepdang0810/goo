package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appConfig "go1/config"
	"go1/pkg/logger"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers string, producerConfig appConfig.KafkaProducerConfig) (*KafkaProducer, error) {
	config := sarama.NewConfig()

	// Configure RequiredAcks
	switch producerConfig.RequiredAcks {
	case "all":
		config.Producer.RequiredAcks = sarama.WaitForAll // Most durable - wait for all in-sync replicas
	case "local":
		config.Producer.RequiredAcks = sarama.WaitForLocal // Wait for leader only
	case "none":
		config.Producer.RequiredAcks = sarama.NoResponse // Fire and forget (fastest, least durable)
	default:
		config.Producer.RequiredAcks = sarama.WaitForAll // Default to most durable
	}

	// Configure retry
	config.Producer.Retry.Max = producerConfig.RetryMax
	if config.Producer.Retry.Max == 0 {
		config.Producer.Retry.Max = 5 // Default
	}

	// Configure compression
	switch producerConfig.Compression {
	case "gzip":
		config.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		config.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		config.Producer.Compression = sarama.CompressionZSTD
	default:
		config.Producer.Compression = sarama.CompressionNone
	}

	// Configure max message bytes
	if producerConfig.MaxMessageBytes > 0 {
		config.Producer.MaxMessageBytes = producerConfig.MaxMessageBytes
	}

	config.Producer.Return.Successes = true // Required for SyncProducer

	brokerList := strings.Split(brokers, ",")
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	logger.Log.Info("✅ Connected to Kafka producer",
		logger.Field{Key: "brokers", Value: brokers},
		logger.Field{Key: "requiredAcks", Value: producerConfig.RequiredAcks},
		logger.Field{Key: "retryMax", Value: producerConfig.RetryMax},
		logger.Field{Key: "compression", Value: producerConfig.Compression})

	return &KafkaProducer{producer: producer}, nil
}

// Publish publishes a message to Kafka with automatic logging and JSON marshaling
// topic: Kafka topic name
// message: Message payload ([]byte, string, or any struct - will auto-marshal to JSON)
// key: Optional partition key (pass empty or omit for default partitioning)
func (k *KafkaProducer) Publish(topic string, message any, key ...string) error {
	// Convert message to bytes
	var valueBytes []byte
	var err error

	switch v := message.(type) {
	case []byte:
		valueBytes = v
	case string:
		valueBytes = []byte(v)
	default:
		// Auto-marshal structs/maps to JSON
		valueBytes, err = json.Marshal(v)
		if err != nil {
			logger.Log.Error("❌ Failed to marshal message to JSON",
				logger.Field{Key: "topic", Value: topic},
				logger.Field{Key: "type", Value: fmt.Sprintf("%T", v)},
				logger.Field{Key: "error", Value: err})
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	// Handle optional key
	var keyBytes []byte
	if len(key) > 0 && key[0] != "" {
		keyBytes = []byte(key[0])
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(keyBytes),
		Value: sarama.ByteEncoder(valueBytes),
	}

	// Send message
	partition, offset, err := k.producer.SendMessage(msg)

	if err != nil {
		logger.Log.Error("❌ Failed to publish message",
			logger.Field{Key: "topic", Value: topic},
			logger.Field{Key: "key", Value: string(keyBytes)},
			logger.Field{Key: "value_size", Value: len(valueBytes)},
			logger.Field{Key: "error", Value: err})
		return err
	}

	// Auto-log success
	logger.Log.Info("📤 Message published",
		logger.Field{Key: "topic", Value: topic},
		logger.Field{Key: "partition", Value: partition},
		logger.Field{Key: "offset", Value: offset},
		logger.Field{Key: "key", Value: string(keyBytes)},
		logger.Field{Key: "value_size", Value: len(valueBytes)})

	return nil
}

// PublishWithHeaders publishes a message with custom headers (used internally for retry/DLQ)
func (k *KafkaProducer) PublishWithHeaders(ctx context.Context, topic string, key, value []byte, headers []sarama.RecordHeader) error {
	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.ByteEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: headers,
	}

	partition, offset, err := k.producer.SendMessage(msg)

	if err != nil {
		logger.Log.Error("❌ Failed to publish message with headers",
			logger.Field{Key: "topic", Value: topic},
			logger.Field{Key: "key", Value: string(key)},
			logger.Field{Key: "headers_count", Value: len(headers)},
			logger.Field{Key: "error", Value: err})
		return err
	}

	// Log with headers info
	logger.Log.Info("📤 Message published (with headers)",
		logger.Field{Key: "topic", Value: topic},
		logger.Field{Key: "partition", Value: partition},
		logger.Field{Key: "offset", Value: offset},
		logger.Field{Key: "key", Value: string(key)},
		logger.Field{Key: "headers_count", Value: len(headers)})

	// Log each header
	for _, h := range headers {
		if string(h.Key) == "x-attempt" {
			logger.Log.Info("  📋 Header",
				logger.Field{Key: "key", Value: string(h.Key)},
				logger.Field{Key: "value", Value: string(h.Value)})
		}
	}

	return nil
}

func (k *KafkaProducer) Close() {
	if k.producer != nil {
		k.producer.Close()
		logger.Log.Info("Kafka producer closed")
	}
}
