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
cd frontend
pnpm install
cd ..

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
if [ ! -f backend/.env ]; then
  cp backend/.env.example backend/.env
  echo "  Created backend/.env from backend/.env.example"
else
  echo "  backend/.env already exists, skipping"
fi

if [ ! -f frontend/.env ]; then
  cp frontend/.env.example frontend/.env
  echo "  Created frontend/.env from frontend/.env.example"
else
  echo "  frontend/.env already exists, skipping"
fi

# 5. Start infrastructure
echo "[5/5] Starting Docker infrastructure..."
echo ""
echo "⚠️  Please review the env files before starting infrastructure."
echo "    Default values from .env.example are placeholders and may not"
echo "    be safe or correct for your setup."
echo ""
echo "    Files to review:"
echo "      - backend/.env"
echo "      - frontend/.env"
echo ""

START_INFRA="skip"
if [ -t 0 ]; then
  # Interactive terminal — ask the developer
  read -r -p "Start Docker infrastructure now? [y/N] " reply
  case "$reply" in
    [yY]|[yY][eE][sS]) START_INFRA="yes" ;;
    *)                  START_INFRA="no"  ;;
  esac
else
  # Non-interactive (CI, piped input) — never start infra implicitly
  echo "  Non-interactive shell detected — skipping Docker startup."
  echo "  Run 'docker compose -f docker-compose.yaml up -d' manually when ready."
fi

if [ "$START_INFRA" = "yes" ]; then
  docker compose -f docker-compose.yaml up -d
fi

echo ""
echo "=== Bootstrap Complete ==="
echo ""
echo "Next steps:"
echo "  1. Start Docker infrastructure (if not done): docker compose -f docker-compose.yaml up -d"
echo "  2. Wait for services to be healthy: docker compose -f docker-compose.yaml ps"
echo "  3. Init PostgreSQL database (first time only): bash infra/scripts/init-postgresql.sh"
echo "  4. Start backend: cd backend && air"
echo "  5. Start frontend: cd frontend && pnpm dev"
echo ""
echo "Ports:"
echo "  web:        http://localhost:3000"
echo "  backend:    http://localhost:8080"
echo "  PostgreSQL: localhost:5432 (postgresql) / localhost:5433 (odoo-db)"
echo "  Redis:      localhost:6379"