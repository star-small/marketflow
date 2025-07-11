#!/bin/bash
# pipeline_test.sh - Quick diagnostic for MarketFlow pipeline

echo "🔍 MarketFlow Pipeline Diagnostic"
echo "================================="

# Step 1: Clear Redis and check baseline
echo "1. Clearing Redis and checking baseline..."
redis-cli -h localhost -p 6379 FLUSHALL > /dev/null
BASELINE_KEYS=$(redis-cli -h localhost -p 6379 KEYS "*" | wc -l)
echo "   Baseline Redis keys: $BASELINE_KEYS"

# Step 2: Switch to test mode
echo "2. Switching to test mode..."
SWITCH_RESPONSE=$(curl -s -X POST http://localhost:8080/mode/test)
echo "   Response: $SWITCH_RESPONSE"

# Step 3: Monitor for 30 seconds
echo "3. Monitoring pipeline for 30 seconds..."
for i in {1..6}; do
    sleep 5
    CURRENT_KEYS=$(redis-cli -h localhost -p 6379 KEYS "*" | wc -l)
    NEW_KEYS=$((CURRENT_KEYS - BASELINE_KEYS))

    # Check health
    HEALTH=$(curl -s http://localhost:8080/health | jq -r '.status')
    BUFFER=$(curl -s http://localhost:8080/health/detailed | jq -r '.system_info.exchange_stats.result_buffer')

    echo "   [$((i*5))s] Keys: $CURRENT_KEYS (+$NEW_KEYS) | Health: $HEALTH | Buffer: $BUFFER"

    if [ "$NEW_KEYS" -gt 0 ]; then
        echo "   🎯 Data detected in Redis!"
        break
    fi
done

# Step 4: Final analysis
echo "4. Final analysis..."
FINAL_KEYS=$(redis-cli -h localhost -p 6379 KEYS "*" | wc -l)
TOTAL_NEW=$((FINAL_KEYS - BASELINE_KEYS))

echo "   Total new keys created: $TOTAL_NEW"

if [ "$TOTAL_NEW" -gt 0 ]; then
    echo "   ✅ Pipeline working!"
    echo "   Key types:"
    redis-cli -h localhost -p 6379 KEYS "*" | head -5

    # Test API endpoints
    echo "5. Testing API endpoints..."
    for exchange in test-exchange1 test-exchange2 test-exchange3; do
        RESPONSE=$(curl -s "http://localhost:8080/prices/latest/$exchange/BTCUSDT")
        if echo "$RESPONSE" | jq -e '.price' > /dev/null 2>&1; then
            PRICE=$(echo "$RESPONSE" | jq -r '.price')
            echo "   ✅ $exchange: $PRICE"
        else
            ERROR=$(echo "$RESPONSE" | jq -r '.message // "unknown error"')
            echo "   ❌ $exchange: $ERROR"
        fi
    done
else
    echo "   ❌ Pipeline broken - no data reaching Redis"
    echo "   Check application logs for 'Result channel full' warnings"
fi