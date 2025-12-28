#!/bin/bash

# Wait for Kafka Connect to be ready
echo "Waiting for Kafka Connect to be ready..."
until curl -s http://localhost:8083/connectors > /dev/null 2>&1; do
  echo "Kafka Connect is not ready yet. Retrying in 5 seconds..."
  sleep 5
done

echo "Kafka Connect is ready!"

# Register Debezium PostgreSQL connector
echo "Registering Debezium PostgreSQL connector..."

curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "users-connector",
    "config": {
      "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
      "tasks.max": "1",
      "database.hostname": "postgres",
      "database.port": "5432",
      "database.user": "postgres",
      "database.password": "postgres",
      "database.dbname": "go1_db",
      "database.server.name": "go1",
      "table.include.list": "public.users",
      "topic.prefix": "dbserver1",
      "plugin.name": "pgoutput",
      "publication.autocreate.mode": "filtered",
      "slot.name": "debezium_slot",
      "heartbeat.interval.ms": "10000",
      "key.converter": "org.apache.kafka.connect.json.JsonConverter",
      "key.converter.schemas.enable": "false",
      "value.converter": "org.apache.kafka.connect.json.JsonConverter",
      "value.converter.schemas.enable": "false",
      "transforms": "unwrap",
      "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
      "transforms.unwrap.drop.tombstones": "false",
      "transforms.unwrap.delete.handling.mode": "rewrite"
    }
  }'

echo ""
echo "Connector registered successfully!"
echo ""
echo "Check connector status:"
echo "curl http://localhost:8083/connectors/users-connector/status"
