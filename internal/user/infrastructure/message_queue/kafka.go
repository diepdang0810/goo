package message_queue

import (
	"context"
	"encoding/json"
	"strconv"

	"go1/internal/user/domain"
	"go1/pkg/kafka"
)

type kafkaUserEvent struct {
	producer *kafka.KafkaProducer
}

func NewKafkaUserEvent(producer *kafka.KafkaProducer) domain.UserEvent {
	return &kafkaUserEvent{producer: producer}
}

func (k *kafkaUserEvent) PublishUserCreated(ctx context.Context, user *domain.User) error {
	eventData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return k.producer.Publish(ctx, "user_created", []byte(strconv.FormatInt(user.ID, 10)), eventData)
}
