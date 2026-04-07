# AGENTS.md — mengri-flow

## Project Overview

Single-binary full-stack app: Go 1.23 backend + Vue 3 frontend. Vite compiles the
frontend into `web/dist/`, which is embedded into the Go binary via `go:embed`.
The binary serves API at `/api/v1/*` and the SPA at `/`. No external web server needed.

## Build & Run

**Go version**: 1.23. Use `GOTOOLCHAIN=local` for all `go` commands (Gin v1.10.0
requires Go 1.23; newer Gin needs 1.25+).

```bash
make build              # full build: frontend + backend → single binary
make build-web          # frontend only (cd web && npm install && npm run build)
make build-server       # backend only (requires web/dist/ to exist)
make run                # go run (backend only)
make dev                # backend + vite dev server (frontend :3000, API :8080)
make docker-build       # docker build
make docker-run         # docker run with .env
```

## Testing

```bash
GOTOOLCHAIN=local go test ./... -v -cover                           # all tests
GOTOOLCHAIN=local go test ./pkg/autowire/... -v                     # single package
GOTOOLCHAIN=local go test ./pkg/autowire/... -run TestAutoBasic -v  # single test
```

**Conventions:**
- Table-driven tests with `t.Parallel()` where possible.
- Mock external interfaces (repositories, external APIs); never mock domain entities.
- Test files: `*_test.go` in the same package.
- File length: 300 lines soft / 900 medium / 1500 hard.

## Linting

```bash
golangci-lint run ./...       # Go (no .golangci.yml committed yet)
cd web && npm run lint        # Frontend (eslint not installed yet)
```

## Architecture — DDD Clean Architecture

```
cmd/server/main.go             # Entrypoint, DI wiring via pkg/autowire
internal/
  domain/                      # ZERO external deps (no gorm, gin, etc.)
    entity/                    # Aggregate roots with business methods
    valueobject/               # Immutable, self-validating value objects
    repository/                # Repository interfaces (defined by consumer)
    errors/                    # Sentinel domain errors (stdlib only)
  app/
    service/                   # Application services (use case orchestration)
    dto/                       # Request/Response DTOs with binding tags
  infra/
    config/                    # YAML config loader
    persistence/mysql/         # GORM repository implementations
    plugin/                    # Plugin framework (types, registry)
  ports/http/
    handler/                   # Gin HTTP handlers with Swagger annotations
    middleware/                # Logger, Recovery, CORS
    router/                    # Route registration + SPA serving
pkg/
  autowire/                    # Custom DI framework (tag-based injection)
  response/                    # Unified { code, data, msg } JSON helpers
  logger/                      # Structured logging (slog)
web/                           # Vue 3 + Vite + TypeScript + Pinia + Element Plus
  embed.go                     # go:embed all:dist (placeholder index.html is git-tracked)
plugins/                       # Plugin directory (build-time integration)
  resource/                    # Resource plugins (HTTP, gRPC, etc.)
  trigger/                     # Trigger plugins (RESTful, Timer, MQ)
  plugins.yaml                 # Build tags configuration
```

**Dependency direction:** Domain <- App <- Infra/Ports. Never import inward.

## Dependency Injection (autowire)

Custom DI framework in `pkg/autowire/`. Registration uses `init()`:
```go
func init() {
    autowire.Auto(func() UserService { return &UserServiceImpl{} })
}
```
Fields injected via struct tags: `autowired:""`. Side-effect imports in
`internal/ports/http/router/` trigger registrations.

## Code Style — Go

### Naming
- **Structs/Interfaces**: PascalCase (`UserService`, `UserRepository`)
- **Methods**: PascalCase exported, camelCase unexported
- **Files**: snake_case (`user_service.go`), interface files use `_iface.go` suffix
- **Packages**: lowercase, single word preferred

### Imports — Three groups separated by blank lines
```go
import (
    "context"
    "fmt"

    "github.com/gin-gonic/gin"

    "mengri-flow/internal/domain/entity"
)
```

### Error Handling
- Always check errors explicitly.
- Wrap with context: `fmt.Errorf("create user: %w", err)`
- Domain errors: sentinel errors in `internal/domain/errors/`.
- Handlers map domain errors to HTTP status via `handleDomainError()`.

### Interface Compliance
Use compile-time checks: `var _ UserService = (*UserServiceImpl)(nil)`

### Complexity & Length Limits
| Metric              | Acceptable | Must Refactor |
|---------------------|------------|---------------|
| Cyclomatic          | <= 10      | > 15          |
| Cognitive           | <= 30      | > 50          |
| Function lines      | <= 60      | —             |
| Function statements | <= 40      | —             |
| Line length         | 120 chars  | (soft limit)  |
| File (production)   | 200 soft   | 1000 hard     |
| File (test)         | 300 soft   | 1500 hard     |

### Key Principles
- **No anemic models**: business logic in Entity methods, not services.
- **Dependency inversion**: interfaces in Domain, implementations in Infra.
- **No global state**: constructor functions + DI.
- **Context propagation**: pass `context.Context` through all layers.
- **Functional Options Pattern** for public APIs with > 2 config fields.

## Code Style — Vue 3 / TypeScript

详见 [docs/WEB_DEV.md](docs/WEB_DEV.md)

## API Response Contract

All endpoints return `{ code: number, data: T, msg: string }`.
`code === 0` means success. Frontend Axios interceptor handles non-zero globally.

## Feature Generation Checklist

When creating a new feature, generate all layers:
1. `internal/domain/entity/<name>.go` — aggregate root with methods
2. `internal/domain/repository/<name>_repository.go` — interface
3. `internal/app/dto/<name>_dto.go` — request/response DTOs
4. `internal/app/service/<name>_service.go` + `_iface.go` — use case + autowire init
5. `internal/infra/persistence/mysql/<name>_repository.go` — GORM impl + autowire init
6. `internal/ports/http/handler/<name>_handler.go` + `_iface.go` — Gin handlers + autowire
7. `web/src/types/<name>.ts`, `api/<name>.ts`, `composables/use<Name>.ts`,
   `stores/<name>.ts`, `views/<name>/index.vue`
