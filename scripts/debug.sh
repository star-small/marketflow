#!/bin/bash

echo "🔍 Debugging MarketFlow Aggregation Process"
echo "==========================================="

# 1. Check PostgreSQL data
echo "1. Checking PostgreSQL aggregated data..."
echo "Total records:"
sudo docker exec marketflow-postgres-1 psql -U marketflow -d marketflow -t -c "SELECT COUNT(*) FROM market_data;" | tr -d ' '

echo ""
echo "Recent records (last 5 minutes):"
sudo docker exec marketflow-postgres-1 psql -U marketflow -d marketflow -c "SELECT COUNT(*) FROM market_data WHERE timestamp > NOW() - INTERVAL '5 minutes';"

echo ""
echo "Latest entries:"
sudo docker exec marketflow-postgres-1 psql -U marketflow -d marketflow -c "SELECT exchange, pair_name, timestamp, average_price FROM market_data ORDER BY timestamp DESC LIMIT 5;"

echo ""
echo "2. Testing with longer periods..."
echo "Trying 10-minute period:"
curl -s "http://localhost:8080/prices/highest/BTCUSDT?period=10m" | jq . || echo "Response: $(curl -s 'http://localhost:8080/prices/highest/BTCUSDT?period=10m')"

echo ""
echo "3. Testing current data flow..."
echo "Current latest price:"
curl -s "http://localhost:8080/prices/latest/BTCUSDT" | jq .

echo ""
echo "4. Check if Redis has price history..."
echo "Redis keys pattern check:"
sudo docker exec marketflow-redis-1 redis-cli KEYS "history:*" | head -5

echo ""
echo "5. Recommendations:"
echo "✓ If PostgreSQL COUNT = 0: Aggregation not working"
echo "✓ If PostgreSQL COUNT > 0 but old data: Wait for new aggregation"
echo "✓ If Redis keys empty: Price history storage issue"
echo "✓ If 10m period works: Try longer periods like 15m, 30m"

echo ""
echo "🔄 Try these manual tests:"
echo "curl \"http://localhost:8080/prices/highest/BTCUSDT?period=15m\""
echo "curl \"http://localhost:8080/prices/average/BTCUSDT?period=30m\""
echo ""
echo "⏰ Wait for next minute mark (XX:17:00) and check logs for 'Saved aggregated data'"