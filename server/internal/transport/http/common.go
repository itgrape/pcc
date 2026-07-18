package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/pkg/types"
)

// ---- 统一响应辅助 ----

func resOK(data interface{}) types.Response {
	return types.Response{Code: 0, Message: "success", Data: data}
}

func resPage(list interface{}, total int64, page, pageSize int) types.Response {
	return resOK(types.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

// ---- 错误处理 ----

func toHTTPError(err error) error {
	if err == nil {
		return nil
	}
	var de *domain.Error
	if !errors.As(err, &de) {
		return huma.NewError(http.StatusInternalServerError, "internal error")
	}
	return huma.NewError(statusFor(de.Kind), de.Message)
}

func statusFor(k domain.Kind) int {
	switch k {
	case domain.KindValidation:
		return http.StatusBadRequest
	case domain.KindUnauthed:
		return http.StatusUnauthorized
	case domain.KindForbidden:
		return http.StatusForbidden
	case domain.KindNotFound:
		return http.StatusNotFound
	case domain.KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// ---- 上下文 key ----

type ctxKey string

const userCtxKey ctxKey = "username"

// ---- Casdoor 鉴权中间件 ----

// AuthMiddleware 从 Authorization header 提取 Casdoor JWT，校验后将用户名写入请求 context。
// 后续 huma handler 可通过 GetUsername(ctx) 获取。
func AuthMiddleware(verifyFn func(string) (string, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		username, err := verifyFn(auth)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.Response{
				Code:    http.StatusUnauthorized,
				Message: "unauthorized",
			})
			return
		}
		// 注入到请求 context，huma handler 的 ctx 可以获取到
		ctx := context.WithValue(c.Request.Context(), userCtxKey, username)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// GetUsername 从 huma handler 的 context 中提取当前用户名。
func GetUsername(ctx context.Context) (string, error) {
	u, ok := ctx.Value(userCtxKey).(string)
	if !ok || u == "" {
		return "", errors.New("unauthenticated")
	}
	return u, nil
}
