# PCC Config Center

PCC is split into a Go administration API and a React administration console.

## Authentication boundary

- `POST /api/v1/auth/callback` is the only public application endpoint.
- Every Huma business route is registered through an authenticated Gin router
  group and lives under `/api/v1/admin`.
- The router test iterates every administrative route and asserts that an
  unauthenticated request receives `401 Unauthorized`.
- Casdoor tokens are verified by the backend; frontend route guards are only a
  user-experience layer and are not a security boundary.

## Backend structure

- `server/cmd/server`: composition root and HTTP server lifecycle.
- `server/internal/config`: environment configuration.
- `server/internal/infra`: database and identity-provider adapters.
- `server/internal/app`: application use cases.
- `server/internal/domain`: domain errors and shared domain values.
- `server/internal/repository`: persistence implementations and transactions.
- `server/internal/transport/http`: HTTP routing, middleware, request/response
  mapping, and resource-specific handlers.
- `server/internal/query`: generated database query code.

Transport handlers depend on application use cases, not repositories directly.
All route registration stays next to the corresponding resource handler.

## Frontend structure

- `web/src/app`: application providers, router, and shared runtime setup.
- `web/src/features`: feature-owned state and screens, such as authentication.
- `web/src/layouts`: route layouts.
- `web/src/pages`: top-level business pages.
- `web/src/components/ui`: reusable presentational components.
- `web/src/shared/api`: the configured Axios client and cross-cutting API logic.
- `web/src/lib/orval`: generated API clients and schemas.

Generated clients must use the Orval mutator in `web/src/shared/api/http.ts`;
application code must not mutate Axios defaults.

## Verification

```sh
cd server && go test ./... && go vet ./...
cd web && npm run generate:api && npm run build && npm run lint
```
