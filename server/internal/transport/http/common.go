package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
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

const userCtxKey ctxKey = "user"

// ---- Casdoor 鉴权中间件 ----

// AuthMiddleware 从 Authorization header 提取 Casdoor JWT，校验后将用户写入请求 context。
func AuthMiddleware(verifyFn func(string) (*casdoorsdk.User, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		user, err := verifyFn(auth)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.Response{
				Code:    http.StatusUnauthorized,
				Message: "unauthorized",
			})
			return
		}
		// 注入到请求 context，huma handler 的 ctx 可以获取到
		ctx := context.WithValue(c.Request.Context(), userCtxKey, user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetUser(ctx context.Context) (*casdoorsdk.User, error) {
	user, ok := ctx.Value(userCtxKey).(*casdoorsdk.User)
	if !ok || user == nil || user.Name == "" {
		return nil, errors.New("unauthenticated")
	}
	return user, nil
}

// GetUsername 从 huma handler 的 context 中提取当前用户名。
func GetUsername(ctx context.Context) (string, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}
