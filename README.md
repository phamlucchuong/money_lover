# Expense Manager

A personal expense manager that integrates an MCP chatbot to replace manual UI interactions — users can ask or instruct the chatbot to perform expense operations instead of clicking through multiple screens.

## Tech Stack

### Backend
- **Go** 1.25+ — primary language
- **Echo** — HTTP framework
- **GORM** — ORM for PostgreSQL
- **PostgreSQL 16** — main database (runs via Docker)
- **Redis 7** — cache / session store (runs via Docker)
- **Viper** — configuration loader (reads `.env`)
- **slog** — structured logging

### Frontend
- **React 19** + **TypeScript**
- **Vite** — dev server / build tool
- **pnpm** — package manager
- **ESLint** — linting

### Infrastructure
- **Docker Compose** — spins up PostgreSQL + Redis

## Directory Structure

```
.
├── backend/                 # Go API server
│   ├── cmd/
│   │   └── server/
│   │       └── main.go      # Entry point
│   └── internal/
│       ├── config/          # Reads & validates config from .env
│       ├── logger/          # slog wrapper
│       ├── platform/
│       │   └── db/          # GORM / Postgres connection
│       └── server/          # Echo HTTP server, routes, health check
│
├── frontend/                # React + Vite SPA
│   ├── src/                 # Source code
│   ├── public/              # Static assets
│   ├── index.html           # HTML entry
│   ├── vite.config.ts       # Vite config
│   └── package.json
│
├── scripts/
│   └── bootstrap.sh         # Dev environment setup
│
├── docker-compose.yaml      # Postgres + Redis services
└── Makefile                 # init, run
```

## Prerequisites

- Node.js 20+
- pnpm 11+
- Go 1.25+
- Docker + Docker Compose
- Make

## Setup

### 1. Initialize the project

Run the following to install dependencies, copy env templates, and prepare the Docker infrastructure:

```bash
make init
```

The script will:
- Install JS dependencies in `frontend/`
- Download Go modules in `backend/`
- Install `air` (live reload for Go)
- Copy `backend/.env.example` → `backend/.env`
- Copy `frontend/.env.example` → `frontend/.env`

After this step, **open `backend/.env` and replace the default values** (especially `POSTGRES_PASSWORD`) before starting the infrastructure. The script will ask whether you want to start Docker now.

### 2. Start infrastructure

If you chose `N` above, or want to restart the services:

```bash
docker compose -f docker-compose.yaml up -d
```

Check service status:

```bash
docker compose -f docker-compose.yaml ps
```

## Running locally

### Backend

```bash
make run
```

Server runs at `http://localhost:8080`.

Verify health:

```bash
curl http://localhost:8080/health
# → {"status":"healthy"}
```

### Frontend

```bash
cd frontend
pnpm dev
```

Frontend runs at `http://localhost:3000`.

### Ports

| Service    | Port  |
|------------|-------|
| Backend    | 8080  |
| Frontend   | 3000  |
| PostgreSQL | 5432  |
| Redis      | 6379  |
