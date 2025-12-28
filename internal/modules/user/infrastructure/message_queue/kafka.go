package message_queue

import (
	"context"
	"strconv"

	"go1/internal/modules/user/domain"
	"go1/pkg/kafka"
)

type kafkaUserEvent struct {
	producer *kafka.KafkaProducer
}

func NewKafkaUserEvent(producer *kafka.KafkaProducer) domain.UserEvent {
	return &kafkaUserEvent{producer: producer}
}

func (k *kafkaUserEvent) PublishUserCreated(ctx context.Context, user *domain.User) error {
	// Publish struct directly - auto-marshaled to JSON!
	// Use user ID as partition key for consistent routing
	key := strconv.FormatInt(user.ID, 10)
	return k.producer.Publish("user_created", user, key)
}
