#!/bin/bash

# Test Kafka Retry/DLQ Mechanism
# This script publishes test messages to different topics

BASE_URL="http://localhost:8080"

echo "========================================="
echo "Kafka Retry/DLQ Test Script"
echo "========================================="
echo ""

# Function to publish messages
publish_messages() {
    local topic=$1
    local count=$2
    local message=$3

    echo "📤 Publishing $count messages to topic: $topic"
    echo "Message: $message"
    echo ""

    curl -X POST "$BASE_URL/api/v1/kafka-test/publish" \
        -H "Content-Type: application/json" \
        -d "{
            \"topic\": \"$topic\",
            \"count\": $count,
            \"message\": \"$message\"
        }" | jq '.'

    echo ""
    echo "---"
    echo ""
}

# Get available topics
echo "📋 Available Topics:"
curl -s "$BASE_URL/api/v1/kafka-test/topics" | jq '.'
echo ""
echo "========================================="
echo ""

# Test 1: Successful processing (no retry)
echo "TEST 1: Publishing to 'test_success' topic"
echo "Expected: Messages will be processed successfully without retry"
publish_messages "test_success" 10 "Test message for successful processing"

sleep 2

# Test 2: Retry then DLQ
echo "TEST 2: Publishing to 'test_retry' topic"
echo "Expected: Messages will fail, retry 3 times, then go to DLQ"
publish_messages "test_retry" 10 "Test message for retry/DLQ testing"

echo "========================================="
echo "✅ Test messages published!"
echo ""
echo "📝 Check worker logs to see:"
echo "  1. test_success: Messages processed successfully"
echo "  2. test_retry: Messages retry 3 times then go to DLQ"
echo ""
echo "🔍 To inspect DLQ messages, consume from topic: test_retry.dlq"
echo ""
echo "Commands:"
echo "  # View worker logs:"
echo "  docker logs -f go1_worker"
echo ""
echo "  # Consume from DLQ:"
echo "  docker exec -it go1_kafka kafka-console-consumer.sh \\"
echo "    --bootstrap-server localhost:9092 \\"
echo "    --topic test_retry.dlq \\"
echo "    --from-beginning \\"
echo "    --property print.headers=true"
echo "========================================="
