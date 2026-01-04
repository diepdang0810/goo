package worker

import (
	"go1/internal/shared/order/infrastructure/repository"
	consumers "go1/internal/worker/consumers"
	"go1/pkg/postgres"

	"go.temporal.io/sdk/client"
)

func (b *WorkerBuilder) WithShipmentEvents(pg *postgres.Postgres, temporalClient client.Client) *WorkerBuilder {
	repo := repository.NewPostgresOrderRepository(pg.Pool)
	handler := consumers.NewShipmentConsumer(temporalClient, repo)
	return b.AddTopic(b.config.Kafka.Topics.ShipmentEvents, handler.Handle())
}

func (b *WorkerBuilder) WithDispatchEvents(pg *postgres.Postgres, temporalClient client.Client) *WorkerBuilder {
	repo := repository.NewPostgresOrderRepository(pg.Pool)
	handler := consumers.NewDispatchConsumer(temporalClient, repo)
	return b.AddTopic(b.config.Kafka.Topics.DispatchEvents, handler.Handle())
}

func (b *WorkerBuilder) WithOrderEvents(temporalClient client.Client) *WorkerBuilder {
	handler := consumers.NewOrderConsumer(temporalClient)
	return b.AddTopic(b.config.Kafka.Topics.OrderEvents, handler.Handle())
}
