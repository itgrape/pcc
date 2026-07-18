package repository

import (
	"context"

	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/internal/query"
	"github.com/itgrape/pcc/server/pkg/idgen"
	"gorm.io/gen/field"
)

// ConfigUpdate 配置更新参数；nil 表示不改动。
type ConfigUpdate struct {
	Content     *string
	Description *string
}

// ConfigRepository 配置主表数据访问。
type ConfigRepository interface {
	Create(ctx context.Context, cfg *model.Config) error
	GetByID(ctx context.Context, id uint64) (*model.Config, error)
	GetByNamespaceAndName(ctx context.Context, nsID uint64, name string) (*model.Config, error)
	ListByNamespaceID(ctx context.Context, nsID uint64) ([]*model.Config, error)
	List(ctx context.Context, offset, limit int) ([]*model.Config, int64, error)
	UpdateContent(ctx context.Context, id uint64, version uint, content, updatedBy, changeReason string) error
	UpdateConfig(ctx context.Context, id uint64, version uint, upd ConfigUpdate) error
	SoftDelete(ctx context.Context, id uint64, version uint) error
}

type configRepository struct {
	q *query.Query
}

func (r *configRepository) Create(ctx context.Context, cfg *model.Config) error {
	return mapError(r.q.Config.WithContext(ctx).Create(cfg))
}

func (r *configRepository) GetByID(ctx context.Context, id uint64) (*model.Config, error) {
	c := r.q.Config
	m, err := c.WithContext(ctx).Where(c.ID.Eq(id)).First()
	return m, mapError(err)
}

func (r *configRepository) GetByNamespaceAndName(ctx context.Context, nsID uint64, name string) (*model.Config, error) {
	c := r.q.Config
	m, err := c.WithContext(ctx).Where(c.NamespaceID.Eq(nsID), c.ConfigName.Eq(name)).First()
	return m, mapError(err)
}

func (r *configRepository) ListByNamespaceID(ctx context.Context, nsID uint64) ([]*model.Config, error) {
	c := r.q.Config
	list, err := c.WithContext(ctx).Where(c.NamespaceID.Eq(nsID)).Order(c.ConfigName).Find()
	return list, mapError(err)
}

func (r *configRepository) List(ctx context.Context, offset, limit int) ([]*model.Config, int64, error) {
	c := r.q.Config
	list, total, err := c.WithContext(ctx).Order(c.CreatedAt.Desc()).FindByPage(offset, limit)
	return list, total, mapError(err)
}

// UpdateContent 原子更新配置内容，自增 version，并写入历史快照。
func (r *configRepository) UpdateContent(ctx context.Context, id uint64, version uint, content, updatedBy, changeReason string) error {
	c := r.q.Config
	assigns := []field.AssignExpr{
		c.Content.Value(content),
		c.UpdatedBy.Value(updatedBy),
		c.Version.Value(version + 1),
	}
	res, err := c.WithContext(ctx).Where(c.ID.Eq(id), c.Version.Eq(version)).UpdateSimple(assigns...)
	if err != nil {
		return mapError(err)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}

func (r *configRepository) UpdateConfig(ctx context.Context, id uint64, version uint, upd ConfigUpdate) error {
	c := r.q.Config
	assigns := make([]field.AssignExpr, 0, 3)
	if upd.Content != nil {
		assigns = append(assigns, c.Content.Value(*upd.Content))
	}
	if upd.Description != nil {
		assigns = append(assigns, c.Description.Value(*upd.Description))
	}
	if len(assigns) == 0 {
		return nil
	}
	assigns = append(assigns, c.Version.Value(version+1))
	res, err := c.WithContext(ctx).Where(c.ID.Eq(id), c.Version.Eq(version)).UpdateSimple(assigns...)
	if err != nil {
		return mapError(err)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}

func (r *configRepository) SoftDelete(ctx context.Context, id uint64, version uint) error {
	c := r.q.Config
	res, err := c.WithContext(ctx).Where(c.ID.Eq(id), c.Version.Eq(version)).Delete()
	if err != nil {
		return mapError(err)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}

// ConfigHistoryRepository 配置历史数据访问（追加式）。
type ConfigHistoryRepository interface {
	Create(ctx context.Context, h *model.ConfigHistory) error
	GetByID(ctx context.Context, id uint64) (*model.ConfigHistory, error)
	ListByConfigID(ctx context.Context, configID uint64) ([]*model.ConfigHistory, error)
	LatestByConfigID(ctx context.Context, configID uint64) (*model.ConfigHistory, error)
}

type configHistoryRepository struct {
	q *query.Query
}

func (r *configHistoryRepository) Create(ctx context.Context, h *model.ConfigHistory) error {
	return mapError(r.q.ConfigHistory.WithContext(ctx).Create(h))
}

func (r *configHistoryRepository) GetByID(ctx context.Context, id uint64) (*model.ConfigHistory, error) {
	ch := r.q.ConfigHistory
	m, err := ch.WithContext(ctx).Where(ch.ID.Eq(id)).First()
	return m, mapError(err)
}

func (r *configHistoryRepository) ListByConfigID(ctx context.Context, configID uint64) ([]*model.ConfigHistory, error) {
	ch := r.q.ConfigHistory
	list, err := ch.WithContext(ctx).Where(ch.ConfigID.Eq(configID)).Order(ch.ConfigVersion.Desc()).Find()
	return list, mapError(err)
}

func (r *configHistoryRepository) LatestByConfigID(ctx context.Context, configID uint64) (*model.ConfigHistory, error) {
	ch := r.q.ConfigHistory
	m, err := ch.WithContext(ctx).Where(ch.ConfigID.Eq(configID)).Order(ch.ConfigVersion.Desc()).First()
	return m, mapError(err)
}

// NewConfigHistory 构造并保存一条历史快照。
func NewConfigHistory(ctx context.Context, q *query.Query, cfg *model.Config, operator, reason string) error {
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
	return q.ConfigHistory.WithContext(ctx).Create(h)
}
