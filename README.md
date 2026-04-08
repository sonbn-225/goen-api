# Goen API (Clean Architecture)

Goen API is the core financial engine for the Sunflower ecosystem, implemented using **Clean Architecture** principles in Go. It manages personal finances, investments, and rotating savings.

## Architecture

The project follows a strict separation of concerns to ensure maintainability and testability:

- **Domain Layer (`internal/domain`)**:
    - `entity`: Core business objects (User, Transaction, etc.).
    - `dto`: Request/Response data transfer objects.
    - `interfaces`: Contracts (ports) for repositories and services.
- **Service Layer (`internal/service`)**:
    - Contains business logic and orchestrates domain entities.
    - Implements domain interfaces.
- **Repository Layer (`internal/repository`)**:
    - Data access implementations (PostgreSQL, Redis).
    - Adapts infrastructure to domain ports.
- **Handler Layer (`internal/handler/http/v1`)**:
    - Delivery layer (REST API).
    - Parses requests, validates DTOs, and calls services.
- **App Layer (`internal/app`)**:
    - Composition root (Dependency Injection).
    - Wiring services, repositories, and routes.
- **Pkg Layer (`internal/pkg`)**:
    - Shared utilities (Config, Database pool, Logger, Response helpers).

## Feature Modules

- **Auth & User**: Secure JWT-based authentication and user profile management.
- **Accounts**: Multi-currency support for Bank, Cash, Savings, and Credit accounts.
- **Transactions**: Double-entry ledger support with categories, tags, and line items.
- **Budgets**: Category-based budget tracking with period analysis.
- **Debt & Contacts**: Management of personal debts, "Split Bill" logic, and contact synchronization.
- **Investments**: Stock/Security portfolio management using FIFO cost-basis.
- **Market Data**: Real-time and historical security price synchronization via Redis Streams.
- **Savings (Instrument)**: Term deposits and savings goals management.
- **Rotating Savings (Hụi/Họ)**: Professional management of rotating savings groups with auto-calculated schedules.
- **Reports**: Financial health dashboards and period-over-period analysis.

## Getting Started

### Prerequisites
- [Go 1.26.1+](https://go.dev/)
- [Docker](https://www.docker.com/)

### Local Development (Recommended)
The easiest way to run the project is using the integrated Mono-repo environment:

1.  **Prepare Environment**:
    ```bash
    cp .env.example .env
    ```
    *Note: The default `.env.example` is configured to work with the `sunflower` mono-repo infrastructure.*

2.  **Run with Docker Compose**:
    ```bash
    docker compose -f docker-compose.dev.yml up -d --build
    ```

3.  **Check Logs**:
    ```bash
    docker compose -f docker-compose.dev.yml logs -f goen-api
    ```

### Manual Run
If you have Postgres and Redis running locally:
```bash
go mod tidy
go run ./cmd/api
```

## API Reference

- **Swagger UI**: `http://localhost:8080/swagger/`
- **Health Check**: `GET /api/v1/health`

### Core v1 Endpoints
- `POST /api/v1/auth/login`
- `GET /api/v1/accounts`
- `POST /api/v1/transactions`
- `GET /api/v1/investment-accounts/{id}/holdings`
- `POST /api/v1/rotating-savings/groups`
- `GET /api/v1/public/profile/{username}`

## Development

### Tools
- **Air**: Hot-reloading for development.
- **Goose**: Database migration management (auto-run on startup if `MIGRATE_ON_START=true`).
- **Swag**: Swagger documentation generation.

### Migrations
New migrations should be placed in the `migrations/` directory.

```bash
# Add a new migration
# goose -dir migrations create <name> sql
```

## License
Private / Proprietary.
