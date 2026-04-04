# Domain Interface Naming Convention

This document defines mandatory naming and wiring conventions for interfaces in all domains.

## 1) Mandatory Interface Names

Every domain must define interface names in `types.go` using these rules:

1. `Service` for domain use-case contract consumed by handler.
2. `Repository` suffix for persistence ports (for example `UserRepository`, `TransactionRepository`).
3. `Storage` suffix for object/file storage ports (for example `MediaStorage`, `AvatarStorage`).
4. `Client` suffix for external HTTP/RPC ports (for example `RateLimitClient`).

Do not use `IService`, `IRepo`, or `Impl` suffixes.

## 2) Mandatory Constructor and Struct Names

1. Concrete service implementation struct name should be lowercase `service`.
2. Service constructor must be `NewService(...)`.
3. Handler constructor must be `NewHandler(service Service)`.
4. Module constructor must be `NewModule(deps ModuleDeps)`.

## 3) Mandatory Module Shape

Each domain module should follow this shape:

1. `type ModuleDeps struct { ...; Service Service }`
2. `type Module struct { Service Service; Handler *Handler }`
3. `NewModule` behavior:
   - use `deps.Service` if provided
   - otherwise build default service from ports in `deps`
   - always inject `Service` into handler

Reference pattern:

```go
func NewModule(deps ModuleDeps) *Module {
    svc := deps.Service
    if svc == nil {
        svc = NewService(deps.Repository, deps.Storage)
    }
    h := NewHandler(svc)
    return &Module{Service: svc, Handler: h}
}
```

## 4) Handler Dependency Rule

Handler must only depend on `Service` interface.

Forbidden in handler:

1. concrete repository types
2. concrete infra/storage clients
3. SQL or transport-specific logic

## 5) Service Method Naming Rule

Service method names should use domain verbs and stable intent:

1. `Get...`, `List...`, `Create...`, `Update...`, `Delete...`, `Change...`, `Upload...`
2. include ownership scope when relevant (for example `GetMe`, `UpdateMySettings`)
3. avoid transport words (`Handle`, `HTTP`, `Route`) in service methods

## 6) Compile-Time Contract Check (Recommended)

Use compile-time assertions in implementation files:

```go
var _ Service = (*service)(nil)
```

This is recommended for all new domains and refactors.

## 7) Domain Compliance Matrix

Compliance status is tracked in:

1. `docs/DOMAIN_COMPLIANCE_MATRIX.md`

This avoids status drift in this convention document and provides one source of truth for migrated vs pending domains.

## 8) Pull Request Gate

A domain-related PR should be blocked if:

1. no `Service` interface exists for the domain
2. handler is wired to repository/storage directly
3. module does not expose `Service` in `Module`
4. naming violates this convention
