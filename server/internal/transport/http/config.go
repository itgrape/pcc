package httpapi

import (
	"context"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/pkg/convert"
	"github.com/itgrape/pcc/server/pkg/types"
)

// ---- 配置列表 ----

type ConfigListInput struct {
	NamespaceID uint64 `path:"namespace_id"`
}

type ConfigListOutput struct {
	Body types.Response
}

func (h *handlers) listConfigs(ctx context.Context, in *ConfigListInput) (*ConfigListOutput, error) {
	list, err := h.deps.Configs.ListByNamespaceID(ctx, in.NamespaceID)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &ConfigListOutput{Body: resOK(convert.ConfigListToDTO(list))}, nil
}

// ---- 配置详情 ----

type ConfigInput struct {
	ConfigID uint64 `path:"config_id"`
}

type ConfigOutput struct {
	Body types.Response
}

func (h *handlers) getConfig(ctx context.Context, in *ConfigInput) (*ConfigOutput, error) {
	cfg, err := h.deps.Configs.GetByID(ctx, in.ConfigID)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &ConfigOutput{Body: resOK(convert.ConfigToDTO(cfg))}, nil
}

// ---- 创建配置 ----

type CreateConfigInput struct {
	NamespaceID uint64 `path:"namespace_id"`
	Body        types.CreateConfigReq
}

func (h *handlers) createConfig(ctx context.Context, in *CreateConfigInput) (*ConfigOutput, error) {
	username, err := GetUsername(ctx)
	if err != nil {
		return nil, toHTTPError(domain.ErrUnauthenticated)
	}
	cfg, err := h.deps.Configs.Create(ctx, in.NamespaceID, in.Body.ConfigName, model.ConfigFormat(in.Body.Format), in.Body.Content, in.Body.Description, username)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &ConfigOutput{Body: resOK(convert.ConfigToDTO(cfg))}, nil
}

// ---- 更新配置 ----

type UpdateConfigInput struct {
	ConfigID uint64 `path:"config_id"`
	Body     types.UpdateConfigReq
}

func (h *handlers) updateConfig(ctx context.Context, in *UpdateConfigInput) (*ConfigOutput, error) {
	username, err := GetUsername(ctx)
	if err != nil {
		return nil, toHTTPError(domain.ErrUnauthenticated)
	}
	cfg, err := h.deps.Configs.Update(ctx, in.ConfigID, in.Body.Content, in.Body.Description, in.Body.ChangeReason, username)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &ConfigOutput{Body: resOK(convert.ConfigToDTO(cfg))}, nil
}

// ---- 删除配置 ----

type ConfigDeleteInput struct {
	ConfigID uint64 `path:"config_id"`
}

func (h *handlers) deleteConfig(ctx context.Context, in *ConfigDeleteInput) (*struct{}, error) {
	if err := h.deps.Configs.Delete(ctx, in.ConfigID); err != nil {
		return nil, toHTTPError(err)
	}
	return nil, nil
}

// ---- 历史列表 ----

type ConfigHistoryListInput struct {
	ConfigID uint64 `path:"config_id"`
}

type ConfigHistoryListOutput struct {
	Body types.Response
}

func (h *handlers) listConfigHistory(ctx context.Context, in *ConfigHistoryListInput) (*ConfigHistoryListOutput, error) {
	list, err := h.deps.Configs.ListHistory(ctx, in.ConfigID)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &ConfigHistoryListOutput{Body: resOK(convert.ConfigHistoryListToDTO(list))}, nil
}

// ---- 历史详情 ----

type ConfigHistoryInput struct {
	HistoryID uint64 `path:"history_id"`
}

type ConfigHistoryOutput struct {
	Body types.Response
}

func (h *handlers) getConfigHistory(ctx context.Context, in *ConfigHistoryInput) (*ConfigHistoryOutput, error) {
	ch, err := h.deps.Configs.GetHistoryByID(ctx, in.HistoryID)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &ConfigHistoryOutput{Body: resOK(convert.ConfigHistoryToDTO(ch))}, nil
}

// ---- 回滚 ----

type RollbackInput struct {
	ConfigID  uint64 `path:"config_id"`
	HistoryID uint64 `path:"history_id"`
	Body      types.RollbackReq
}

func (h *handlers) rollbackConfig(ctx context.Context, in *RollbackInput) (*ConfigOutput, error) {
	username, err := GetUsername(ctx)
	if err != nil {
		return nil, toHTTPError(domain.ErrUnauthenticated)
	}
	cfg, err := h.deps.Configs.Rollback(ctx, in.ConfigID, in.HistoryID, in.Body.ChangeReason, username)
	if err != nil {
		return nil, toHTTPError(err)
	}
	return &ConfigOutput{Body: resOK(convert.ConfigToDTO(cfg))}, nil
}
