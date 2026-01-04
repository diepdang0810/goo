#!/bin/bash

echo "Deleting orders-connector..."
curl -X DELETE http://localhost:8083/connectors/orders-connector

echo ""
echo "Connector deleted!"
