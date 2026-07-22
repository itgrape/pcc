package httpapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/internal/app"
	"github.com/itgrape/pcc/server/internal/infra"
	"github.com/itgrape/pcc/server/internal/repository"
)

type Deps struct {
	Namespaces *app.NamespaceUsecase
	Configs    *app.ConfigUsecase
	AccessKeys *app.AccessKeyUsecase
	Bus        *app.EventBus
	Store      *repository.Store
	Auth       *infra.Casdoor
}

type handlers struct {
	deps Deps
}

func New(deps Deps) *gin.Engine {
	r := gin.Default()
	h := &handlers{deps: deps}
	r.POST("/api/v1/auth/callback", h.exchangeOAuthCode)

	cfg := huma.DefaultConfig("PCC Config Center", "1.0.0")
	api := humagin.New(r, cfg)

	// All Huma business routes below are administrative APIs.
	r.Use(AuthMiddleware(deps.Auth.GetUserFromToken))
	h.register(api)

	return r
}

func (h *handlers) register(api huma.API) {
	// huma 注册接口（huma handler 的上下文通过 ctx.Value 传递）
	base := "/api/v1/admin"

	huma.Register(api, huma.Operation{
		OperationID: "get-user-info", Method: http.MethodGet, Path: base + "/user/info",
		Summary: "获取当前用户信息", Tags: []string{"user"},
	}, h.getUserInfo)

	// ---- 命名空间（所有 id 参数统一用 namespace_id）----
	huma.Register(api, huma.Operation{
		OperationID: "list-namespaces", Method: http.MethodGet, Path: base + "/namespaces",
		Summary: "命名空间列表", Tags: []string{"namespaces"},
	}, h.listNamespaces)
	huma.Register(api, huma.Operation{
		OperationID: "get-namespace", Method: http.MethodGet, Path: base + "/namespaces/{namespace_id}",
		Summary: "命名空间详情", Tags: []string{"namespaces"},
	}, h.getNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "create-namespace", Method: http.MethodPost, Path: base + "/namespaces",
		Summary: "创建命名空间", Tags: []string{"namespaces"}, DefaultStatus: http.StatusCreated,
	}, h.createNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "update-namespace", Method: http.MethodPut, Path: base + "/namespaces/{namespace_id}",
		Summary: "更新命名空间", Tags: []string{"namespaces"},
	}, h.updateNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "delete-namespace", Method: http.MethodDelete, Path: base + "/namespaces/{namespace_id}",
		Summary: "删除命名空间", Tags: []string{"namespaces"},
	}, h.deleteNamespace)

	// ---- 配置（id 参数统一用 config_id）----
	huma.Register(api, huma.Operation{
		OperationID: "list-configs", Method: http.MethodGet, Path: base + "/namespaces/{namespace_id}/configs",
		Summary: "配置列表", Tags: []string{"configs"},
	}, h.listConfigs)
	huma.Register(api, huma.Operation{
		OperationID: "get-config", Method: http.MethodGet, Path: base + "/configs/{config_id}",
		Summary: "配置详情", Tags: []string{"configs"},
	}, h.getConfig)
	huma.Register(api, huma.Operation{
		OperationID: "create-config", Method: http.MethodPost, Path: base + "/namespaces/{namespace_id}/configs",
		Summary: "创建配置", Tags: []string{"configs"}, DefaultStatus: http.StatusCreated,
	}, h.createConfig)
	huma.Register(api, huma.Operation{
		OperationID: "update-config", Method: http.MethodPut, Path: base + "/configs/{config_id}",
		Summary: "更新配置", Tags: []string{"configs"},
	}, h.updateConfig)
	huma.Register(api, huma.Operation{
		OperationID: "delete-config", Method: http.MethodDelete, Path: base + "/configs/{config_id}",
		Summary: "删除配置", Tags: []string{"configs"},
	}, h.deleteConfig)

	// ---- 历史 & 回滚 ----
	huma.Register(api, huma.Operation{
		OperationID: "list-config-history", Method: http.MethodGet, Path: base + "/configs/{config_id}/history",
		Summary: "配置历史版本列表", Tags: []string{"configs"},
	}, h.listConfigHistory)
	huma.Register(api, huma.Operation{
		OperationID: "get-config-history", Method: http.MethodGet, Path: base + "/config/history/{history_id}",
		Summary: "历史版本详情", Tags: []string{"configs"},
	}, h.getConfigHistory)
	huma.Register(api, huma.Operation{
		OperationID: "rollback-config", Method: http.MethodPost, Path: base + "/configs/{config_id}/rollback/{history_id}",
		Summary: "回滚配置", Tags: []string{"configs"},
	}, h.rollbackConfig)

	// ---- 密钥（id 参数统一用 key_id）----
	huma.Register(api, huma.Operation{
		OperationID: "list-access-keys", Method: http.MethodGet, Path: base + "/access-keys",
		Summary: "密钥列表", Tags: []string{"access-keys"},
	}, h.listAccessKeys)
	huma.Register(api, huma.Operation{
		OperationID: "get-access-key", Method: http.MethodGet, Path: base + "/access-keys/{key_id}",
		Summary: "密钥详情", Tags: []string{"access-keys"},
	}, h.getAccessKey)
	huma.Register(api, huma.Operation{
		OperationID: "create-access-key", Method: http.MethodPost, Path: base + "/access-keys",
		Summary: "创建密钥", Tags: []string{"access-keys"}, DefaultStatus: http.StatusCreated,
	}, h.createAccessKey)
	huma.Register(api, huma.Operation{
		OperationID: "update-access-key", Method: http.MethodPut, Path: base + "/access-keys/{key_id}",
		Summary: "更新密钥", Tags: []string{"access-keys"},
	}, h.updateAccessKey)
	huma.Register(api, huma.Operation{
		OperationID: "delete-access-key", Method: http.MethodDelete, Path: base + "/access-keys/{key_id}",
		Summary: "删除密钥", Tags: []string{"access-keys"},
	}, h.deleteAccessKey)

	// ---- 授权 ----
	huma.Register(api, huma.Operation{
		OperationID: "list-key-namespaces", Method: http.MethodGet, Path: base + "/access-keys/{key_id}/namespaces",
		Summary: "密钥已授权的命名空间", Tags: []string{"access-keys"},
	}, h.listKeyNamespaces)
	huma.Register(api, huma.Operation{
		OperationID: "grant-key-namespace", Method: http.MethodPost, Path: base + "/access-keys/{key_id}/namespaces",
		Summary: "新增授权", Tags: []string{"access-keys"}, DefaultStatus: http.StatusCreated,
	}, h.grantKeyNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "update-key-namespace-permission", Method: http.MethodPut, Path: base + "/access-keys/{key_id}/namespaces/{namespace_id}",
		Summary: "更新授权", Tags: []string{"access-keys"},
	}, h.updateKeyNamespacePermission)
	huma.Register(api, huma.Operation{
		OperationID: "revoke-key-namespace", Method: http.MethodDelete, Path: base + "/access-keys/{key_id}/namespaces/{namespace_id}",
		Summary: "解除授权", Tags: []string{"access-keys"},
	}, h.revokeKeyNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "list-namespace-access-keys", Method: http.MethodGet, Path: base + "/namespaces/{namespace_id}/access-keys",
		Summary: "命名空间下的授权密钥", Tags: []string{"namespaces"},
	}, h.listNamespaceAccessKeys)

	// ---- 系统 ----
	huma.Register(api, huma.Operation{
		OperationID: "get-system-enums", Method: http.MethodGet, Path: base + "/system/enums",
		Summary: "系统枚举值", Tags: []string{"system"},
	}, h.getSystemEnums)
}
