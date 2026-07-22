package httpapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/pkg/convert"
	"github.com/itgrape/pcc/server/pkg/types"
)

// ---- 密钥列表 ----

func (h *handlers) registerAccessKeyRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-access-keys", Method: http.MethodGet, Path: adminBasePath + "/access-keys",
		Summary: "密钥列表", Tags: []string{"access-keys"},
	}, h.listAccessKeys)
	huma.Register(api, huma.Operation{
		OperationID: "get-access-key", Method: http.MethodGet, Path: adminBasePath + "/access-keys/{key_id}",
		Summary: "密钥详情", Tags: []string{"access-keys"},
	}, h.getAccessKey)
	huma.Register(api, huma.Operation{
		OperationID: "create-access-key", Method: http.MethodPost, Path: adminBasePath + "/access-keys",
		Summary: "创建密钥", Tags: []string{"access-keys"}, DefaultStatus: http.StatusCreated,
	}, h.createAccessKey)
	huma.Register(api, huma.Operation{
		OperationID: "update-access-key", Method: http.MethodPut, Path: adminBasePath + "/access-keys/{key_id}",
		Summary: "更新密钥", Tags: []string{"access-keys"},
	}, h.updateAccessKey)
	huma.Register(api, huma.Operation{
		OperationID: "delete-access-key", Method: http.MethodDelete, Path: adminBasePath + "/access-keys/{key_id}",
		Summary: "删除密钥", Tags: []string{"access-keys"},
	}, h.deleteAccessKey)
	huma.Register(api, huma.Operation{
		OperationID: "list-key-namespaces", Method: http.MethodGet, Path: adminBasePath + "/access-keys/{key_id}/namespaces",
		Summary: "密钥已授权的命名空间", Tags: []string{"access-keys"},
	}, h.listKeyNamespaces)
	huma.Register(api, huma.Operation{
		OperationID: "grant-key-namespace", Method: http.MethodPost, Path: adminBasePath + "/access-keys/{key_id}/namespaces",
		Summary: "新增授权", Tags: []string{"access-keys"}, DefaultStatus: http.StatusCreated,
	}, h.grantKeyNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "update-key-namespace-permission", Method: http.MethodPut, Path: adminBasePath + "/access-keys/{key_id}/namespaces/{namespace_id}",
		Summary: "更新授权", Tags: []string{"access-keys"},
	}, h.updateKeyNamespacePermission)
	huma.Register(api, huma.Operation{
		OperationID: "revoke-key-namespace", Method: http.MethodDelete, Path: adminBasePath + "/access-keys/{key_id}/namespaces/{namespace_id}",
		Summary: "解除授权", Tags: []string{"access-keys"},
	}, h.revokeKeyNamespace)
	huma.Register(api, huma.Operation{
		OperationID: "list-namespace-access-keys", Method: http.MethodGet, Path: adminBasePath + "/namespaces/{namespace_id}/access-keys",
		Summary: "命名空间下的授权密钥", Tags: []string{"namespaces"},
	}, h.listNamespaceAccessKeys)
}

type AccessKeyListInput struct {
	Page     int    `query:"page" default:"1" minimum:"1"`
	PageSize int    `query:"page_size" default:"20" maximum:"100"`
	Keyword  string `query:"keyword,omitempty"`
	Status   int    `query:"status,omitempty"`
}

type AccessKeyListOutput struct {
	Body types.Response
}

func (h *handlers) listAccessKeys(ctx context.Context, in *AccessKeyListInput) (*AccessKeyListOutput, error) {
	list, total, err := h.deps.AccessKeys.List(ctx, domain.PageReq{Page: in.Page, PageSize: in.PageSize, Keyword: in.Keyword}, in.Status)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &AccessKeyListOutput{Body: resPage(convert.AccessKeyListToDTO(list), total, in.Page, in.PageSize)}, nil
}

// ---- 密钥详情 ----

type AccessKeyInput struct {
	KeyID uint64 `path:"key_id"`
}

type AccessKeyOutput struct {
	Body types.Response
}

func (h *handlers) getAccessKey(ctx context.Context, in *AccessKeyInput) (*AccessKeyOutput, error) {
	ak, err := h.deps.AccessKeys.GetByID(ctx, in.KeyID)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &AccessKeyOutput{Body: resOK(convert.AccessKeyToDTO(ak))}, nil
}

// ---- 创建密钥 ----

type CreateAccessKeyInput struct {
	Body types.CreateAccessKeyReq
}

type CreateAccessKeyOutput struct {
	Body types.Response
}

func (h *handlers) createAccessKey(ctx context.Context, in *CreateAccessKeyInput) (*CreateAccessKeyOutput, error) {
	created, err := h.deps.AccessKeys.Create(ctx, in.Body.Description)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &CreateAccessKeyOutput{Body: resOK(convert.AccessKeyCreatedToDTO(created))}, nil
}

// ---- 更新密钥 ----

type UpdateAccessKeyInput struct {
	KeyID uint64 `path:"key_id"`
	Body  types.UpdateAccessKeyReq
}

func (h *handlers) updateAccessKey(ctx context.Context, in *UpdateAccessKeyInput) (*AccessKeyOutput, error) {
	ak, err := h.deps.AccessKeys.Update(ctx, in.KeyID, in.Body.Description, in.Body.Status)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &AccessKeyOutput{Body: resOK(convert.AccessKeyToDTO(ak))}, nil
}

// ---- 删除密钥 ----

type AccessKeyDeleteInput struct {
	KeyID uint64 `path:"key_id"`
}

func (h *handlers) deleteAccessKey(ctx context.Context, in *AccessKeyDeleteInput) (*struct{}, error) {
	if err := h.deps.AccessKeys.Delete(ctx, in.KeyID); err != nil {
		return nil, toHTTPError(err)
	}
	return nil, nil
}

// ---- 密钥授权列表 ----

type KeyNamespaceListOutput struct {
	Body types.Response
}

func (h *handlers) listKeyNamespaces(ctx context.Context, in *AccessKeyInput) (*KeyNamespaceListOutput, error) {
	list, err := h.deps.AccessKeys.ListNamespaces(ctx, in.KeyID)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &KeyNamespaceListOutput{Body: resOK(convert.AKNamespaceListToDTO(list))}, nil
}

// ---- 新增授权 ----

type GrantNamespaceInput struct {
	KeyID uint64 `path:"key_id"`
	Body  types.GrantNamespaceReq
}

type GrantNamespaceOutput struct {
	Body types.Response
}

func (h *handlers) grantKeyNamespace(ctx context.Context, in *GrantNamespaceInput) (*GrantNamespaceOutput, error) {
	namespaceID, err := strconv.ParseUint(in.Body.NamespaceID, 10, 64)
	if err != nil {
		return nil, toHTTPError(domain.ErrValidation)
	}
	akn, err := h.deps.AccessKeys.GrantNamespace(ctx, in.KeyID, namespaceID, model.Permission(in.Body.Permission))
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &GrantNamespaceOutput{Body: resOK(convert.AKNamespaceToDTO(akn))}, nil
}

// ---- 更新授权 ----

type UpdatePermissionInput struct {
	KeyID       uint64 `path:"key_id"`
	NamespaceID uint64 `path:"namespace_id"`
	Body        types.UpdatePermissionReq
}

func (h *handlers) updateKeyNamespacePermission(ctx context.Context, in *UpdatePermissionInput) (*GrantNamespaceOutput, error) {
	akn, err := h.deps.AccessKeys.UpdatePermission(ctx, in.KeyID, in.NamespaceID, model.Permission(in.Body.Permission))
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &GrantNamespaceOutput{Body: resOK(convert.AKNamespaceToDTO(akn))}, nil
}

// ---- 解除授权 ----

type RevokeInput struct {
	KeyID       uint64 `path:"key_id"`
	NamespaceID uint64 `path:"namespace_id"`
}

func (h *handlers) revokeKeyNamespace(ctx context.Context, in *RevokeInput) (*struct{}, error) {
	if err := h.deps.AccessKeys.RevokeNamespace(ctx, in.KeyID, in.NamespaceID); err != nil {
		return nil, toHTTPError(err)
	}
	return nil, nil
}

// ---- 命名空间下的授权密钥 ----

type NamespaceAccessKeysInput struct {
	NamespaceID uint64 `path:"namespace_id"`
}

func (h *handlers) listNamespaceAccessKeys(ctx context.Context, in *NamespaceAccessKeysInput) (*KeyNamespaceListOutput, error) {
	list, err := h.deps.AccessKeys.ListByNamespace(ctx, in.NamespaceID)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &KeyNamespaceListOutput{Body: resOK(convert.AKNamespaceListToDTO(list))}, nil
}
