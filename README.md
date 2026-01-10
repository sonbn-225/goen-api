# goen-api

Personal finance REST API for the Goen ecosystem.

## Architecture

This project uses **Clean Architecture** with **feature-based modules**:

```
goen-api/
├── cmd/api/              # Application entry point
├── internal/
│   ├── app/              # Application wiring & dependency injection
│   ├── apperrors/        # Unified error handling
│   ├── config/           # Environment configuration
│   ├── domain/           # Domain entities & repository interfaces
│   ├── httpapi/          # HTTP middleware (auth, cors, logging)
│   ├── modules/          # Feature modules (vertical slices)
│   │   ├── account/      # Bank accounts & wallets
│   │   ├── auth/         # Authentication (register, login, JWT)
│   │   ├── budget/       # Budget management
│   │   ├── category/     # Transaction categories
│   │   ├── debt/         # Debt tracking
│   │   ├── diagnostics/  # Health checks (healthz, readyz, ping)
│   │   ├── investment/   # Investment portfolio
│   │   ├── marketdata/   # Stock/crypto market data
│   │   ├── rotating_savings/ # Rotating savings (hui, arisan)
│   │   ├── savings/      # Savings goals
│   │   ├── tag/          # Transaction tags
│   │   └── transaction/  # Income/expense transactions
│   ├── response/         # HTTP response utilities
│   └── storage/          # PostgreSQL repositories
├── migrations/           # Database migrations (goose)
└── docs/                 # Swagger documentation
```

### Package Responsibilities

| Package | Purpose |
|---------|---------|
| **cmd/api** | `main.go` - bootstrap config, database, start HTTP server |
| **internal/app** | Wire dependencies, create modules, setup router |
| **internal/apperrors** | Sentinel errors (`ErrNotFound`), error kinds, HTTP status mapping |
| **internal/config** | Load & validate env vars (DB URLs, JWT secrets, ports) |
| **internal/domain** | Pure domain entities & repository interfaces (no dependencies) |
| **internal/httpapi** | Middleware: `AuthMiddleware`, `CORSMiddleware`, `RequestLogger` |
| **internal/modules** | Feature modules - each contains `module.go`, `service.go`, `handler.go` |
| **internal/response** | `WriteJSON()`, `WriteError()` - HTTP response helpers |
| **internal/storage** | Repository implementations using pgx/v5 for PostgreSQL |

### Module Structure

Each module follows the same pattern:

```
modules/account/
├── module.go      # New() constructor, RegisterRoutes()
├── service.go     # Business logic, validation, orchestration
└── handler.go     # HTTP handlers (parse request → call service → write response)
```

### Data Flow

```
Request → Middleware → Handler → Service → Storage → PostgreSQL
                          ↓
                       Domain (entities)
                          ↓
Response ← Middleware ← Handler ← Service ← Storage
```

### File Naming Conventions

| Layer | Convention | Example |
|-------|------------|---------|
| **domain/** | Singular | `user.go`, `account.go`, `transaction.go` |
| **storage/** | Plural | `users.go`, `accounts.go`, `transactions.go` |
| **modules/** | Folder per feature | `account/`, `auth/`, `transaction/` |

---

## What’s included

- Go HTTP server with routes:
  - `GET /api/v1/healthz`
  - `GET /api/v1/readyz` (checks Postgres + Redis if configured)
  - `GET /api/v1/ping` (for goen-web browser test)
  - `GET /api/v1/connectivity` (probes Postgres + Redis and returns details)
- Dockerfile (builds a static-ish Linux binary)
- `docker-compose.dev.yml` (Traefik HTTP only)
- `docker-compose.prod.yml` (Traefik HTTPS via `le` resolver)

## Run (dev)

Prereqs:
- Traefik dev stack running on the same Docker network (`proxy_network` by default).
- Postgres/Redis stacks running if you want `/readyz` to pass.

Steps:
- Copy env: `cp .env.example .env` (on Windows: duplicate file in Explorer)
- Start: `docker compose --env-file .env -f docker-compose.dev.yml up -d --build`

Then:

- Pick a domain for Traefik routing by setting `API_DOMAIN` in `.env`.
  - Example (dev): `API_DOMAIN=api.your-dev-domain.localhost`

- Endpoints (replace `<API_DOMAIN>` with your configured value):
  - `http://<API_DOMAIN>/api/v1/healthz`
  - `http://<API_DOMAIN>/api/v1/readyz`
  - `http://<API_DOMAIN>/api/v1/ping`
  - `http://<API_DOMAIN>/api/v1/connectivity`
  - Swagger UI: `http://<API_DOMAIN>/swagger/`

PowerShell note (Windows): if `<API_DOMAIN>` does not resolve in PowerShell, call through `http://localhost` and set the `Host` header:

```powershell
$apiDomain = "api.your-dev-domain.localhost" # set to your API_DOMAIN
Invoke-RestMethod -Uri "http://localhost/api/v1/ping" -Headers @{ Host = $apiDomain }
```

## Add a new API endpoint (Module Pattern)

This repo uses **feature-based modules**. Each feature is a self-contained module in `internal/modules/`.

### Steps to add a new module:

1. **Create module folder**
   ```
   internal/modules/yourfeature/
   ├── module.go
   ├── service.go
   └── handler.go
   ```

2. **Define the module** (`module.go`)
   ```go
   package yourfeature

   import (
       "github.com/go-chi/chi/v5"
       "github.com/user/goen-api/internal/storage"
   )

   type Module struct {
       service *Service
       handler *Handler
   }

   func New(store *storage.YourStore) *Module {
       svc := NewService(store)
       return &Module{
           service: svc,
           handler: NewHandler(svc),
       }
   }

   func (m *Module) RegisterRoutes(r chi.Router) {
       r.Route("/yourfeature", func(r chi.Router) {
           r.Get("/", m.handler.List)
           r.Post("/", m.handler.Create)
           r.Get("/{id}", m.handler.Get)
       })
   }
   ```

3. **Implement service** (`service.go`) - business logic
4. **Implement handlers** (`handler.go`) - HTTP request/response

5. **Register in app** (`internal/app/app.go`)
   ```go
   yourMod := yourfeature.New(yourStore)
   yourMod.RegisterRoutes(r)
   ```

6. **(Optional) Add Swagger docs**
   - Add Swaggo annotations above your handlers
   - Dev mode auto-regenerates Swagger docs on rebuild

7. **Rebuild/restart**
   - Dev: container auto-rebuilds on `.go` changes (Air)
   - Prod: `docker compose --env-file .env -f docker-compose.prod.yml up -d --build`

## Run (prod)

- Configure `.env` (set `API_DOMAIN`, `DATABASE_URL`, `REDIS_URL`)
- Start: `docker compose --env-file .env -f docker-compose.prod.yml up -d --build`

## Database Migrations

This project uses [goose](https://github.com/pressly/goose) for database migrations. Goose is installed in the Docker container.

### Create a new migration

```bash
docker exec goen-api goose -dir ./migrations create migration_name sql
```

This will create a new SQL file in the `migrations/` directory.

### Run migrations

To apply all available migrations:

```bash
docker exec goen-api sh -c 'export GOOSE_DRIVER=postgres && export GOOSE_DBSTRING=$DATABASE_URL && goose -dir ./migrations up'
```

### Check migration status

```bash
docker exec goen-api sh -c 'export GOOSE_DRIVER=postgres && export GOOSE_DBSTRING=$DATABASE_URL && goose -dir ./migrations status'
```
