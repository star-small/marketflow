#!/bin/bash

# MarketFlow Exchange Simulators - Load and Start Script
# Loads tar files and starts exchange containers

echo "🔄 MarketFlow Exchange Simulators - Load and Start"
echo "================================================="

# Check if in correct directory
if [ ! -d "exchanges" ]; then
    echo "❌ Error: exchanges/ directory not found"
    echo "Please run this script from the marketflow root directory"
    exit 1
fi

cd exchanges/

# Check for tar files
TAR_FILES=("exchange1_amd64.tar" "exchange2_amd64.tar" "exchange3_amd64.tar")
CONTAINER_NAMES=("exchange1-amd64" "exchange2-amd64" "exchange3-amd64")
IMAGE_NAMES=("exchange1" "exchange2" "exchange3")
PORTS=("40101" "40102" "40103")

echo "1. Checking for exchange tar files..."
for tar_file in "${TAR_FILES[@]}"; do
    if [ -f "$tar_file" ]; then
        echo "✅ Found $tar_file"
    else
        echo "❌ Missing $tar_file"
        exit 1
    fi
done

echo ""
echo "2. Stopping and removing existing containers..."
for container in "${CONTAINER_NAMES[@]}"; do
    echo "🛑 Stopping $container..."
    sudo docker stop "$container" 2>/dev/null || echo "  Container $container was not running"
    echo "🗑️  Removing $container..."
    sudo docker rm "$container" 2>/dev/null || echo "  Container $container was not found"
done

echo ""
echo "3. Loading Docker images from tar files..."
for tar_file in "${TAR_FILES[@]}"; do
    echo "📦 Loading $tar_file..."
    sudo docker load -i "$tar_file"
    if [ $? -eq 0 ]; then
        echo "✅ Successfully loaded $tar_file"
    else
        echo "❌ Failed to load $tar_file"
        exit 1
    fi
done

echo ""
echo "4. Checking loaded images..."
sudo docker images | grep exchange

echo ""
echo "5. Starting exchange containers..."

# Start Exchange 1
echo "🚀 Starting exchange1 on port 40101..."
sudo docker run -p 40101:40101 --name exchange1-amd64 -d exchange1

# Start Exchange 2
echo "🚀 Starting exchange2 on port 40102..."
sudo docker run -p 40102:40102 --name exchange2-amd64 -d exchange2

# Start Exchange 3
echo "🚀 Starting exchange3 on port 40103..."
sudo docker run -p 40103:40103 --name exchange3-amd64 -d exchange3

echo ""
echo "6. Waiting for exchanges to start..."
sleep 10

echo ""
echo "7. Verifying exchanges are running..."
for i in "${!CONTAINER_NAMES[@]}"; do
    container="${CONTAINER_NAMES[$i]}"
    port="${PORTS[$i]}"

    echo -n "📡 Testing $container on port $port... "

    if sudo docker ps | grep -q "$container"; then
        if timeout 5 nc -z 127.0.0.1 "$port" 2>/dev/null; then
            echo "✅ Running and responsive"
        else
            echo "⚠️  Running but not responsive yet (give it a moment)"
        fi
    else
        echo "❌ Failed to start"
        echo "   Checking logs:"
        sudo docker logs "$container" 2>/dev/null | tail -3
    fi
done

echo ""
echo "8. Testing data flow from exchanges..."
for i in "${!PORTS[@]}"; do
    port="${PORTS[$i]}"
    echo "📊 Sample data from exchange$((i+1)) (port $port):"
    timeout 3 nc 127.0.0.1 "$port" | head -2 2>/dev/null || echo "  No data yet (container still starting)"
    echo ""
done

echo ""
echo "9. Final container status:"
sudo docker ps --filter "name=exchange" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "✅ Exchange setup complete!"
echo ""
echo "🎯 Next steps:"
echo "1. Wait 30 seconds for full startup"
echo "2. Test direct exchange connection:"
echo "   nc 127.0.0.1 40101"
echo "3. Test your MarketFlow:"
echo "   curl http://localhost:8080/prices/latest/BTCUSDT"
echo "4. If MarketFlow still returns null, restart it:"
echo "   # Stop current MarketFlow (Ctrl+C)"
echo "   ./marketflow --port 8080"

echo ""
echo "🔍 Monitor live data:"
echo "   watch 'nc 127.0.0.1 40101 | head -1'"