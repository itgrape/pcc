package repository

import (
	"context"

	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/internal/query"
)

// NamespaceRepository 命名空间数据访问。
type NamespaceRepository interface {
	Create(ctx context.Context, ns *model.Namespace) error
	GetByID(ctx context.Context, id uint64) (*model.Namespace, error)
	GetByNamespace(ctx context.Context, ns string) (*model.Namespace, error)
	List(ctx context.Context, offset, limit int) ([]*model.Namespace, int64, error)
	Update(ctx context.Context, ns *model.Namespace) error
	SoftDelete(ctx context.Context, id uint64, version uint) error
}

type namespaceRepository struct {
	q *query.Query
}

func (r *namespaceRepository) Create(ctx context.Context, ns *model.Namespace) error {
	return mapError(r.q.Namespace.WithContext(ctx).Create(ns))
}

func (r *namespaceRepository) GetByID(ctx context.Context, id uint64) (*model.Namespace, error) {
	n := r.q.Namespace
	m, err := n.WithContext(ctx).Where(n.ID.Eq(id)).First()
	return m, mapError(err)
}

func (r *namespaceRepository) GetByNamespace(ctx context.Context, ns string) (*model.Namespace, error) {
	n := r.q.Namespace
	m, err := n.WithContext(ctx).Where(n.Namespace.Eq(ns)).First()
	return m, mapError(err)
}

func (r *namespaceRepository) List(ctx context.Context, offset, limit int) ([]*model.Namespace, int64, error) {
	n := r.q.Namespace
	list, total, err := n.WithContext(ctx).Order(n.CreatedAt.Desc()).FindByPage(offset, limit)
	return list, total, mapError(err)
}

func (r *namespaceRepository) Update(ctx context.Context, ns *model.Namespace) error {
	return mapError(r.q.Namespace.WithContext(ctx).Save(ns))
}

func (r *namespaceRepository) SoftDelete(ctx context.Context, id uint64, version uint) error {
	n := r.q.Namespace
	res, err := n.WithContext(ctx).Where(n.ID.Eq(id), n.Version.Eq(version)).Delete()
	if err != nil {
		return mapError(err)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}
