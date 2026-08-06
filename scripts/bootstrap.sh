#!/usr/bin/env bash
# Bootstrap dev environment for monorepo.
# Usage: bash scripts/bootstrap-dev.sh

set -euo pipefail

echo "=== Dev Bootstrap ==="
echo ""

# Check prerequisites
command -v node >/dev/null 2>&1 || { echo "ERROR: Node.js not found. Install from https://nodejs.org"; exit 1; }
command -v pnpm >/dev/null 2>&1 || { echo "ERROR: pnpm not found. Run: npm install -g pnpm"; exit 1; }
command -v go >/dev/null 2>&1 || { echo "ERROR: Go not found. Install from https://go.dev/dl/"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "ERROR: Docker not found. Install Docker Desktop."; exit 1; }

echo "Node.js: $(node --version)"
echo "pnpm: $(pnpm --version)"
echo "Go: $(go version | awk '{print $3}')"
echo "Docker: $(docker --version | awk '{print $3}' | tr -d ',')"
echo ""

# 1. Install JS dependencies
echo "[1/5] Installing JS dependencies..."
pnpm install

# 2. Install Go dependencies
echo "[2/5] Installing Go dependencies..."
cd backend
go mod download
cd ..

# 3. Install Go dev tools
echo "[3/5] Installing Go dev tools..."
go install github.com/air-verse/air@latest

# 4. Copy env file
echo "[4/5] Setting up environment..."
if [ ! -f .env ]; then
  cp .env.example .env
  echo "  Created .env from .env.example"
else
  echo "  .env already exists, skipping"
fi

if [ ! -f backend/.env ]; then
  cp backend/.env.example backend/.env
  echo "  Created backend/.env from backend/.env.example"
else
  echo "  backend/.env already exists, skipping"
fi

# 5. Start infrastructure
echo "[5/5] Starting Docker infrastructure..."
docker compose -f docker-compose.dev.yml up -d

echo ""
echo "=== Bootstrap Complete ==="
echo ""
echo "Next steps:"
echo "  1. Wait for Docker services to be healthy: docker compose -f docker-compose.dev.yml ps"
echo "  2. Init Garage bucket (first time only): bash infra/scripts/init-garage-bucket.sh"
echo "  3. Start backend: cd backend && air"
echo "  4. Start frontends: pnpm dev"
echo ""
echo "Ports:"
echo "  web:    http://localhost:3000"
echo "  backend:     http://localhost:8080"
echo "  PostgreSQL:  localhost:5432 (go-db) / localhost:5433 (odoo-db)"
echo "  Redis:       localhost:6379"