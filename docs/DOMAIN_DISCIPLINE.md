# Domain Discipline

This document is the non-optional discipline checklist for domain development in goen-api-v2.

Primary companion document:

1. `docs/DOMAIN_INTERFACE_NAMING_CONVENTION.md` for mandatory interface naming and module wiring rules.

## 1) Boundary First

1. One domain = one responsibility area.
2. Do not put cross-domain behavior into a convenient existing domain.
3. Route namespace must match domain namespace.
4. If a feature grows beyond current boundary, create a new domain module.

## 2) Required Domain Shape

Each domain must follow this minimal shape:

1. `types.go`
2. `service.go`
3. `handler.go`
4. `routes.go`
5. `module.go`

Rules:

1. Handler depends on `Service` interface, not repository, not infra client.
2. Service depends on domain ports (for example repository/storage interfaces).
3. Module wires dependencies and exposes routes.

## 3) Layer Responsibilities

1. Handler layer:
   - parse request
   - get params/context
   - call service
   - write response
   - no business logic, no persistence logic
2. Service layer:
   - validation and business rules
   - error mapping to app error kinds
   - orchestration between ports
3. Repository/infra layer:
   - SQL, external API, object storage, transport details
   - no HTTP concerns

## 3.1 Repository Ownership Model

Repository ownership is domain/use-case oriented, not strictly one-table-one-repository.

Rules:

1. One table should have one primary write owner repository.
2. Cross-table reads and joins are allowed when serving the same domain use case.
3. Cross-domain projection reads are allowed (for example analytics/reporting style queries).
4. Cross-domain writes should be avoided unless explicitly owned by domain boundary decision.

Examples in current codebase:

1. `account` repository writes `accounts` + `user_accounts` to preserve ownership linkage.
2. `transaction` repository writes `transactions` + `transaction_line_items` + `transaction_line_item_tags` + `group_expense_participants` as transaction extension persistence.
3. `debt` repository writes `debts` + `debt_payment_links` + `debt_installments` as debt aggregate persistence.

## 4) Forbidden Shortcuts

1. No handler -> repository direct calls.
2. No handler -> infra direct calls.
3. No cross-domain imports of concrete services when an interface port is expected.
4. No domain with only handler+routes for non-trivial feature.

## 5) Error Discipline

1. Use `apperrors` kinds for all service errors.
2. Return `validation`, `not_found`, `conflict`, `unauthorized`, `internal` intentionally.
3. Do not leak raw infrastructure errors to handler responses.

## 6) Test Discipline (Minimum)

Each domain should include:

1. `module_test.go` for wiring sanity.
2. `handler_test.go` for route/contract behavior.
3. `service_test.go` for business rules and error mapping.

Minimum negative coverage:

1. invalid payload/path
2. not found
3. conflict/unauthorized when applicable
4. internal error fallback

## 7) Pull Request Checklist

Before merge:

1. Confirm route prefix follows domain ownership.
2. Confirm handler does not import repository or infra concrete types.
3. Confirm service exists and owns business logic.
4. Confirm tests include positive + negative cases.
5. Run:
   - `go test ./...`
   - `make quality`

## 8) Current Enforcement Notes

1. `auth` domain is auth-only (signup/signin/login/register/refresh).
2. `media` domain owns media proxy endpoints (protected route under auth middleware).
3. `profile` and `setting` domains own user profile/settings operations with dedicated services.
4. `category`, `tag`, `budget`, `contact`, `debt`, `investment`, `savings`, `rotating_savings`, and `report` are migrated with interface-driven service/module wiring.
5. `debt` domain owns debt payment/installment routes and integrates with `transaction`/`contact` via interfaces in `types.go`.
6. `group_expense` is intentionally kept as a `transaction` extension (extension fields + participants table), not a standalone domain module at this stage.

## 9) Compliance Tracking

Use `docs/DOMAIN_COMPLIANCE_MATRIX.md` as the source of truth for interface/module wiring compliance across domains.

Rules:

1. Update matrix status in every domain-related PR.
2. Mark domain as `PASS` only when interface/module wiring requirements are verified.
3. Keep planned domains as `PENDING` until migration starts.

If a future change breaks the rules above, stop and refactor before continuing.