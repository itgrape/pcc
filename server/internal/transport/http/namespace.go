package httpapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/pkg/convert"
	"github.com/itgrape/pcc/server/pkg/types"
)

// ---- 命名空间 ----

func (h *handlers) registerNamespaceRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-namespaces", Method: http.MethodGet, Path: adminBasePath + "/namespaces",
		Summary: "命名空间列表", Tags: []string{"namespaces"},
	}, h.listNamespaces)
	huma.Register(api, huma.Operation{
		OperationID: "get-namespace", Method: http.MethodGet, Path: adminBasePath + "/namespaces/{namespace_id}",
		Summary: "命名空间详情", Tags: []string{"namespaces"},
	}, h.getNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "create-namespace", Method: http.MethodPost, Path: adminBasePath + "/namespaces",
		Summary: "创建命名空间", Tags: []string{"namespaces"}, DefaultStatus: http.StatusCreated,
	}, h.createNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "update-namespace", Method: http.MethodPut, Path: adminBasePath + "/namespaces/{namespace_id}",
		Summary: "更新命名空间", Tags: []string{"namespaces"},
	}, h.updateNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "delete-namespace", Method: http.MethodDelete, Path: adminBasePath + "/namespaces/{namespace_id}",
		Summary: "删除命名空间", Tags: []string{"namespaces"},
	}, h.deleteNamespace)
}

type NamespaceListInput struct {
	Page     int    `query:"page" default:"1" minimum:"1"`
	PageSize int    `query:"page_size" default:"20" maximum:"100"`
	Keyword  string `query:"keyword,omitempty"`
}

type NamespaceListOutput struct {
	Body types.Response
}

func (h *handlers) listNamespaces(ctx context.Context, in *NamespaceListInput) (*NamespaceListOutput, error) {
	list, total, err := h.deps.Namespaces.List(ctx, domain.PageReq{Page: in.Page, PageSize: in.PageSize, Keyword: in.Keyword})
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &NamespaceListOutput{Body: resPage(convert.NamespaceListToDTO(list), total, in.Page, in.PageSize)}, nil
}

type NamespaceInput struct {
	NamespaceID uint64 `path:"namespace_id"`
}

type NamespaceOutput struct {
	Body types.Response
}

func (h *handlers) getNamespace(ctx context.Context, in *NamespaceInput) (*NamespaceOutput, error) {
	ns, err := h.deps.Namespaces.GetByID(ctx, in.NamespaceID)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &NamespaceOutput{Body: resOK(convert.NamespaceToDTO(ns))}, nil
}

type CreateNamespaceInput struct {
	Body types.CreateNamespaceReq
}

func (h *handlers) createNamespace(ctx context.Context, in *CreateNamespaceInput) (*NamespaceOutput, error) {
	ns, err := h.deps.Namespaces.Create(ctx, in.Body.Namespace, in.Body.Description)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &NamespaceOutput{Body: resOK(convert.NamespaceToDTO(ns))}, nil
}

type UpdateNamespaceInput struct {
	NamespaceID uint64 `path:"namespace_id"`
	Body        types.UpdateNamespaceReq
}

func (h *handlers) updateNamespace(ctx context.Context, in *UpdateNamespaceInput) (*NamespaceOutput, error) {
	ns, err := h.deps.Namespaces.Update(ctx, in.NamespaceID, in.Body.Description)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &NamespaceOutput{Body: resOK(convert.NamespaceToDTO(ns))}, nil
}

type NamespaceDeleteInput struct {
	NamespaceID uint64 `path:"namespace_id"`
}

func (h *handlers) deleteNamespace(ctx context.Context, in *NamespaceDeleteInput) (*struct{}, error) {
	if err := h.deps.Namespaces.Delete(ctx, in.NamespaceID); err != nil {
		return nil, toHTTPError(err)
	}
	return nil, nil
}
