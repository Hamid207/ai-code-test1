#!/bin/bash

# ===========================================
# Secure Environment Setup Script
# ===========================================
# This script creates a .env file with strong passwords
# NEVER commit the generated .env file!

set -e

echo "🔐 iOS Backend - Secure Environment Setup"
echo "=========================================="
echo ""

# Check if .env already exists
if [ -f .env ]; then
    echo "⚠️  WARNING: .env file already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Setup cancelled."
        exit 1
    fi
fi

echo "📝 Generating .env file with strong passwords..."
echo ""

# Generate strong passwords
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
REDIS_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
JWT_SECRET=$(openssl rand -base64 48 | tr -d "=+/" | cut -c1-48)

# Create .env file
cat > .env << EOF
# ===========================================
# iOS Backend - Environment Configuration
# ===========================================
# AUTO-GENERATED: $(date)
# DO NOT COMMIT THIS FILE!

# ===========================================
# Server Configuration
# ===========================================
SERVER_PORT=8080
APP_PORT=8080
APP_ENV=development

# ===========================================
# Database Configuration (PostgreSQL)
# ===========================================
DB_USER=postgres
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=appdb
DB_PORT=5432
DB_SSL_MODE=disable

# Database connection pool settings
DB_MAX_CONNS=25
DB_MIN_CONNS=5
DATABASE_MAX_CONNECTIONS=25
DATABASE_MIN_CONNECTIONS=5

# For Docker Compose (PostgreSQL container)
POSTGRES_USER=postgres
POSTGRES_PASSWORD=${DB_PASSWORD}
POSTGRES_DB=appdb

# ===========================================
# Redis Configuration
# ===========================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0
REDIS_PASSWORD=${REDIS_PASSWORD}
REDIS_URL=redis:6379

# Redis connection pool settings
REDIS_MAX_CONNS=10
REDIS_MIN_IDLE_CONNS=2

# ===========================================
# Apple OAuth Configuration
# ===========================================
APPLE_TEAM_ID=REPLACE_WITH_YOUR_APPLE_TEAM_ID
APPLE_CLIENT_ID=REPLACE_WITH_YOUR_APPLE_CLIENT_ID
APPLE_KEY_ID=REPLACE_WITH_YOUR_APPLE_KEY_ID
APPLE_PRIVATE_KEY_PATH=./configs/AuthKey_XXXXX.p8

# ===========================================
# Google OAuth Configuration
# ===========================================
GOOGLE_CLIENT_ID=REPLACE_WITH_YOUR_GOOGLE_CLIENT_ID
GOOGLE_CLIENT_SECRET=REPLACE_WITH_YOUR_GOOGLE_CLIENT_SECRET

# ===========================================
# JWT Configuration
# ===========================================
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRATION=24h

# ===========================================
# OpenAI Configuration
# ===========================================
OPENAI_API_KEY=REPLACE_WITH_YOUR_OPENAI_API_KEY

# ===========================================
# CORS Configuration
# ===========================================
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
EOF

echo "✅ .env file created successfully!"
echo ""
echo "🔑 Generated Credentials:"
echo "========================"
echo "DB_PASSWORD:     ${DB_PASSWORD}"
echo "REDIS_PASSWORD:  ${REDIS_PASSWORD}"
echo "JWT_SECRET:      ${JWT_SECRET:0:20}... (truncated)"
echo ""
echo "⚠️  IMPORTANT: You still need to manually add:"
echo "   1. APPLE_TEAM_ID, APPLE_CLIENT_ID, APPLE_KEY_ID"
echo "   2. GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET"
echo "   3. OPENAI_API_KEY (generate new one if exposed)"
echo ""
echo "📝 Edit .env file:"
echo "   nano .env"
echo ""
echo "🔒 Security Checklist:"
echo "   ✅ Strong passwords generated automatically"
echo "   ✅ .env file is in .gitignore (DO NOT COMMIT!)"
echo "   ⚠️  Add your real OAuth credentials"
echo "   ⚠️  Add your OpenAI API key (generate new if old one exposed)"
echo ""
echo "🚀 Start Docker containers:"
echo "   docker-compose up -d"
echo ""
