package httpapi

import (
	"context"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/pkg/types"
)

type UserInfoInput struct{}

type UserInfoOutput struct {
	Body types.Response
}

func (h *handlers) getUserInfo(ctx context.Context, in *UserInfoInput) (*UserInfoOutput, error) {
	username, err := GetUsername(ctx)
	if err != nil {
		return nil, toHTTPError(domain.ErrUnauthenticated)
	}
	return &UserInfoOutput{Body: resOK(types.UserInfo{Username: username})}, nil
}
