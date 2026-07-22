package app

import (
	"context"
	"errors"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/internal/repository"
	"github.com/itgrape/pcc/server/pkg/idgen"
)

// ConfigUsecase 配置用例，包含配置 CRUD、历史、回滚。
type ConfigUsecase struct {
	store *repository.Store
	bus   *EventBus
}

func NewConfigUsecase(store *repository.Store, bus *EventBus) *ConfigUsecase {
	return &ConfigUsecase{store: store, bus: bus}
}

func (uc *ConfigUsecase) ListByNamespaceID(ctx context.Context, nsID uint64) ([]*model.Config, error) {
	list, err := uc.store.Config.ListByNamespaceID(ctx, nsID)
	return list, domain.MapError(err)
}

func (uc *ConfigUsecase) Count(ctx context.Context) (int64, error) {
	_, total, err := uc.store.Config.List(ctx, 0, 1)
	return total, domain.MapError(err)
}

func (uc *ConfigUsecase) GetByID(ctx context.Context, id uint64) (*model.Config, error) {
	cfg, err := uc.store.Config.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, domain.MapError(err)
	}
	return cfg, nil
}

func (uc *ConfigUsecase) Create(ctx context.Context, nsID uint64, configName string, format model.ConfigFormat, content, description, createdBy string) (*model.Config, error) {
	cfg := &model.Config{
		NamespaceID: nsID,
		ConfigName:  configName,
		Format:      format,
		Content:     content,
		Description: description,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}
	if err := uc.store.Config.Create(ctx, cfg); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, domain.ErrDuplicate
		}
		return nil, domain.MapError(err)
	}
	return cfg, nil
}

func (uc *ConfigUsecase) Update(ctx context.Context, id uint64, content, description, changeReason, updatedBy string) (*model.Config, error) {
	cfg, err := uc.store.Config.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, domain.MapError(err)
	}

	// 使用事务原子地写入历史快照并更新内容
	if err := uc.store.Tx(ctx, func(tx *repository.Store) error {
		// 写入当前版本到历史表
		if err := saveConfigHistory(ctx, tx, cfg, cfg.UpdatedBy, changeReason); err != nil {
			return err
		}

		// 更新内容
		if err := tx.Config.UpdateContent(ctx, id, cfg.Version, content, updatedBy, changeReason); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, domain.ErrConflict
		}
		return nil, domain.MapError(err)
	}

	// 重新读取
	cfg, err = uc.store.Config.GetByID(ctx, id)
	if err != nil {
		return nil, domain.MapError(err)
	}

	// 发布变更事件
	uc.bus.Publish(ConfigEvent{
		Type:        EventConfigUpdated,
		NamespaceID: cfg.NamespaceID,
		ConfigID:    cfg.ID,
		ConfigName:  cfg.ConfigName,
		Format:      cfg.Format,
		Content:     cfg.Content,
		Version:     cfg.Version,
	})

	return cfg, nil
}

func (uc *ConfigUsecase) Delete(ctx context.Context, id uint64) error {
	cfg, err := uc.store.Config.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrConfigNotFound
		}
		return domain.MapError(err)
	}

	if err := uc.store.Config.SoftDelete(ctx, id, cfg.Version); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return domain.ErrConflict
		}
		return domain.MapError(err)
	}

	uc.bus.Publish(ConfigEvent{
		Type:        EventConfigDeleted,
		NamespaceID: cfg.NamespaceID,
		ConfigID:    cfg.ID,
		ConfigName:  cfg.ConfigName,
	})

	return nil
}

// ---- 历史 ----

func (uc *ConfigUsecase) ListHistory(ctx context.Context, configID uint64) ([]*model.ConfigHistory, error) {
	list, err := uc.store.ConfigHistory.ListByConfigID(ctx, configID)
	return list, domain.MapError(err)
}

func (uc *ConfigUsecase) GetHistoryByID(ctx context.Context, id uint64) (*model.ConfigHistory, error) {
	h, err := uc.store.ConfigHistory.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrHistoryNotFound
		}
		return nil, domain.MapError(err)
	}
	return h, nil
}

// ---- 回滚 ----

func (uc *ConfigUsecase) Rollback(ctx context.Context, configID, historyID uint64, changeReason, operator string) (*model.Config, error) {
	// 读取目标历史
	history, err := uc.store.ConfigHistory.GetByID(ctx, historyID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrHistoryNotFound
		}
		return nil, domain.MapError(err)
	}

	cfg, err := uc.store.Config.GetByID(ctx, configID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, domain.MapError(err)
	}

	// 事务：写入当前快照 + 覆盖为历史内容
	if err := uc.store.Tx(ctx, func(tx *repository.Store) error {
		if err := saveConfigHistory(ctx, tx, cfg, operator, "rollback before: "+changeReason); err != nil {
			return err
		}
		if err := tx.Config.UpdateContent(ctx, configID, cfg.Version, history.Content, operator, changeReason); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, domain.ErrConflict
		}
		return nil, domain.MapError(err)
	}

	cfg, err = uc.store.Config.GetByID(ctx, configID)
	if err != nil {
		return nil, domain.MapError(err)
	}

	uc.bus.Publish(ConfigEvent{
		Type:        EventConfigUpdated,
		NamespaceID: cfg.NamespaceID,
		ConfigID:    cfg.ID,
		ConfigName:  cfg.ConfigName,
		Format:      cfg.Format,
		Content:     cfg.Content,
		Version:     cfg.Version,
	})

	return cfg, nil
}

// saveConfigHistory 保存一条配置历史快照。
func saveConfigHistory(ctx context.Context, tx *repository.Store, cfg *model.Config, operator, reason string) error {
	id, err := idgen.NextID()
	if err != nil {
		return err
	}
	h := &model.ConfigHistory{
		ID:            id,
		ConfigID:      cfg.ID,
		NamespaceID:   cfg.NamespaceID,
		ConfigName:    cfg.ConfigName,
		Content:       cfg.Content,
		ConfigVersion: cfg.Version,
		Operator:      operator,
		ChangeReason:  reason,
	}
	return tx.ConfigHistory.Create(ctx, h)
}
