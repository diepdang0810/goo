#!/bin/bash

echo "=== Kafka Connect Status ==="
curl -s http://localhost:8083/ | jq .

echo ""
echo "=== Registered Connectors ==="
curl -s http://localhost:8083/connectors | jq .

echo ""
echo "=== Users Connector Status ==="
curl -s http://localhost:8083/connectors/users-connector/status | jq .

echo ""
echo "=== Connector Config ==="
curl -s http://localhost:8083/connectors/users-connector | jq .
