package httpapi

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/pkg/types"
)

type UserInfoInput struct{}

type UserInfoOutput struct {
	Body types.Response
}

func (h *handlers) getUserInfo(ctx context.Context, in *UserInfoInput) (*UserInfoOutput, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return nil, toHTTPError(domain.ErrUnauthenticated)
	}
	return &UserInfoOutput{Body: resOK(user)}, nil
}

type oauthCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

func (h *handlers) exchangeOAuthCode(c *gin.Context) {
	var req oauthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Code:    http.StatusBadRequest,
			Message: "authorization code is required",
		})
		return
	}

	token, err := h.deps.Auth.ExchangeCode(req.Code, req.State)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.Response{
			Code:    http.StatusUnauthorized,
			Message: "authorization code exchange failed",
		})
		return
	}

	c.JSON(http.StatusOK, resOK(token))
}
