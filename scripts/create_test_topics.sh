#!/bin/bash

# Create Kafka topics for testing retry/DLQ mechanism

KAFKA_CONTAINER="go1_kafka"
BOOTSTRAP_SERVER="localhost:9092"

echo "========================================="
echo "Creating Kafka Test Topics"
echo "========================================="
echo ""

# Function to create topic
create_topic() {
    local topic=$1
    local partitions=$2
    local replication=$3

    echo "Creating topic: $topic (partitions=$partitions, replication=$replication)"

    docker exec $KAFKA_CONTAINER kafka-topics.sh \
        --bootstrap-server $BOOTSTRAP_SERVER \
        --create \
        --topic $topic \
        --partitions $partitions \
        --replication-factor $replication \
        --if-not-exists

    echo ""
}

# Create base topics
echo "📝 Creating base topics..."
create_topic "test_success" 3 1
create_topic "test_retry" 3 1

# Create retry topics
echo "📝 Creating retry topics..."
create_topic "test_retry.retry" 3 1

# Create DLQ topics
echo "📝 Creating DLQ topics..."
create_topic "test_success.dlq" 1 1
create_topic "test_retry.dlq" 1 1

echo "========================================="
echo "✅ All test topics created!"
echo ""
echo "📋 List of topics:"
docker exec $KAFKA_CONTAINER kafka-topics.sh \
    --bootstrap-server $BOOTSTRAP_SERVER \
    --list | grep -E "(test_success|test_retry)"

echo ""
echo "========================================="
