package kafka

import (
	"context"
	"strings"
	"time"

	"go1/pkg/logger"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaProducer struct {
	Client *kgo.Client
}

func NewProducer(brokers string) (*KafkaProducer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	// Ping to check connection (optional, franz-go connects lazily but good for startup check)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx); err != nil {
		// Just log warning, kafka might not be ready yet
		logger.Log.Warn("Kafka ping failed", logger.Field{Key: "error", Value: err})
	} else {
		logger.Log.Info("Connected to Kafka", logger.Field{Key: "brokers", Value: brokers})
	}

	return &KafkaProducer{Client: client}, nil
}

func (k *KafkaProducer) Publish(ctx context.Context, topic string, key, value []byte) error {
	record := &kgo.Record{Topic: topic, Key: key, Value: value}
	return k.Client.ProduceSync(ctx, record).FirstErr()
}

func (k *KafkaProducer) Close() {
	if k.Client != nil {
		k.Client.Close()
	}
}
