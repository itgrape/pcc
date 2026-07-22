package httpapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/pkg/types"
)

const (
	runtimeBasePath    = "/api/v1/runtime"
	accessKeyHeader    = "X-PCC-Access-Key"
	accessSecretHeader = "X-PCC-Secret"
)

func (h *handlers) registerRuntimeConfigRoutes(r *gin.Engine) {
	runtime := r.Group(runtimeBasePath)
	runtime.GET("/namespaces/:namespace/configs", h.listRuntimeConfigs)
	runtime.GET("/namespaces/:namespace/configs/:config_name", h.getRuntimeConfig)
}

func (h *handlers) listRuntimeConfigs(c *gin.Context) {
	accessKey, secret, ok := runtimeCredentials(c)
	if !ok {
		writeRuntimeError(c, domain.ErrUnauthenticated)
		return
	}

	configs, err := h.deps.RuntimeConfigs.List(
		c.Request.Context(),
		accessKey,
		secret,
		c.Param("namespace"),
	)
	if err != nil {
		writeRuntimeError(c, err)
		return
	}

	result := make([]types.RuntimeConfig, 0, len(configs))
	for _, config := range configs {
		result = append(result, runtimeConfigDTO(config))
	}
	c.Header("Cache-Control", "private, no-cache")
	c.JSON(http.StatusOK, resOK(result))
}

func (h *handlers) getRuntimeConfig(c *gin.Context) {
	accessKey, secret, ok := runtimeCredentials(c)
	if !ok {
		writeRuntimeError(c, domain.ErrUnauthenticated)
		return
	}

	config, err := h.deps.RuntimeConfigs.Get(
		c.Request.Context(),
		accessKey,
		secret,
		c.Param("namespace"),
		c.Param("config_name"),
	)
	if err != nil {
		writeRuntimeError(c, err)
		return
	}

	etag := fmt.Sprintf(`"%d-%d"`, config.ID, config.Version)
	c.Header("Cache-Control", "private, no-cache")
	c.Header("ETag", etag)
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.JSON(http.StatusOK, resOK(runtimeConfigDTO(config)))
}

func runtimeCredentials(c *gin.Context) (string, string, bool) {
	accessKey := c.GetHeader(accessKeyHeader)
	secret := c.GetHeader(accessSecretHeader)
	return accessKey, secret, accessKey != "" && secret != ""
}

func runtimeConfigDTO(config *model.Config) types.RuntimeConfig {
	return types.RuntimeConfig{
		Name:        config.ConfigName,
		Format:      int(config.Format),
		Content:     config.Content,
		Description: config.Description,
		Version:     config.Version,
		UpdatedAt:   config.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func writeRuntimeError(c *gin.Context, err error) {
	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		domainErr = domain.ErrInternal
	}
	status := statusFor(domainErr.Kind)
	if status == http.StatusUnauthorized {
		c.Header("WWW-Authenticate", `PCC realm="runtime-config"`)
	}
	c.AbortWithStatusJSON(status, types.Response{
		Code:    status,
		Message: domainErr.Message,
	})
}
