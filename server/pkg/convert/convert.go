// Package convert 实现内部 model ↔ 对外 types 的转换。
package convert

import (
	"strconv"

	"github.com/itgrape/pcc/server/internal/app"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/pkg/types"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"

func idToString(id uint64) string {
	return strconv.FormatUint(id, 10)
}

// ---- 命名空间 ----

func NamespaceToDTO(ns *model.Namespace) types.Namespace {
	return types.Namespace{
		ID:          idToString(ns.ID),
		Namespace:   ns.Namespace,
		Description: ns.Description,
		CreatedAt:   ns.CreatedAt.Format(timeLayout),
		UpdatedAt:   ns.UpdatedAt.Format(timeLayout),
	}
}

func NamespaceListToDTO(list []*model.Namespace) []types.Namespace {
	out := make([]types.Namespace, 0, len(list))
	for _, ns := range list {
		out = append(out, NamespaceToDTO(ns))
	}
	return out
}

// ---- 配置 ----

func ConfigToDTO(cfg *model.Config) types.Config {
	return types.Config{
		ID:          idToString(cfg.ID),
		NamespaceID: idToString(cfg.NamespaceID),
		ConfigName:  cfg.ConfigName,
		Format:      int8(cfg.Format),
		Content:     cfg.Content,
		Description: cfg.Description,
		Version:     cfg.Version,
		CreatedBy:   cfg.CreatedBy,
		UpdatedBy:   cfg.UpdatedBy,
		CreatedAt:   cfg.CreatedAt.Format(timeLayout),
		UpdatedAt:   cfg.UpdatedAt.Format(timeLayout),
	}
}

func ConfigListToDTO(list []*model.Config) []types.Config {
	out := make([]types.Config, 0, len(list))
	for _, cfg := range list {
		dto := ConfigToDTO(cfg)
		dto.Content = ""
		out = append(out, dto)
	}
	return out
}

// ---- 配置历史 ----

func ConfigHistoryToDTO(h *model.ConfigHistory) types.ConfigHistory {
	return types.ConfigHistory{
		ID:            idToString(h.ID),
		ConfigVersion: h.ConfigVersion,
		Operator:      h.Operator,
		ChangeReason:  h.ChangeReason,
		Content:       h.Content,
		CreatedAt:     h.CreatedAt.Format(timeLayout),
	}
}

func ConfigHistoryListToDTO(list []*model.ConfigHistory) []types.ConfigHistory {
	out := make([]types.ConfigHistory, 0, len(list))
	for _, h := range list {
		dto := ConfigHistoryToDTO(h)
		dto.Content = ""
		out = append(out, dto)
	}
	return out
}

// ---- 访问密钥 ----

func AccessKeyToDTO(ak *model.AccessKey) types.AccessKey {
	return types.AccessKey{
		ID:          idToString(ak.ID),
		AccessKey:   ak.AccessKey,
		Status:      int(ak.Status),
		Description: ak.Description,
		CreatedAt:   ak.CreatedAt.Format(timeLayout),
		UpdatedAt:   ak.UpdatedAt.Format(timeLayout),
	}
}

func AccessKeyListToDTO(list []*model.AccessKey) []types.AccessKey {
	out := make([]types.AccessKey, 0, len(list))
	for _, ak := range list {
		out = append(out, AccessKeyToDTO(ak))
	}
	return out
}

func AccessKeyCreatedToDTO(ak *app.AccessKeyCreated) types.AccessKeyCreated {
	return types.AccessKeyCreated{
		ID:          idToString(ak.ID),
		AccessKey:   ak.AccessKey.AccessKey,
		Secret:      ak.Secret,
		Description: ak.Description,
		Status:      int(ak.Status),
	}
}

// ---- 密钥-命名空间授权 ----

func AKNamespaceToDTO(akn *model.AccessKeyNamespace) types.AccessKeyNamespace {
	return types.AccessKeyNamespace{
		AccessKeyID: idToString(akn.AccessKeyID),
		NamespaceID: idToString(akn.NamespaceID),
		Permission:  int(akn.Permission),
		CreatedAt:   akn.CreatedAt.Format(timeLayout),
	}
}

func AKNamespaceListToDTO(list []*model.AccessKeyNamespace) []types.AccessKeyNamespace {
	out := make([]types.AccessKeyNamespace, 0, len(list))
	for _, akn := range list {
		out = append(out, AKNamespaceToDTO(akn))
	}
	return out
}
