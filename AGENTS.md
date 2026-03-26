# AGENTS.md — mengri-flow

## Project Overview

mengri-flow is a single-binary full-stack application (Go backend + Vue 3 frontend).
The frontend is compiled by Vite into `web/dist/`, then embedded into the Go binary via
`go:embed`. The final binary serves both API (`/api/v1/*`) and SPA (`/`) with no external
web server required. Designed for private deployment (私有化部署).

## Build & Run

**Go version**: 1.23 (via gvm). Must use `GOTOOLCHAIN=local` for all `go` commands
because Gin v1.10.0 is compatible with Go 1.23, but newer Gin versions require Go 1.25+.

```bash
# Full build (frontend + backend → single binary)
make build                    # or: make build-web && make build-server

# Frontend only
cd web && npm install && npm run build

# Backend only (requires web/dist/ to exist)
GOTOOLCHAIN=local go build -v -ldflags "-s -w" -o bin/mengri-flow ./cmd/server

# Run (development)
make run                      # backend only
make dev                      # backend + vite dev server (frontend at :3000, API at :8080)

# Docker
make docker-build
make docker-run
```

## Testing

```bash
# All Go tests
GOTOOLCHAIN=local go test ./... -v -cover

# Single package
GOTOOLCHAIN=local go test ./internal/domain/entity/... -v

# Single test function
GOTOOLCHAIN=local go test ./internal/domain/entity/... -run TestUserActivate -v

# Frontend (no test runner configured yet)
cd web && npm run lint
```

**Testing conventions:**
- Table-driven tests with `t.Parallel()` where possible.
- Mock external interfaces; never mock domain entities.
- Test files: `*_test.go` in the same package.
- Test file length limit: 300 lines soft / 900 medium / 1500 hard.

## Linting

```bash
# Go (requires golangci-lint)
golangci-lint run ./...

# Frontend
cd web && npm run lint
```

**Note:** `.golangci.yml`, `.eslintrc`, `.prettierrc` do not yet exist in the repo.
The DDD skill doc defines the intended golangci-lint settings (see below).

## Architecture — Go DDD Clean Architecture

```
cmd/server/main.go          # Entrypoint, dependency injection wiring
internal/
  domain/                   # ZERO external dependencies (no gorm, gin, etc.)
    entity/                 # Aggregate roots with business methods
    valueobject/            # Immutable, self-validating value objects
    repository/             # Repository interfaces (defined by consumer)
    errors/                 # Sentinel domain errors (stdlib only)
  app/
    service/                # Application services (use case orchestration)
    dto/                    # Request/Response DTOs with binding tags
  infra/
    config/                 # YAML config loader with env var expansion
    persistence/mysql/      # GORM repository implementations + models
  ports/http/
    handler/                # Gin HTTP handlers with Swagger annotations
    middleware/             # Logger, Recovery, CORS
    router/                 # Route registration + SPA serving
pkg/
  response/                 # Unified { code, data, msg } JSON helpers
  logger/                   # Structured logging (slog)
web/                        # Vue 3 frontend (Vite + TypeScript + Pinia)
  embed.go                  # go:embed all:dist
```

**Dependency direction:** Domain ← App ← Infra/Ports. Never import inward.

## Code Style — Go

### Naming
- **Structs/Interfaces**: PascalCase (`UserService`, `UserRepository`)
- **Methods/Functions**: PascalCase for exported, camelCase for unexported
- **Files**: snake_case (`user_service.go`, `user_repository.go`)
- **Packages**: lowercase, single word preferred

### Error Handling
- Always check and handle errors explicitly.
- Wrap errors with context: `fmt.Errorf("create user: %w", err)`
- Domain errors are sentinel errors in `internal/domain/errors/`.
- Handlers map domain errors to HTTP status codes.

### Imports
Standard library first, then external packages, then internal packages.
Separate each group with a blank line:
```go
import (
    "context"
    "fmt"

    "github.com/gin-gonic/gin"

    "mengri-flow/internal/domain/entity"
)
```

### Complexity Limits
- Cyclomatic complexity: <= 10 (acceptable), > 15 (must refactor)
- Cognitive complexity: <= 30 (acceptable), > 50 (must refactor)
- Function length: <= 60 lines, <= 40 statements
- Line length: 120 chars soft limit

### File Length Limits
- Production code: 200 lines soft / 600 medium / 1000 hard
- Test code: 300 / 900 / 1500 (1.5x for table-driven tests)
- Generated code is exempt

### Key Principles
- **No anemic models**: Business logic lives in Entity methods, not services.
- **Dependency inversion**: Interfaces in Domain, implementations in Infra.
- **No global state**: Use constructor functions and dependency injection.
- **Context propagation**: Pass `context.Context` through all layers.
- Use **Functional Options Pattern** for public APIs with > 2 config fields.

## Code Style — Vue 3 / TypeScript

### Component Structure
- Always use `<script setup lang="ts">`.
- Never use `any` — define TypeScript interfaces for all API types.
- Complex logic goes in `src/composables/`, not inline in components.
- State management via Pinia stores in `src/stores/`.

### Directory Layout
```
web/src/
  api/              # API call functions (one file per domain)
  composables/      # Reusable logic hooks (useUser, useAuth, etc.)
  components/common/# Shared base components
  stores/           # Pinia stores
  types/            # TypeScript interfaces/types
  utils/            # Utilities (Axios instance, helpers)
  views/            # Page components
    feature/
      index.vue
      components/   # Page-private components
  router/           # Vue Router config
```

### Naming
- Variables/functions: camelCase
- Components: PascalCase (files and usage)
- CSS classes / asset files: kebab-case
- Type/Interface names: PascalCase

### API Response Format
All backend responses follow `{ code: number, data: T, msg: string }`.
- `code === 0` means success.
- Frontend Axios interceptor handles non-zero codes globally.

## Embedding & Deployment

- `web/embed.go` uses `//go:embed all:dist` to embed the entire `web/dist/` directory.
- `web/dist/` must contain at least one file for the embed directive to compile.
- A placeholder `web/dist/index.html` is git-tracked (`.gitignore` excludes `web/dist/*`
  but includes `!web/dist/index.html`).
- The Go router serves static assets from the embedded FS at `/` and falls back to
  `index.html` for all non-API, non-static routes (SPA client-side routing).

## Feature Generation

When creating a new feature, generate all layers simultaneously:

1. **Domain**: `internal/domain/entity/<name>.go` (aggregate root with methods)
2. **Domain**: `internal/domain/repository/<name>_repository.go` (interface)
3. **App**: `internal/app/dto/<name>_dto.go` (request/response DTOs)
4. **App**: `internal/app/service/<name>_service.go` (use case orchestration)
5. **Infra**: `internal/infra/persistence/mysql/<name>_repository.go` (GORM impl)
6. **Ports**: `internal/ports/http/handler/<name>_handler.go` (Gin handlers + Swagger)
7. **Frontend**: `web/src/types/<name>.ts`, `web/src/api/<name>.ts`,
   `web/src/composables/use<Name>.ts`, `web/src/stores/<name>.ts`,
   `web/src/views/<name>/index.vue`

## Known Issues / TODOs

- `internal/app/service/user_service.go:33` — password stored in plaintext; must use bcrypt.
- No test files exist yet (`*_test.go`).
- No linting config files (`.golangci.yml`, `.eslintrc`, `.prettierrc`, `.editorconfig`).
- Vite build warns about chunk size > 500 kB; consider code-splitting with dynamic imports
  or `manualChunks` in `vite.config.ts`.
