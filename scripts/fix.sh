#!/bin/bash

echo "🔄 Fixing MarketFlow Data Processing Pipeline"
echo "============================================="

# 1. Check current status
echo "1. Current system status:"
curl -s http://localhost:8080/status | jq .
curl -s http://localhost:8080/health | jq .

echo ""
echo "2. Restart the data pipeline by switching modes:"

# Force restart by switching modes
echo "Switching to test mode..."
curl -s -X POST http://localhost:8080/mode/test | jq .

echo ""
echo "⏳ Waiting 10 seconds for test data to start..."
sleep 10

echo ""
echo "3. Check if real-time data is flowing:"
curl -s http://localhost:8080/prices/latest/BTCUSDT | jq .

echo ""
echo "4. Switch to live mode and test:"
curl -s -X POST http://localhost:8080/mode/live | jq .

echo ""
echo "⏳ Waiting 10 seconds for live data..."
sleep 10

echo ""
echo "5. Test live data flow:"
curl -s http://localhost:8080/prices/latest/BTCUSDT | jq .

echo ""
echo "6. Wait for next aggregation (watch for 'Saved aggregated data' in logs):"
echo "The aggregation runs every minute at XX:X7:00, XX:X8:00, etc."
echo ""
echo "🔍 Check your MarketFlow server logs for:"
echo "- 'Starting data aggregation'"
echo "- 'Saved aggregated data count=X'"
echo "- Any error messages"

echo ""
echo "7. After 1-2 minutes, test short periods:"
echo "curl \"http://localhost:8080/prices/highest/BTCUSDT?period=2m\""
echo "curl \"http://localhost:8080/prices/average/ETHUSDT?period=3m\""