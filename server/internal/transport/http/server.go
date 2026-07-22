package httpapi

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/internal/app"
	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
)

type Authenticator interface {
	ExchangeCode(code, state string) (string, error)
	Authenticate(authHeader string) (*domain.Identity, error)
}

type RuntimeConfigProvider interface {
	Get(ctx context.Context, accessKey, secret, namespace, configName string) (*model.Config, error)
	List(ctx context.Context, accessKey, secret, namespace string) ([]*model.Config, error)
}

type Deps struct {
	Namespaces     *app.NamespaceUsecase
	Configs        *app.ConfigUsecase
	AccessKeys     *app.AccessKeyUsecase
	RuntimeConfigs RuntimeConfigProvider
	Auth           Authenticator
}

type handlers struct {
	deps Deps
}

const adminBasePath = "/api/v1/admin"

func New(deps Deps) (*gin.Engine, error) {
	if deps.Namespaces == nil || deps.Configs == nil || deps.AccessKeys == nil || deps.RuntimeConfigs == nil || deps.Auth == nil {
		return nil, fmt.Errorf("httpapi: all dependencies are required")
	}

	r := gin.Default()
	h := &handlers{deps: deps}
	r.POST("/api/v1/auth/callback", h.exchangeOAuthCode)
	h.registerRuntimeConfigRoutes(r)

	cfg := huma.DefaultConfig("PCC Config Center", "1.0.0")
	if cfg.Components.SecuritySchemes == nil {
		cfg.Components.SecuritySchemes = make(map[string]*huma.SecurityScheme)
	}
	cfg.Components.SecuritySchemes["bearerAuth"] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
	cfg.Security = []map[string][]string{{"bearerAuth": {}}}

	// Register all Huma routes on an authenticated router group. This makes
	// authentication default-deny for every current and future business route.
	protected := r.Group("")
	protected.Use(AuthMiddleware(deps.Auth.Authenticate))
	api := humagin.NewWithGroup(r, protected, cfg)
	h.register(api)

	return r, nil
}

func (h *handlers) register(api huma.API) {
	h.registerUserRoutes(api)
	h.registerNamespaceRoutes(api)
	h.registerConfigRoutes(api)
	h.registerAccessKeyRoutes(api)
	h.registerSystemRoutes(api)
}
