# PCC Config Center

PCC is split into a Go administration API and a React administration console.

## Authentication boundary

- `POST /api/v1/auth/callback` is the only endpoint that requires no
  credentials.
- Every Huma business route is registered through an authenticated Gin router
  group and lives under `/api/v1/admin`.
- `/api/v1/runtime/*` is isolated from the administration API and requires an
  enabled access key, its secret, and an explicit namespace grant.
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

## Service configuration API

Grant an access key read-only or read-write access to a namespace in the
console, then send both credentials as headers:

```sh
curl \
  -H 'X-PCC-Access-Key: pcc_ak_...' \
  -H 'X-PCC-Secret: pcc_sk_...' \
  http://localhost:8080/api/v1/runtime/namespaces/production/configs/database
```

List every configuration visible to that key:

```text
GET /api/v1/runtime/namespaces/{namespace}/configs
```

Go services can use the standalone module in `sdk/pcc-go-sdk`:

```go
import pccclient "github.com/itgrape/pcc/sdk/pcc-go-sdk"

type ApplicationConfig struct {
    Server struct {
        Host    string        `json:"host" yaml:"host"`
        Port    int           `json:"port" yaml:"port"`
        Timeout time.Duration `json:"timeout" yaml:"timeout"`
    } `json:"server" yaml:"server"`
}

client, err := pccclient.New(
    pccURL,
    accessKey,
    secret,
    pccclient.WithNamespace("production"),
    pccclient.WithConfig("application"),
    pccclient.WithTimeout(5*time.Second),
)
if err != nil {
    return err
}

var config ApplicationConfig
if err := client.Load(&config); err != nil {
    return err
}

var port int
if err := client.LoadField("server.port", &port); err != nil {
    return err
}
```

`Load` decodes the configured JSON, YAML, Properties, or Text document directly
into a pointer. `LoadField` supports dotted paths and array indexes such as
`features[0].enabled`. Services that select namespace and configuration at
runtime can use `LoadConfig(ctx, namespace, name, &target)` and
`LoadConfigField(ctx, namespace, name, path, &target)`. Raw metadata remains
available through `GetConfig` and `ListConfigs`.

Credentials are accepted only through request headers. Disabled keys, invalid
secrets, and ungranted namespaces are rejected before configuration data is
read.

## Verification

```sh
cd server && go test ./... && go vet ./...
cd web && npm run generate:api && npm run build && npm run lint
```
