package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/pkg/types"
)

type contextKey string

const userContextKey contextKey = "authenticated-user"

func AuthMiddleware(authenticate func(string) (*domain.Identity, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := authenticate(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.Response{
				Code:    http.StatusUnauthorized,
				Message: "unauthorized",
			})
			return
		}

		ctx := context.WithValue(c.Request.Context(), userContextKey, identity)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetIdentity(ctx context.Context) (*domain.Identity, error) {
	identity, ok := ctx.Value(userContextKey).(*domain.Identity)
	if !ok || identity == nil || identity.Username == "" {
		return nil, errors.New("unauthenticated")
	}
	return identity, nil
}

func GetUsername(ctx context.Context) (string, error) {
	identity, err := GetIdentity(ctx)
	if err != nil {
		return "", err
	}
	return identity.Username, nil
}
