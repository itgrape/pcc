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
	Login(username, password string) (string, error)
	Authenticate(token string) (*domain.Identity, error)
	Logout(token string)
	SecureCookie() bool
}

type RuntimeConfigProvider interface {
	Get(ctx context.Context, accessKey, secret, namespace, configName string) (*model.Config, error)
	List(ctx context.Context, accessKey, secret, namespace string) ([]*model.Config, error)
}

type AdminDeps struct {
	Namespaces *app.NamespaceUsecase
	Configs    *app.ConfigUsecase
	AccessKeys *app.AccessKeyUsecase
	Auth       Authenticator
}

type RuntimeDeps struct {
	Configs RuntimeConfigProvider
}

type dependencies struct {
	AdminDeps
	RuntimeConfigs RuntimeConfigProvider
}

type handlers struct {
	deps    dependencies
	limiter loginLimiter
}

const adminBasePath = "/api/v1/admin"

func NewAdmin(deps AdminDeps) (*gin.Engine, error) {
	if deps.Namespaces == nil || deps.Configs == nil || deps.AccessKeys == nil || deps.Auth == nil {
		return nil, fmt.Errorf("httpapi: all admin dependencies are required")
	}

	r := gin.Default()
	h := &handlers{deps: dependencies{AdminDeps: deps}, limiter: loginLimiter{attempts: make(map[string]loginAttempt)}}
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(204)
	})
	r.POST("/api/v1/auth/login", requireSameOrigin(), h.login)
	r.POST("/api/v1/auth/logout", requireSameOrigin(), h.logout)

	cfg := huma.DefaultConfig("PCC Config Center", "1.0.0")
	if cfg.Components.SecuritySchemes == nil {
		cfg.Components.SecuritySchemes = make(map[string]*huma.SecurityScheme)
	}
	cfg.Components.SecuritySchemes["sessionCookie"] = &huma.SecurityScheme{
		Type: "apiKey",
		In:   "cookie",
		Name: sessionCookieName,
	}
	cfg.Security = []map[string][]string{{"sessionCookie": {}}}

	// Register all Huma routes on an authenticated router group. This makes
	// authentication default-deny for every current and future business route.
	protected := r.Group("")
	protected.Use(AuthMiddleware(deps.Auth.Authenticate), requireSameOrigin(), func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	})
	api := humagin.NewWithGroup(r, protected, cfg)
	h.register(api)
	if err := registerSPA(r); err != nil {
		return nil, err
	}

	return r, nil
}

func NewRuntime(deps RuntimeDeps) (*gin.Engine, error) {
	if deps.Configs == nil {
		return nil, fmt.Errorf("httpapi: runtime config provider is required")
	}

	r := gin.Default()
	h := &handlers{deps: dependencies{RuntimeConfigs: deps.Configs}}
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(204)
	})
	h.registerRuntimeConfigRoutes(r)
	return r, nil
}

func (h *handlers) register(api huma.API) {
	h.registerUserRoutes(api)
	h.registerNamespaceRoutes(api)
	h.registerConfigRoutes(api)
	h.registerAccessKeyRoutes(api)
	h.registerSystemRoutes(api)
}
