#!/bin/bash

echo "Deleting users-connector..."
curl -X DELETE http://localhost:8083/connectors/users-connector

echo ""
echo "Connector deleted!"
