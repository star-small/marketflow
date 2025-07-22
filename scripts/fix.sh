#!/bin/bash

echo "🔄 Setting up fresh database..."

# 1. Stop containers and remove volumes to start fresh
echo "🛑 Stopping containers and removing old data..."
sudo docker-compose down --volumes --remove-orphans

# 2. Remove any dangling volumes
echo "🧹 Cleaning up volumes..."
sudo docker volume prune -f

# 3. Start containers fresh (this will trigger init.sql)
echo "🚀 Starting fresh containers..."
sudo docker-compose up -d

# 4. Wait for postgres to fully initialize
echo "⏳ Waiting for PostgreSQL initialization (30 seconds)..."
sleep 30

# 5. Check if initialization worked
echo "🔍 Checking database setup..."
sudo docker-compose logs postgres | tail -20

# 6. Test connection
echo "🧪 Testing database connection..."
sudo docker exec marketflow-postgres-1 psql -U marketflow -d marketflow -c "SELECT 'Connection successful!' as status;"

if [ $? -eq 0 ]; then
    echo "✅ Database setup successful!"
    echo "🎯 Now try running your application:"
    echo "  ./marketflow --port 8080"
else
    echo "❌ Database setup failed. Let's try manual setup..."

    # Manual user creation as fallback
    echo "🔧 Creating user manually..."
    sudo docker exec marketflow-postgres-1 psql -U postgres -c "
        CREATE USER marketflow WITH PASSWORD 'password' SUPERUSER;
        CREATE DATABASE marketflow OWNER marketflow;
        GRANT ALL PRIVILEGES ON DATABASE marketflow TO marketflow;
    " 2>/dev/null || echo "User might already exist or postgres user not available"

    # Run schema creation
    echo "📋 Creating schema..."
    sudo docker exec -i marketflow-postgres-1 psql -U marketflow -d marketflow << 'EOF'
CREATE TABLE IF NOT EXISTS market_data (
    id SERIAL PRIMARY KEY,
    pair_name VARCHAR(20) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    average_price DECIMAL(20, 8) NOT NULL,
    min_price DECIMAL(20, 8) NOT NULL,
    max_price DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_market_data_pair_exchange ON market_data(pair_name, exchange);
CREATE INDEX IF NOT EXISTS idx_market_data_timestamp ON market_data(timestamp);
CREATE INDEX IF NOT EXISTS idx_market_data_created_at ON market_data(created_at);
EOF

    echo "🔍 Testing connection again..."
    sudo docker exec marketflow-postgres-1 psql -U marketflow -d marketflow -c "SELECT 'Manual setup successful!' as status;"
fi

echo ""
echo "📊 Container status:"
sudo docker-compose ps

echo ""
echo "🎉 Setup complete! Try running:"
echo "  ./marketflow --port 8080"