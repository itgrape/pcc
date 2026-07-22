package httpapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/pkg/types"
)

type UserInfoInput struct{}

type UserInfoOutput struct {
	Body types.Response
}

func (h *handlers) registerUserRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-user-info",
		Method:      http.MethodGet,
		Path:        adminBasePath + "/user/info",
		Summary:     "获取当前用户信息",
		Tags:        []string{"user"},
	}, h.getUserInfo)
}

func (h *handlers) getUserInfo(ctx context.Context, in *UserInfoInput) (*UserInfoOutput, error) {
	identity, err := GetIdentity(ctx)
	if err != nil {
		return nil, toHTTPError(domain.ErrUnauthenticated)
	}
	return &UserInfoOutput{Body: resOK(types.UserInfo{
		Owner:       identity.Owner,
		Username:    identity.Username,
		DisplayName: identity.DisplayName,
		Email:       identity.Email,
		Phone:       identity.Phone,
		Avatar:      identity.Avatar,
		Roles:       identity.Roles,
		Groups:      identity.Groups,
		Properties:  identity.Properties,
	})}, nil
}
