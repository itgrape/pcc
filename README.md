# PCC 配置中心

PCC 是一个面向服务端应用的轻量配置中心，提供配置管理、版本历史、访问密钥、命名空间授权和 Go SDK。管理端由 Go API 与 React 控制台组成，生产环境使用一个容器、一个 Go 进程同时提供 API 和前端页面。

## 核心能力

- 使用命名空间隔离业务、团队和环境配置。
- 支持 JSON、YAML、Properties 和纯文本配置。
- 记录配置版本历史，支持查看和回滚。
- 使用 Access Key 与 Secret 进行服务间鉴权。
- Access Key 可绑定多个命名空间并设置只读或读写权限。
- 使用 Casdoor 完成管理控制台登录。
- 提供独立 Go SDK，可将配置直接解析到结构体或读取单个字段。
- 提供基于 Podman 的 Gitea 自动构建、部署、健康检查、回滚和清理流程。

## 目录结构

```text
.
├── server/                 Go 后端
│   ├── cmd/server/         服务入口与生命周期
│   ├── internal/app/       应用用例
│   ├── internal/domain/    领域类型与错误
│   ├── internal/infra/     数据库与 Casdoor 适配器
│   ├── internal/repository/持久化与事务
│   └── internal/transport/ HTTP 路由、中间件和页面托管
├── web/                    React 管理控制台
├── sdk/pcc-go-sdk/         独立 Go SDK Module
├── schema/                 数据库结构
├── Containerfile           前后端统一镜像
└── .gitea/workflows/       Gitea 自动部署工作流
```

前端 Orval 客户端统一使用 `web/src/shared/api/http.ts` 中的 Axios 实例，不修改 Axios 全局默认值。

## 鉴权边界

- `/healthz` 和 `POST /api/v1/auth/callback` 为公开引导接口。
- `/api/v1/admin/*` 使用 Casdoor Bearer Token，并由后端统一鉴权。
- `/api/v1/runtime/*` 使用 Access Key、Secret 和命名空间授权。
- 前端路由守卫仅用于交互体验，不作为安全边界。
- 后端回归测试会遍历管理接口，确保未登录请求返回 `401 Unauthorized`。

## 本地开发

### 后端配置

复制示例文件：

```sh
cp server/.env.example server/.env
```

至少配置以下变量：

```dotenv
PCC_DB_DSN=user:pass@tcp(127.0.0.1:3306)/pcc?charset=utf8mb4&parseTime=true&loc=Local
PCC_CASDOOR_ENDPOINT=https://door.example.com
PCC_CASDOOR_CLIENT_ID=pcc-web
PCC_CASDOOR_CLIENT_SECRET=replace-me
PCC_CASDOOR_ORGANIZATION=built-in
PCC_CASDOOR_APPLICATION=app-pcc
```

启动后端：

```sh
cd server
go run ./cmd/server
```

### 前端配置

复制示例文件并填写浏览器可公开的 Casdoor 参数：

```sh
cp web/.env.example web/.env
cd web
npm ci
npm run dev
```

Vite 会将 `/api` 代理到 `VITE_API_PROXY_TARGET`。

## 服务读取配置

先在控制台创建 Access Key，并将其授权给目标命名空间。

读取单个配置：

```sh
curl \
  -H 'X-PCC-Access-Key: pcc_ak_...' \
  -H 'X-PCC-Secret: pcc_sk_...' \
  http://localhost:8080/api/v1/runtime/namespaces/production/configs/application
```

列出命名空间下的配置：

```text
GET /api/v1/runtime/namespaces/{namespace}/configs
```

禁用的密钥、错误的 Secret 或未授权的命名空间会在读取配置前被拒绝。

## Go SDK

SDK 是独立 Module：

```sh
go get github.com/itgrape/pcc/sdk/pcc-go-sdk
```

完整配置解析：

```go
import (
    "time"

    pccclient "github.com/itgrape/pcc/sdk/pcc-go-sdk"
)

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
```

读取单个字段：

```go
var port int
if err := client.LoadField("server.port", &port); err != nil {
    return err
}
```

`LoadField` 支持点路径和数组下标，例如 `features[0].enabled`。动态选择配置时可使用 `LoadConfig` 和 `LoadConfigField`，原始数据可通过 `GetConfig` 和 `ListConfigs` 获取。

## 统一容器镜像

Containerfile 使用 Node 和 Go 多阶段构建。最终镜像：

- 仅包含 Go 服务、CA 证书、时区数据和前端产物。
- 使用非 root 用户运行。
- 前端固定放在 `/web`。
- Go 服务自动检测 `/web/index.html` 并托管 SPA。
- `/assets/*` 使用长期缓存，HTML 使用 `no-cache`。
- 内置 `/healthz` 容器健康检查。
- 不在镜像中保存后端环境变量或密钥。

手工构建：

```sh
podman build -f Containerfile -t pcc:latest \
  --build-arg PCC_CASDOOR_ENDPOINT=https://door.example.com \
  --build-arg PCC_CASDOOR_CLIENT_ID=pcc-web \
  --build-arg PCC_CASDOOR_APPLICATION=app-pcc \
  --build-arg PCC_CASDOOR_ORGANIZATION=built-in \
  .
```

## Gitea 自动部署

工作流位于 `.gitea/workflows/deploy.yml`，使用 `runs-on: debian-latest`，直接调用宿主机 Podman。

### Gitea Variables

- `PCC_DEPLOY_NAME`：可选，镜像名和容器名，默认 `pcc`。
- `PCC_HOST_PORT`：可选，宿主机端口，默认 `8080`。
- `PCC_CASDOOR_ENDPOINT`：必填，Casdoor 服务地址。
- `PCC_CASDOOR_CLIENT_ID`：必填，Casdoor 客户端 ID。
- `PCC_CASDOOR_ORGANIZATION`：必填，Casdoor 组织名。
- `PCC_CASDOOR_APPLICATION`：必填，Casdoor 应用名。
- `PCC_SNOWFLAKE_NODE_ID`：可选，雪花 ID 节点号，默认 `1`。

### Gitea Secrets

- `PCC_DB_DSN`：数据库连接串。
- `PCC_CASDOOR_CLIENT_SECRET`：Casdoor 客户端密钥。

Casdoor 的公开参数只配置一份：构建镜像时映射为前端的
`VITE_CASDOOR_*`，启动容器时作为后端的 `PCC_CASDOOR_*` 使用。
工作流通过当前进程环境将运行参数直接传给 Podman，不创建 `.env`
文件或 Podman Secret，不在宿主机留下额外配置状态。

### 部署流程

1. 检出代码并检查宿主机 Podman。
2. 校验 Gitea Variables 和 Secrets。
3. 将当前 `latest` 标记为 `rollback`。
4. 构建带提交 SHA 标签的新镜像。
5. 直接从工作流环境替换并启动容器。
6. 等待 `/healthz` 变为健康状态。
7. 成功后更新 `latest`；失败则输出日志并自动启动 `rollback`。
8. 清理停止的容器、悬空层和旧的 PCC 镜像标签。

## 验证

```sh
cd server
go test ./...
go vet ./...

cd ../web
npm run generate:api
npm run build
npm run lint

cd ../sdk/pcc-go-sdk
go test ./...
go vet ./...
go test -race ./...
```
