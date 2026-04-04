# Domain Compliance Matrix

This table tracks domain compliance for interface and module wiring discipline.

Status legend:

1. `PASS`: domain is migrated and meets required interface/module wiring rules.
2. `PENDING`: domain is not migrated yet, compliance will be checked during migration.

## Compliance Rules In Scope

1. `types.go` defines `Service interface`.
2. `ModuleDeps` includes `Service Service`.
3. `Module` includes `Service Service`.
4. `NewModule(deps)` supports injected `deps.Service` with fallback constructor.
5. `Handler` depends on `Service` interface (not repository/infra concrete types).

## Current Matrix

| Domain | Migrated in v2 | Interface + Module Wiring | Notes |
|---|---|---|---|
| auth | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| account | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| transaction | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| profile | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| setting | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| media | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| debt | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| group_expense | n/a (transaction extension) | PASS | Implemented as transaction extension (`group_participants`, `owner_original_amount`) with create + patch support in transaction module |
| savings | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| rotating_savings | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| investment | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| category | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| tag | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| budget | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| contact | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| report | yes | PASS | Compliant: `Service` in `types.go`, module injection/fallback, handler via service interface |
| public | no | PENDING | Planned migration wave: pending domain extraction |
| marketdata | no | PENDING | Planned migration wave: pending domain extraction |

## Maintenance Rule

Update this matrix in every domain-related PR:

1. When a domain is newly migrated, change status from `PENDING` to `PASS` only after tests and quality checks pass.
2. If a migrated domain violates interface/module wiring rules, set status to `FAIL` and fix before merge.
