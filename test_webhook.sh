#!/bin/bash

# Test script to verify webhook signature verification works

SECRET="test-secret-123"
PAYLOAD='{"action":"completed","workflow_run":{"id":123,"name":"Test","status":"completed","conclusion":"success","run_started_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T12:10:00Z","head_branch":"main","event":"push","head_sha":"abc123"},"repository":{"full_name":"owner/repo"}}'

# Generate signature
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" -binary | xxd -p -c 256)

echo "Testing webhook with valid signature..."
echo "Payload: $PAYLOAD"
echo "Secret: $SECRET"
echo "Generated signature: sha256=$SIGNATURE"

# Start server in background
GITHUB_WEBHOOK_SECRET="$SECRET" ./gh-actions-exporter &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo ""
echo "=== Testing with valid signature ==="
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: workflow_run" \
  -H "X-Hub-Signature-256: sha256=$SIGNATURE" \
  -d "$PAYLOAD" \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "=== Testing with invalid signature ==="
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: workflow_run" \
  -H "X-Hub-Signature-256: sha256=invalid" \
  -d "$PAYLOAD" \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "=== Testing without signature ==="
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: workflow_run" \
  -d "$PAYLOAD" \
  -w "\nHTTP Status: %{http_code}\n"

# Clean up
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo ""
echo "Test completed!"
