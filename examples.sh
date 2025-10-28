#!/bin/bash

# Example API requests for testing the api-pubsub-logger

echo "=== API Request Examples ==="
echo ""

echo "1. Health Check (not logged):"
echo "curl http://localhost:8080/health"
curl http://localhost:8080/health
echo ""
echo ""

echo "2. Get all items:"
echo "curl http://localhost:8080/v1/items"
curl http://localhost:8080/v1/items
echo ""
echo ""

echo "3. Create a new item:"
echo 'curl -X POST http://localhost:8080/v1/items \'
echo '  -H "Content-Type: application/json" \'
echo '  -H "X-User-ID: user-123" \'
echo '  -d '"'"'{'
echo '    "name": "Test Item",'
echo '    "description": "A test item",'
echo '    "email": "test@example.com",'
echo '    "phone_number": "+1-555-1234"'
echo '  }'"'"
curl -X POST http://localhost:8080/v1/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user-123" \
  -d '{
    "name": "Test Item",
    "description": "A test item",
    "email": "test@example.com",
    "phone_number": "+1-555-1234"
  }'
echo ""
echo ""

echo "4. Create item with custom Request ID:"
echo 'curl -X POST http://localhost:8080/v1/items \'
echo '  -H "Content-Type: application/json" \'
echo '  -H "X-User-ID: user-456" \'
echo '  -H "X-Request-ID: custom-req-123" \'
echo '  -d '"'"'{'
echo '    "name": "Another Item",'
echo '    "description": "Item with custom request ID"'
echo '  }'"'"
curl -X POST http://localhost:8080/v1/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user-456" \
  -H "X-Request-ID: custom-req-123" \
  -d '{
    "name": "Another Item",
    "description": "Item with custom request ID"
  }'
echo ""
echo ""

echo "=== Check the Pub/Sub subscriber terminal to see the logged events! ==="
