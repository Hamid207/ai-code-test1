#!/bin/bash

echo "🔍 Redis Debug Information"
echo "=========================="
echo ""

echo "📝 Checking .env file..."
if [ -f .env ]; then
    echo "✅ .env file exists"
    echo ""
    echo "REDIS_PASSWORD value:"
    grep "^REDIS_PASSWORD=" .env || echo "❌ REDIS_PASSWORD not found in .env"
    echo ""
else
    echo "❌ .env file NOT found!"
    exit 1
fi

echo "📋 Redis container logs:"
echo "========================"
docker-compose logs redis | tail -50

echo ""
echo "🔍 Redis container inspect:"
echo "=========================="
docker-compose exec redis env | grep REDIS || echo "Cannot connect to Redis"

echo ""
echo "💡 Try manual Redis connection:"
echo "================================"
echo "docker-compose exec redis redis-cli -a 2aW8eR1tY4uI7oP0sAf8K2mP9nQ4rT7w ping"
