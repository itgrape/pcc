package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/pkg/types"
)

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

	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, resOK(token))
}
