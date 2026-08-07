# Quản Lý Chi Tiêu

Dự án cá nhân để quản lý chi tiêu, tích hợp chatbot MCP để thay thế cho các thao tác giao diện thủ công — người dùng có thể h�i hoặc nhờ chatbot thực hiện các thao tác quản lý chi tiêu thay vì click qua nhiều màn hình.

## Công nghệ sử dụng

### Backend
- **Go** 1.25+ — ngôn ngữ chính
- **Echo** — HTTP framework
- **GORM** — ORM cho PostgreSQL
- **PostgreSQL 16** — cơ sở dữ liệu chính (chạy qua Docker)
- **Redis 7** — cache / session (chạy qua Docker)
- **Viper** — đọc cấu hình từ file `.env`
- **slog** — logging có cấu trúc

### Frontend
- **React 19** + **TypeScript**
- **Vite** — dev server / build tool
- **pnpm** — package manager
- **ESLint** — linting

### Infrastructure
- **Docker Compose** — khởi động PostgreSQL + Redis

## Cấu trúc thư mục

```
.
├── backend/                 # Go API server
│   ├── cmd/
│   │   └── server/
│   │       └── main.go      # Entry point
│   └── internal/
│       ├── config/          # Đọc & validate config từ .env
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

## Yêu cầu môi trường

- Node.js 20+
- pnpm 11+
- Go 1.25+
- Docker + Docker Compose
- Make

## Hư�ng dẫn Setup

### 1. Khởi tạo dự án

Chạy lệnh sau để cài dependencies, copy file env mẫu, và chuẩn bị Docker infrastructure:

```bash
make init
```

Script sẽ:
- Cài JS dependencies cho `frontend/`
- Tải Go modules cho `backend/`
- Cài `air` (live reload cho Go)
- Copy `backend/.env.example` → `backend/.env`
- Copy `frontend/.env.example` → `frontend/.env`

Sau bước này, **mở `backend/.env` và sửa các giá trị mặc định** (đặc biệt `POSTGRES_PASSWORD`) trước khi khởi động infrastructure. Script sẽ hỏi bạn có muốn khởi động Docker ngay hay không.

### 2. Khởi động infrastructure

Nếu đã chọn `N` ở bước trước, hoặc muốn khởi động lại:

```bash
docker compose -f docker-compose.yaml up -d
```

Kiểm tra trạng thái:

```bash
docker compose -f docker-compose.yaml ps
```

## Chạy dự án ở local

### Backend

```bash
make run
```

Server sẽ chạy ở `http://localhost:8080`.

Kiểm tra health:

```bash
curl http://localhost:8080/health
# → {"status":"healthy"}
```

### Frontend

```bash
cd frontend
pnpm dev
```

Frontend sẽ chạy ở `http://localhost:3000`.

### Ports

| Service    | Port  |
|------------|-------|
| Backend    | 8080  |
| Frontend   | 3000  |
| PostgreSQL | 5432  |
| Redis      | 6379  |
