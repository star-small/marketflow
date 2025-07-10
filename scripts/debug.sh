#!/bin/bash
# Save as db_debug.sh and run: chmod +x db_debug.sh && ./db_debug.sh

echo "🔍 Database and Cache Pipeline Debug"
echo "===================================="

# 1. Check PostgreSQL connection and data
echo "1. PostgreSQL Database Check"
echo "----------------------------"

# Check if PostgreSQL container is running
if docker ps --format '{{.Names}}' | grep -q "db\|postgres"; then
    DB_CONTAINER=$(docker ps --format '{{.Names}}' | grep -E "db|postgres" | head -1)
    echo "✅ Found PostgreSQL container: $DB_CONTAINER"

    # Check database connection
    echo "Testing database connection..."
    if docker exec $DB_CONTAINER pg_isready -U postgres; then
        echo "✅ Database is ready"

        # Check if prices table exists
        echo "Checking prices table..."
        TABLE_EXISTS=$(docker exec $DB_CONTAINER psql -U postgres -d market -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'prices';" 2>/dev/null | tr -d ' ')

        if [ "$TABLE_EXISTS" = "1" ]; then
            echo "✅ Prices table exists"

            # Check table structure
            echo "Table structure:"
            docker exec $DB_CONTAINER psql -U postgres -d market -c "\d prices"

            # Check if there's any data
            echo "Checking for price data in database..."
            ROW_COUNT=$(docker exec $DB_CONTAINER psql -U postgres -d market -t -c "SELECT COUNT(*) FROM prices;" 2>/dev/null | tr -d ' ')
            echo "Total rows in prices table: $ROW_COUNT"

            if [ "$ROW_COUNT" -gt 0 ]; then
                echo "Recent price data:"
                docker exec $DB_CONTAINER psql -U postgres -d market -c "SELECT * FROM prices ORDER BY timestamp DESC LIMIT 5;"
            else
                echo "❌ No data in prices table"
            fi
        else
            echo "❌ Prices table does not exist"
        fi
    else
        echo "❌ Database connection failed"
    fi
else
    echo "❌ PostgreSQL container not found"
fi

# 2. Redis Cache Detailed Check
echo -e "\n2. Redis Cache Detailed Check"
echo "------------------------------"

if command -v redis-cli &> /dev/null; then
    echo "Testing Redis connection..."
    if redis-cli -h localhost -p 6379 ping > /dev/null; then
        echo "✅ Redis is responding"

        # Check all keys
        echo "All Redis keys:"
        redis-cli -h localhost -p 6379 KEYS "*"

        # Check for specific pattern keys
        echo -e "\nLatest price keys:"
        redis-cli -h localhost -p 6379 KEYS "latest:*"

        echo -e "\nTimeseries keys:"
        redis-cli -h localhost -p 6379 KEYS "timeseries:*"

        # Check memory usage
        echo -e "\nRedis info:"
        redis-cli -h localhost -p 6379 INFO memory | grep used_memory_human
        redis-cli -h localhost -p 6379 INFO keyspace

        # If there are any latest keys, show their values
        LATEST_KEYS=$(redis-cli -h localhost -p 6379 KEYS "latest:*")
        if [ ! -z "$LATEST_KEYS" ]; then
            echo -e "\nSample latest price data:"
            echo "$LATEST_KEYS" | head -3 | while read key; do
                if [ ! -z "$key" ]; then
                    echo "$key: $(redis-cli -h localhost -p 6379 GET "$key")"
                fi
            done
        fi
    else
        echo "❌ Redis connection failed"
    fi
else
    echo "❌ redis-cli not available"
fi

# 3. Application Data Pipeline Test
echo -e "\n3. Application Data Pipeline Test"
echo "----------------------------------"

# Force test mode and monitor
echo "Forcing test mode switch..."
curl -s -X POST http://localhost:8080/mode/test > /dev/null

# Wait and check for immediate data
echo "Waiting 10 seconds for data generation..."
sleep 10

# Check if any new keys appeared
echo "Checking for new Redis keys after test mode..."
NEW_KEYS=$(redis-cli -h localhost -p 6379 KEYS "*" | wc -l)
echo "Total Redis keys now: $NEW_KEYS"

# 4. Application Health Deep Dive
echo -e "\n4. Application Health Deep Dive"
echo "--------------------------------"

DETAILED_HEALTH=$(curl -s http://localhost:8080/health/detailed)
echo "Database health: $(echo $DETAILED_HEALTH | jq -r '.checks.database.status')"
echo "Cache health: $(echo $DETAILED_HEALTH | jq -r '.checks.cache.status')"
echo "Exchange service health: $(echo $DETAILED_HEALTH | jq -r '.checks.exchange_service.status')"

# Exchange stats breakdown
echo -e "\nExchange service stats:"
echo $DETAILED_HEALTH | jq -r '.system_info.exchange_stats'

# 5. Direct API Test
echo -e "\n5. Direct API Test"
echo "------------------"

# Test multiple symbols
for symbol in BTCUSDT ETHUSDT DOGEUSDT; do
    echo "Testing $symbol..."
    RESPONSE=$(curl -s "http://localhost:8080/prices/latest/$symbol")
    if echo $RESPONSE | jq -e '.price' > /dev/null 2>&1; then
        PRICE=$(echo $RESPONSE | jq -r '.price')
        echo "✅ $symbol: $PRICE"
    else
        ERROR=$(echo $RESPONSE | jq -r '.message // "unknown error"')
        echo "❌ $symbol: $ERROR"
    fi
done

# 6. Manual Cache Test
echo -e "\n6. Manual Cache Test"
echo "--------------------"

# Set a test key to verify cache is working
echo "Setting test key in Redis..."
redis-cli -h localhost -p 6379 SET "test:manual" "hello" EX 60
GET_RESULT=$(redis-cli -h localhost -p 6379 GET "test:manual")
echo "Test key result: $GET_RESULT"

if [ "$GET_RESULT" = "hello" ]; then
    echo "✅ Redis read/write working"
else
    echo "❌ Redis read/write failed"
fi

# Clean up test key
redis-cli -h localhost -p 6379 DEL "test:manual"

echo -e "\n7. Diagnosis Summary"
echo "===================="

# Check key components
DB_OK=$(docker exec $(docker ps --format '{{.Names}}' | grep -E "db|postgres" | head -1) pg_isready -U postgres 2>/dev/null && echo "yes" || echo "no")
REDIS_OK=$(redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -q PONG && echo "yes" || echo "no")
APP_OK=$(curl -s http://localhost:8080/health | jq -r '.status')

echo "Database OK: $DB_OK"
echo "Redis OK: $REDIS_OK"
echo "App Health: $APP_OK"
echo "Total Redis Keys: $(redis-cli -h localhost -p 6379 KEYS "*" | wc -l)"
echo "Price Keys: $(redis-cli -h localhost -p 6379 KEYS "latest:*" | wc -l)"

if [ "$NEW_KEYS" -gt 1 ]; then
    echo "🎯 ISSUE: Data is being generated but not in expected format"
else
    echo "🎯 ISSUE: No data is being generated by adapters"
fi