# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Personal expense manager. The intended design is an MCP chatbot as the primary interface, replacing manual UI clicks for expense operations. The frontend and `mcp-server/` directory are scaffolded only — most real code lives in `backend/`.

Monorepo: `backend/` (Go API), `frontend/` (React + Vite), `mcp-server/` (placeholder, not implemented), `scripts/bootstrap.sh` (dev setup), `docker-compose.yaml` (Postgres + Redis).

## Common commands

```bash
make init                              # one-shot setup: install deps, copy env templates
docker compose -f docker-compose.yaml up -d   # start Postgres + Redis
docker compose -f docker-compose.yaml ps      # check service health

# Backend (run from repo root — Makefile cd's into backend/)
make run                               # go run ./cmd/server/main.go on :8080
cd backend && go build ./...           # compile check
cd backend && go test ./...            # tests (none yet)

# Frontend
cd frontend && pnpm dev                # Vite dev server on :3000
cd frontend && pnpm build              # tsc + vite build
cd frontend && pnpm lint               # eslint
```

Verify backend health:
```bash
curl http://localhost:8080/health      # → {"status":"healthy"}
```

## Backend architecture

`backend/cmd/server/main.go` wires everything together: load config → init logger → connect Postgres → build Echo server → start. It uses `signal.NotifyContext` for graceful shutdown on SIGINT/SIGTERM.

Layered by concern under `internal/`:

- **`config/`** — Viper-backed. Reads `backend/.env` (relative path, so `go run` must happen with `backend/` as cwd; `make run` handles this). `viper.ReadInConfig()` errors are intentionally swallowed — missing `.env` falls back to `SetDefault` values.
- **`logger/`** — slog wrapper. `development` env → Info level; anything else → Debug + source location, both as JSON.
- **`platform/db/`** — `NewGormConfig` returns `*gorm.DB` connecting to Postgres on `localhost` with `sslmode=disable`. Sets pool sizes (10 open, 5 idle).
- **`server/`** — Echo v5 HTTP server. `NewServer` constructs Echo and calls `SetupRoutes`. `Start` uses `echo.StartConfig` with `GracefulTimeout: 10s`. Only one route currently: `GET /health`.
- **`pkg/`** — Response helpers. Use `JSONOK/JSONList/JSONCreated/JSONError` from `pkg/api_response.go` for every handler — they wrap responses in a uniform envelope with `data/error/meta/timestamp`. Don't `c.JSON(...)` directly.
- **`auth/`** — Empty placeholder directory; auth is not implemented yet.

## Key conventions

- **Postgres env var naming**: image-native `POSTGRES_USER`/`POSTGRES_PASSWORD`/`POSTGRES_DB` (so docker-compose `env_file` works without a mapping block). `POSTGRESQL_PORT` is the only `POSTGRESQL_*` name kept, used by both compose (port mapping) and Go config (DSN port).
- **Echo v5**: signature is `func(c *echo.Context) error` (note `*echo.Context`, not `echo.Context`). Use `c.JSON(status, body)` — v5 dropped `c.Blob`/`c.String` style helpers in some places.
- **`make run` must cd into `backend/`**, not `backend/cmd/server/`, because Viper's `SetConfigFile(".env")` is relative to cwd.
- **No Co-Authored-By trailer in commits** — when asked to commit, omit the trailer.
- **Never push without an explicit ask** — `git push` only when the user says "push" / "commit and push" / "push lên remote" / similar. A bare "commit" means local-only.
- **Ask when ambiguous** — if a request has multiple valid interpretations where the choice changes the outcome, ask via AskUserQuestion first instead of guessing. Don't ask about trivia you can verify by reading the repo.

## Things that are easy to miss

- `backend/.env` is gitignored; only `.env.example` is tracked. `bootstrap.sh` copies one from the other.
- The frontend's `App.tsx` is still the Vite starter template — real UI work hasn't started.
- The repo recently moved on GitHub; `git push` may print a redirect notice to `money_lover.git`. Push still works.
