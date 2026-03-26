# mengri-flow

Single-binary full-stack application built with Go (DDD Clean Architecture) and Vue 3.
The frontend is compiled by Vite and embedded into the Go binary via `go:embed` — no
Nginx or separate frontend deployment required. Designed for private deployment (私有化部署).

## Tech Stack

**Backend:** Go 1.23 | Gin | GORM | MySQL | slog

**Frontend:** Vue 3 (Composition API) | TypeScript | Vite | Pinia | Tailwind CSS | Element Plus

## Architecture

```
cmd/server/main.go            Entrypoint + dependency injection
internal/
  domain/                     Pure business logic (zero external deps)
    entity/                   Aggregate roots
    valueobject/              Immutable value objects
    repository/               Repository interfaces
    errors/                   Sentinel domain errors
  app/
    service/                  Use case orchestration
    dto/                      Request/Response DTOs
  infra/
    config/                   YAML config with env var expansion
    persistence/mysql/        GORM implementations
  ports/http/
    handler/                  Gin HTTP handlers
    middleware/               Logger, Recovery, CORS
    router/                   Route registration + SPA serving
pkg/
  response/                   Unified { code, data, msg } JSON helpers
  logger/                     Structured logging (slog)
web/                          Vue 3 frontend
  embed.go                    go:embed all:dist
```

Dependency direction: **Domain <- App <- Infra/Ports**. Domain layer has zero external imports.

## Prerequisites

- Go 1.23 (via [gvm](https://github.com/moovweb/gvm) or direct install)
- Node.js 22+
- MySQL 8.0+
- (Optional) Docker

## Quick Start

### 1. Clone and install dependencies

```bash
git clone <repo-url> && cd mengri-flow
make deps
```

### 2. Configure

Copy the example env file and edit as needed:

```bash
cp .env.example .env
```

Configuration is loaded from `config.yaml` with environment variable expansion:

| Variable        | Default           | Description          |
|-----------------|-------------------|----------------------|
| `DB_HOST`       | `127.0.0.1`       | MySQL host           |
| `DB_PORT`       | `3306`            | MySQL port           |
| `DB_USER`       | `root`            | MySQL user           |
| `DB_PASSWORD`   | `123456`          | MySQL password       |
| `DB_NAME`       | `mengri_flow`     | Database name        |
| `REDIS_ADDR`    | `127.0.0.1:6379`  | Redis address        |
| `REDIS_PASSWORD` | (empty)          | Redis password       |
| `CONFIG_PATH`   | `config.yaml`     | Config file path     |

### 3. Database migration

```bash
# Create database
mysql -u root -p -e "CREATE DATABASE mengri_flow CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Run migrations (requires golang-migrate)
make migrate-up

# Or let GORM auto-migrate in debug mode (default)
```

### 4. Build and run

```bash
# Full build: frontend + backend → single binary (~12MB)
make build
./bin/mengri-flow

# Development mode (hot reload)
make dev
# Backend: http://localhost:8080
# Frontend: http://localhost:3000 (proxied to backend)
```

## Available Commands

```
make build         Full build (frontend + backend → single binary)
make build-web     Build frontend only → web/dist/
make build-server  Build backend only (requires web/dist/)
make run           Run backend
make dev           Development mode (backend + Vite dev server)
make test          Run all Go tests
make lint          Run golangci-lint
make clean         Remove build artifacts
make deps          Install all dependencies (Go + npm)
make sqlc          Generate SQLC code
make migrate-up    Run database migrations
make migrate-down  Rollback database migrations
make swagger       Generate Swagger docs
make docker-build  Build Docker image
make docker-run    Run Docker container
```

**Note:** All `go` commands require `GOTOOLCHAIN=local` when run manually (the Makefile
handles this automatically for most targets).

## Docker

```bash
# Build (3-stage: Node → Go → Alpine)
make docker-build

# Run
make docker-run
# or
docker run -p 8080:8080 --env-file .env mengri-flow
```

The Docker image uses a multi-stage build:
1. **Node 22 Alpine** — builds the Vue 3 frontend
2. **Go 1.23 Alpine** — compiles the backend with embedded frontend
3. **Alpine 3.19** — minimal runtime (~12MB binary + config only)

## API

All API endpoints are under `/api/v1/`. Responses use a unified format:

```json
{
  "code": 0,
  "data": {},
  "msg": "success"
}
```

`code === 0` indicates success. Non-zero codes indicate errors.

### User endpoints

| Method | Path              | Description      |
|--------|-------------------|------------------|
| POST   | `/api/v1/users`   | Create user      |
| GET    | `/api/v1/users`   | List users       |
| GET    | `/api/v1/users/:id` | Get user by ID |
| PUT    | `/api/v1/users/:id` | Update user    |
| DELETE | `/api/v1/users/:id` | Delete user    |

## Project Structure — Frontend

```
web/src/
  api/                API call functions (one file per domain)
  composables/        Reusable logic hooks (useUser, useAuth, ...)
  components/common/  Shared base components
  stores/             Pinia stores
  types/              TypeScript interfaces
  utils/              Axios instance, helpers
  views/              Page components
  router/             Vue Router config
```

## License

Private / Proprietary
