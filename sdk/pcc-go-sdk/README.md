# PCC Go SDK

Typed configuration client for Go services.

```sh
go get github.com/itgrape/pcc/sdk/pcc-go-sdk
```

```go
import pccclient "github.com/itgrape/pcc/sdk/pcc-go-sdk"

client, err := pccclient.New(
    endpoint,
    accessKey,
    secret,
    pccclient.WithNamespace("production"),
    pccclient.WithConfig("application"),
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

Supported formats:

- JSON
- YAML
- Properties
- Text

Use `LoadConfig` and `LoadConfigField` when the namespace or configuration
name is selected dynamically. `GetConfig` and `ListConfigs` expose the raw
configuration metadata.
