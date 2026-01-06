# goen-api

Minimal scaffold for Goen REST API (`/api/v1`) based on `goen-docs`.

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

## Add a new API endpoint

This repo uses a simple pattern: handlers in `internal/handlers`, routes in `internal/httpapi/router.go` under `/api/v1`.

Steps:

1. Add a handler
  - Create a new file under `internal/handlers/` (e.g. `example.go`).
  - Follow the existing handler style (accept `handlers.Deps`, return `http.HandlerFunc`).

2. Register the route
  - Edit `internal/httpapi/router.go` and add a new route under `r.Route("/api/v1", ...)`.

3. If the endpoint needs config or dependencies
  - Config/env: update `internal/config/config.go`, `.env.example`, and compose env (dev/prod) as needed.
  - New Go deps: run `go mod tidy` (or rebuild the Docker image; it runs `go mod download` during build).

4. (Optional) Add Swagger docs for the endpoint
  - Add Swaggo annotations above your handler.
  - Dev mode auto-regenerates Swagger docs on rebuild (Air runs `swag init` before compiling).
  - Manual regenerate (if you are not using the dev container):

    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    swag init -g cmd/api/main.go -o docs
    ```

5. Rebuild/restart
   - Dev (hot reload): container will auto-rebuild on `.go` changes (Air). If you change deps or env, restart:
     - `docker compose --env-file .env -f docker-compose.dev.yml up -d --build`
   - Prod: rebuild/restart:
     - `docker compose --env-file .env -f docker-compose.prod.yml up -d --build`

## Run (prod)

- Configure `.env` (set `API_DOMAIN`, `DATABASE_URL`, `REDIS_URL`)
- Start: `docker compose --env-file .env -f docker-compose.prod.yml up -d --build`
