#!/bin/bash
set -e

# Deploy script for randomcitypicker
# Run this from the project root on the droplet

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "📥 Pulling latest changes..."
git pull origin main

echo "🐳 Rebuilding and restarting production containers..."
docker compose -f docker-compose.prod.yml up -d --build

echo "🧹 Cleaning up old Docker images..."
docker system prune -f

echo "✅ Deploy complete!"
