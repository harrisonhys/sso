#!/bin/bash

# SSO System Production Deployment Script

echo "🚀 Starting Production Deployment..."

# 1. Check requirements
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# 2. Build and Start Services
echo "📦 Building and starting services..."
cd deploy
docker-compose up --build -d

# 3. Status Check
echo "⏳ Waiting for services to initialize..."
sleep 10
docker-compose ps

echo ""
echo "✅ Deployment Complete!"
echo "Access points:"
echo "- Login UI: http://localhost"
echo "- Admin Panel: http://localhost/admin-panel/"
echo "- API: http://localhost/api"
