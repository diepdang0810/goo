#!/bin/bash

echo "=== Kafka Connect Status ==="
curl -s http://localhost:8083/

echo ""
echo "=== Registered Connectors ==="
curl -s http://localhost:8083/connectors

echo ""
echo "=== Orders Connector Status ==="
curl -s http://localhost:8083/connectors/orders-connector/status

echo ""
echo "=== Connector Config ==="
curl -s http://localhost:8083/connectors/orders-connector
