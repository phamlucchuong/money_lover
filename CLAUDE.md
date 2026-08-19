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

## Commit grouping

When asked to commit, split the work into multiple commits grouped by **concern / layer**, not by file or by hunk size. One commit = one coherent change that can stand on its own, be reverted independently, and reviewed without reading the whole stack.

Group by these axes — pick whichever applies, but never blend them in a single commit:

- **Refactor vs. feature vs. fix vs. chore** — a model rename is `refactor:`, never folded into a `feat:` that adds new behavior.
- **Layer / package** — `feat(db):` (helpers in `internal/platform/db`) is separate from `feat(user):` (feature logic), even when both ship in the same PR.
- **Migrations vs. code** — schema changes (under `backend/migrations/`) get their own commit so a rollback can drop the index without reverting application code.
- **Tests vs. production code** — a `test:` or `test(<area>):` commit lands separately when it covers behavior introduced by an earlier commit in the same branch. Avoid bundling a fix with its test in one commit unless the test is non-load-bearing (e.g. snapshot noise).
- **Wiring / plumbing** — route registration, config keys, generated mocks belong in `chore(<area>):`, not in the `feat:` that introduces the handler.

Concrete rules:

1. Stage with `git add -p` (or write a partial patch and `git apply --cached`) when a single file mixes concerns. Do not `git add .` and accept whatever hunks land together.
2. After staging, run `git diff --cached --stat` to confirm only the intended files are staged. If a file shows up that you didn't plan to include, `git restore --staged <file>` and re-stage the hunks you want.
3. Build (`cd backend && go build ./...`) and run the relevant tests after each commit — not just after the last one. A commit that breaks the tree is not acceptable.
4. Commit messages follow Conventional Commits: `<type>(<scope>): <imperative summary>`. Body explains *why*, not *what*. Don't add a `Co-Authored-By: Claude` trailer (see project memory).
5. Never push without an explicit ask — see project memory.
