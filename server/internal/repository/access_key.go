package repository

import (
	"context"

	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/internal/query"
)

// AccessKeyRepository 访问密钥数据访问。
type AccessKeyRepository interface {
	Create(ctx context.Context, ak *model.AccessKey) error
	GetByID(ctx context.Context, id uint64) (*model.AccessKey, error)
	GetByAccessKey(ctx context.Context, ak string) (*model.AccessKey, error)
	List(ctx context.Context, offset, limit int) ([]*model.AccessKey, int64, error)
	Update(ctx context.Context, ak *model.AccessKey) error
	UpdateStatus(ctx context.Context, id uint64, version uint, status model.AccessKeyStatus) error
	SoftDelete(ctx context.Context, id uint64, version uint) error
}

type accessKeyRepository struct {
	q *query.Query
}

func (r *accessKeyRepository) Create(ctx context.Context, ak *model.AccessKey) error {
	return mapError(r.q.AccessKey.WithContext(ctx).Create(ak))
}

func (r *accessKeyRepository) GetByID(ctx context.Context, id uint64) (*model.AccessKey, error) {
	a := r.q.AccessKey
	m, err := a.WithContext(ctx).Where(a.ID.Eq(id)).First()
	return m, mapError(err)
}

func (r *accessKeyRepository) GetByAccessKey(ctx context.Context, ak string) (*model.AccessKey, error) {
	a := r.q.AccessKey
	m, err := a.WithContext(ctx).Where(a.AccessKey.Eq(ak)).First()
	return m, mapError(err)
}

func (r *accessKeyRepository) List(ctx context.Context, offset, limit int) ([]*model.AccessKey, int64, error) {
	a := r.q.AccessKey
	list, total, err := a.WithContext(ctx).Order(a.CreatedAt.Desc()).FindByPage(offset, limit)
	return list, total, mapError(err)
}

func (r *accessKeyRepository) Update(ctx context.Context, ak *model.AccessKey) error {
	return mapError(r.q.AccessKey.WithContext(ctx).Save(ak))
}

func (r *accessKeyRepository) UpdateStatus(ctx context.Context, id uint64, version uint, status model.AccessKeyStatus) error {
	a := r.q.AccessKey
	res, err := a.WithContext(ctx).
		Where(a.ID.Eq(id), a.Version.Eq(version)).
		UpdateSimple(a.Status.Value(int8(status)), a.Version.Value(version+1))
	if err != nil {
		return mapError(err)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}

func (r *accessKeyRepository) SoftDelete(ctx context.Context, id uint64, version uint) error {
	a := r.q.AccessKey
	res, err := a.WithContext(ctx).Where(a.ID.Eq(id), a.Version.Eq(version)).Delete()
	if err != nil {
		return mapError(err)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}

// AccessKeyNamespaceRepository 密钥-命名空间关联数据访问。
type AccessKeyNamespaceRepository interface {
	Create(ctx context.Context, akn *model.AccessKeyNamespace) error
	GetByID(ctx context.Context, id uint64) (*model.AccessKeyNamespace, error)
	GetByAccessKeyAndNamespace(ctx context.Context, akID, nsID uint64) (*model.AccessKeyNamespace, error)
	ListByAccessKeyID(ctx context.Context, akID uint64) ([]*model.AccessKeyNamespace, error)
	ListByNamespaceID(ctx context.Context, nsID uint64) ([]*model.AccessKeyNamespace, error)
	UpdatePermission(ctx context.Context, id uint64, version uint, perm model.Permission) error
	SoftDelete(ctx context.Context, id uint64, version uint) error
}

type accessKeyNamespaceRepository struct {
	q *query.Query
}

func (r *accessKeyNamespaceRepository) Create(ctx context.Context, akn *model.AccessKeyNamespace) error {
	a := r.q.AccessKeyNamespace
	existing, err := a.WithContext(ctx).
		Unscoped().
		Where(a.AccessKeyID.Eq(akn.AccessKeyID), a.NamespaceID.Eq(akn.NamespaceID)).
		First()
	if err == nil {
		if !existing.DeletedAt.Valid {
			return ErrDuplicate
		}
		_, err = a.WithContext(ctx).
			Unscoped().
			Where(a.ID.Eq(existing.ID)).
			Updates(map[string]any{
				"deleted_at": nil,
				"permission": akn.Permission,
				"version":    existing.Version + 1,
			})
		if err != nil {
			return mapError(err)
		}
		akn.Base = existing.Base
		akn.DeletedAt.Valid = false
		akn.Version = existing.Version + 1
		return nil
	}
	if mapped := mapError(err); mapped != ErrNotFound {
		return mapped
	}
	return mapError(a.WithContext(ctx).Create(akn))
}

func (r *accessKeyNamespaceRepository) GetByID(ctx context.Context, id uint64) (*model.AccessKeyNamespace, error) {
	a := r.q.AccessKeyNamespace
	m, err := a.WithContext(ctx).Where(a.ID.Eq(id)).First()
	return m, mapError(err)
}

func (r *accessKeyNamespaceRepository) GetByAccessKeyAndNamespace(ctx context.Context, akID, nsID uint64) (*model.AccessKeyNamespace, error) {
	a := r.q.AccessKeyNamespace
	m, err := a.WithContext(ctx).Where(a.AccessKeyID.Eq(akID), a.NamespaceID.Eq(nsID)).First()
	return m, mapError(err)
}

func (r *accessKeyNamespaceRepository) ListByAccessKeyID(ctx context.Context, akID uint64) ([]*model.AccessKeyNamespace, error) {
	a := r.q.AccessKeyNamespace
	list, err := a.WithContext(ctx).Where(a.AccessKeyID.Eq(akID)).Find()
	return list, mapError(err)
}

func (r *accessKeyNamespaceRepository) ListByNamespaceID(ctx context.Context, nsID uint64) ([]*model.AccessKeyNamespace, error) {
	a := r.q.AccessKeyNamespace
	list, err := a.WithContext(ctx).Where(a.NamespaceID.Eq(nsID)).Find()
	return list, mapError(err)
}

func (r *accessKeyNamespaceRepository) UpdatePermission(ctx context.Context, id uint64, version uint, perm model.Permission) error {
	a := r.q.AccessKeyNamespace
	res, err := a.WithContext(ctx).
		Where(a.ID.Eq(id), a.Version.Eq(version)).
		UpdateSimple(a.Permission.Value(int8(perm)), a.Version.Value(version+1))
	if err != nil {
		return mapError(err)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}

func (r *accessKeyNamespaceRepository) SoftDelete(ctx context.Context, id uint64, version uint) error {
	a := r.q.AccessKeyNamespace
	res, err := a.WithContext(ctx).Where(a.ID.Eq(id), a.Version.Eq(version)).Delete()
	if err != nil {
		return mapError(err)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}
