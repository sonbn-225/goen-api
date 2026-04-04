# goen-api-v2

goen-api-v2 is the clean architecture implementation for the next generation of goen-api.

## Goals

- Keep operational baseline stable (Docker, Jenkins, migrations)
- Move from all-in-one wiring to clean domain modules
- Separate concerns clearly: transport, usecase, domain ports, infrastructure adapters
- Keep API path stable on v1 during development

## Current scope

Wave 1 implementation includes:

- Core platform foundation (config, error mapping, response helpers, auth middleware)
- Auth domain
- Account domain
- Transaction domain
- Category domain
- Tag domain
- Budget domain
- Contact domain
- Debt domain
- Investment domain
- Rotating savings domain
- Report domain
- Savings domain

## Architecture

```text
cmd/api                 -> process bootstrap
internal/app            -> composition root and route registration
internal/core           -> shared platform concerns (config/httpx/errors/response/security)
internal/domains/*      -> domain modules (handler + usecase + ports + entities)
internal/infra/*        -> infrastructure clients and adapters (postgres pool, migrations)
internal/repository/*   -> repository implementations by data source
migrations              -> preserved SQL migrations from v1
```

## Run locally

```bash
go mod tidy
go run ./cmd/api
```

Server defaults to `:8080`.

Health endpoint:

- `GET /healthz`

Swagger:

- Generate docs: `make swagger`
- Swagger UI: `GET /swagger/`

v1 endpoints:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/signup`
- `POST /api/v1/auth/signin`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me`
- `PATCH /api/v1/auth/me/profile`
- `PATCH /api/v1/auth/me/settings`
- `POST /api/v1/auth/me/change-password`
- `GET /api/v1/accounts`
- `POST /api/v1/accounts`
- `GET /api/v1/categories`
- `GET /api/v1/categories/{categoryId}`
- `GET /api/v1/tags`
- `POST /api/v1/tags`
- `GET /api/v1/tags/{tagId}`
- `GET /api/v1/budgets`
- `POST /api/v1/budgets`
- `GET /api/v1/budgets/{budgetId}`
- `GET /api/v1/contacts`
- `POST /api/v1/contacts`
- `GET /api/v1/contacts/{contactId}`
- `PATCH /api/v1/contacts/{contactId}`
- `DELETE /api/v1/contacts/{contactId}`
- `GET /api/v1/debts`
- `POST /api/v1/debts`
- `GET /api/v1/debts/{debtId}`
- `GET /api/v1/debts/{debtId}/payments`
- `POST /api/v1/debts/{debtId}/payments`
- `GET /api/v1/debts/{debtId}/installments`
- `POST /api/v1/debts/{debtId}/installments`
- `GET /api/v1/rotating-savings/groups`
- `POST /api/v1/rotating-savings/groups`
- `GET /api/v1/rotating-savings/groups/{groupId}`
- `PATCH /api/v1/rotating-savings/groups/{groupId}`
- `DELETE /api/v1/rotating-savings/groups/{groupId}`
- `GET /api/v1/rotating-savings/groups/{groupId}/contributions`
- `POST /api/v1/rotating-savings/groups/{groupId}/contributions`
- `DELETE /api/v1/rotating-savings/groups/{groupId}/contributions/{contributionId}`
- `GET /api/v1/savings/instruments`
- `POST /api/v1/savings/instruments`
- `GET /api/v1/savings/instruments/{instrumentId}`
- `PATCH /api/v1/savings/instruments/{instrumentId}`
- `DELETE /api/v1/savings/instruments/{instrumentId}`
- `GET /api/v1/reports/dashboard`
- `GET /api/v1/transactions/{transactionId}/debt-links`
- `GET /api/v1/investment-accounts`
- `GET /api/v1/investment-accounts/{investmentAccountId}`
- `PATCH /api/v1/investment-accounts/{investmentAccountId}`
- `POST /api/v1/investment-accounts/{investmentAccountId}/trades`
- `GET /api/v1/investment-accounts/{investmentAccountId}/trades`
- `GET /api/v1/investment-accounts/{investmentAccountId}/holdings`
- `GET /api/v1/securities`
- `GET /api/v1/securities/{securityId}`
- `GET /api/v1/securities/{securityId}/prices-daily`
- `GET /api/v1/securities/{securityId}/events`
- `GET /api/v1/transactions`
- `POST /api/v1/transactions`
- `PATCH /api/v1/transactions/{transactionId}`
- `GET /api/v1/transactions/{transactionId}/group-expense-participants`

Transaction group expense extension (same `POST /api/v1/transactions` endpoint):

1. `group_participants` and `owner_original_amount` are optional extension fields for `expense` transactions.
2. For non-`expense` types, sending these extension fields returns validation errors.
3. When `group_participants[].share_amount` is omitted, service calculates share values proportionally from `original_amount`.
4. Transaction and group expense participants are created atomically in one repository transaction.

Transaction line items extension (same `POST /api/v1/transactions` endpoint):

1. `line_items` is required for non-transfer transaction types.
2. For `transfer`, `line_items` must be empty.
3. Each `line_items[]` item requires `category_id` and `amount > 0`.
4. Transaction amount is normalized from sum of `line_items[].amount` (legacy parity behavior).
5. `transaction_line_items` and `transaction_line_item_tags` are persisted atomically with the transaction row.

## Notes

- Database migrations are still preserved and can be executed with `MIGRATE_ON_START=true` and `DATABASE_URL` configured.
- Auth/Account/Transaction currently use Postgres-backed repository implementations in `internal/repository`.
- The infra layer is intentionally limited to external systems connectivity (for example Postgres pool and migrations).
- Air dev mode auto-runs Swagger generation before rebuild (`swag init ... -o docs`).
- `PATCH /api/v1/transactions/{transactionId}` now supports `group_participants` replacement for unsettled participants on expense transactions.

## Domain Discipline

- Read and follow `docs/DOMAIN_DISCIPLINE.md` before adding or refactoring any domain module.
- Enforce interface naming and service-layer wiring via `docs/DOMAIN_INTERFACE_NAMING_CONVENTION.md`.
- Track migrated/pending domain compliance in `docs/DOMAIN_COMPLIANCE_MATRIX.md`.
