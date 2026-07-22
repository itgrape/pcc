package httpapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/itgrape/pcc/server/pkg/types"
)

type SysEnumsInput struct{}

type SysEnumsOutput struct {
	Body types.Response
}

func (h *handlers) registerSystemRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-system-enums",
		Method:      http.MethodGet,
		Path:        adminBasePath + "/system/enums",
		Summary:     "系统枚举值",
		Tags:        []string{"system"},
	}, h.getSystemEnums)
}

func (h *handlers) getSystemEnums(ctx context.Context, in *SysEnumsInput) (*SysEnumsOutput, error) {
	return &SysEnumsOutput{Body: resOK(types.SysEnums{
		ConfigFormats: []types.EnumItem{
			{Value: 1, Label: "json"},
			{Value: 2, Label: "yaml"},
			{Value: 3, Label: "text"},
			{Value: 4, Label: "properties"},
		},
		Permissions: []types.EnumItem{
			{Value: 1, Label: "只读"},
			{Value: 2, Label: "读写"},
		},
		KeyStatus: []types.EnumItem{
			{Value: 1, Label: "启用"},
			{Value: 0, Label: "禁用"},
		},
	})}, nil
}
