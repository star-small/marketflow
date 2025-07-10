#!/bin/bash
# Save as monitor_pipeline_fixed.sh

echo "🔍 Real-time Pipeline Monitoring"
echo "================================="

# Quick pipeline test
test_pipeline() {
    echo "Testing data pipeline..."

    # Force test mode
    curl -s -X POST http://localhost:8080/mode/test > /dev/null
    echo "Switched to test mode"

    # Monitor for 30 seconds
    echo "Monitoring for 30 seconds..."
    START_KEYS=$(redis-cli -h localhost -p 6379 KEYS "*" | wc -l)

    for i in {1..15}; do
        sleep 2
        CURRENT_KEYS=$(redis-cli -h localhost -p 6379 KEYS "*" | wc -l)
        NEW_KEYS=$((CURRENT_KEYS - START_KEYS))
        echo "[$i] Keys: $CURRENT_KEYS (+$NEW_KEYS from start)"

        # Show what keys exist
        if [ "$CURRENT_KEYS" -gt 0 ]; then
            echo "    Existing keys:"
            redis-cli -h localhost -p 6379 KEYS "*" | head -3
        fi

        if [ "$NEW_KEYS" -gt 0 ]; then
            echo "🎯 Data is reaching Redis!"
            break
        fi
    done

    FINAL_KEYS=$(redis-cli -h localhost -p 6379 KEYS "*" | wc -l)
    TOTAL_NEW=$((FINAL_KEYS - START_KEYS))

    echo ""
    echo "Results:"
    echo "--------"
    echo "Start keys: $START_KEYS"
    echo "Final keys: $FINAL_KEYS"
    echo "New keys: $TOTAL_NEW"

    if [ "$TOTAL_NEW" -gt 0 ]; then
        echo "✅ Pipeline working - $TOTAL_NEW new keys created"
        echo "All keys:"
        redis-cli -h localhost -p 6379 KEYS "*"
        echo "Latest keys:"
        redis-cli -h localhost -p 6379 KEYS "latest:*"
    else
        echo "❌ Pipeline broken - no new keys created"
        echo "Current keys (if any):"
        redis-cli -h localhost -p 6379 KEYS "*"
    fi
}

test_pipeline