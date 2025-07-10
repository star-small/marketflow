#!/bin/bash

# Comprehensive integration test script
set -e

echo "Running MarketFlow Integration Tests..."

# Configuration
BASE_URL="http://localhost:8080"
TIMEOUT=30

# Function to wait for service
wait_for_service() {
    local url=$1
    local timeout=$2
    local count=0

    echo "Waiting for service at $url..."
    while [ $count -lt $timeout ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            echo "Service is ready!"
            return 0
        fi
        sleep 1
        count=$((count + 1))
    done

    echo "Service failed to start within $timeout seconds"
    return 1
}

# Function to test API endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=${3:-200}
    local description=$4

    echo "Testing: $description"
    echo "  $method $endpoint"

    response=$(curl -s -w "%{http_code}" -X "$method" "$BASE_URL$endpoint")
    status_code="${response: -3}"
    body="${response%???}"

    if [ "$status_code" -eq "$expected_status" ]; then
        echo "  ✓ Status: $status_code"
        if [ ! -z "$body" ] && [ "$body" != "null" ]; then
            echo "  ✓ Response: $(echo "$body" | jq -c . 2>/dev/null || echo "$body")"
        fi
    else
        echo "  ✗ Expected status $expected_status, got $status_code"
        echo "  ✗ Response: $body"
        return 1
    fi
}

# Start services if not running
echo "Ensuring services are running..."
if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    echo "Starting MarketFlow services..."
    make docker-up
    make build
    ./build/marketflow &
    APP_PID=$!

    # Wait for application to start
    wait_for_service "$BASE_URL/health" $TIMEOUT
else
    echo "Services already running"
fi

# Test sequence
echo ""
echo "=== Testing Health Endpoints ==="
test_endpoint "GET" "/health" 200 "Basic health check"
test_endpoint "GET" "/health/detailed" 200 "Detailed health check"

echo ""
echo "=== Testing Mode Management ==="
test_endpoint "GET" "/mode/current" 200 "Get current mode"
test_endpoint "POST" "/mode/test" 200 "Switch to test mode"
test_endpoint "GET" "/mode/current" 200 "Verify test mode"

# Wait for test data generation
echo ""
echo "Waiting for test data generation..."
sleep 10

echo ""
echo "=== Testing Latest Price Endpoints ==="
for symbol in BTCUSDT ETHUSDT DOGEUSDT TONUSDT SOLUSDT; do
    test_endpoint "GET" "/prices/latest/$symbol" 200 "Latest price for $symbol"
done

echo ""
echo "=== Testing Price Statistics (Short Period) ==="
for symbol in BTCUSDT ETHUSDT; do
    test_endpoint "GET" "/prices/highest/$symbol?period=30s" 200 "Highest $symbol (30s)"
    test_endpoint "GET" "/prices/lowest/$symbol?period=30s" 200 "Lowest $symbol (30s)"
    test_endpoint "GET" "/prices/average/$symbol?period=30s" 200 "Average $symbol (30s)"
done

echo ""
echo "=== Testing Exchange-Specific Endpoints ==="
for exchange in test-exchange1 test-exchange2; do
    test_endpoint "GET" "/prices/latest/$exchange/BTCUSDT" 200 "Latest BTCUSDT from $exchange"
done

echo ""
echo "=== Testing Error Cases ==="
test_endpoint "GET" "/prices/latest/INVALID" 400 "Invalid symbol"
test_endpoint "GET" "/prices/latest/exchange/INVALID" 400 "Invalid symbol with exchange"
test_endpoint "GET" "/invalid/endpoint" 404 "Invalid endpoint"

echo ""
echo "=== Testing Mode Switch ==="
test_endpoint "POST" "/mode/live" 200 "Switch to live mode"
sleep 2
test_endpoint "POST" "/mode/test" 200 "Switch back to test mode"

# Performance test
echo ""
echo "=== Performance Test ==="
echo "Running 100 concurrent requests..."
time bash -c 'for i in {1..100}; do curl -s "$BASE_URL/prices/latest/BTCUSDT" & done; wait'

# Cleanup
if [ ! -z "$APP_PID" ]; then
    echo ""
    echo "Stopping test application..."
    kill $APP_PID 2>/dev/null || true
fi

echo ""
echo "✓ All integration tests completed successfully!"