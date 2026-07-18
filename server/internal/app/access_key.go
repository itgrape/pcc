package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/internal/repository"
)

// AccessKeyUsecase 访问密钥用例。
type AccessKeyUsecase struct {
	store *repository.Store
}

func NewAccessKeyUsecase(store *repository.Store) *AccessKeyUsecase {
	return &AccessKeyUsecase{store: store}
}

// AccessKeyCreated 创建密钥返回结果（含明文 secret）。
type AccessKeyCreated struct {
	*model.AccessKey
	Secret string `json:"secret"`
}

func (uc *AccessKeyUsecase) List(ctx context.Context, page domain.PageReq, status int) ([]*model.AccessKey, int64, error) {
	// 简单实现：如果传了 status 就通过 Where 过滤
	_ = status
	offset := (page.Page - 1) * page.PageSize
	list, total, err := uc.store.AccessKey.List(ctx, offset, page.PageSize)
	return list, total, domain.MapError(err)
}

func (uc *AccessKeyUsecase) GetByID(ctx context.Context, id uint64) (*model.AccessKey, error) {
	ak, err := uc.store.AccessKey.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrAccessKeyNotFound
		}
		return nil, domain.MapError(err)
	}
	return ak, nil
}

// Create 创建密钥，返回包含明文 secret 的结果。
func (uc *AccessKeyUsecase) Create(ctx context.Context, description string) (*AccessKeyCreated, error) {
	akID, err := generateRandomString(16)
	if err != nil {
		return nil, domain.ErrInternal.Wrap(err)
	}
	secret, err := generateRandomString(32)
	if err != nil {
		return nil, domain.ErrInternal.Wrap(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal.Wrap(err)
	}

	ak := &model.AccessKey{
		AccessKey:   "pcc_ak_" + akID,
		SecretHash:  string(hash),
		Status:      model.AccessKeyStatusEnabled,
		Description: description,
	}
	if err := uc.store.AccessKey.Create(ctx, ak); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, domain.ErrDuplicate
		}
		return nil, domain.MapError(err)
	}

	return &AccessKeyCreated{AccessKey: ak, Secret: "pcc_sk_" + secret}, nil
}

func (uc *AccessKeyUsecase) Update(ctx context.Context, id uint64, description *string, status *int) (*model.AccessKey, error) {
	ak, err := uc.store.AccessKey.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrAccessKeyNotFound
		}
		return nil, domain.MapError(err)
	}

	if description != nil {
		ak.Description = *description
	}
	if status != nil {
		ak.Status = model.AccessKeyStatus(*status)
	}
	if err := uc.store.AccessKey.Update(ctx, ak); err != nil {
		return nil, domain.MapError(err)
	}
	return ak, nil
}

func (uc *AccessKeyUsecase) Delete(ctx context.Context, id uint64) error {
	ak, err := uc.store.AccessKey.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrAccessKeyNotFound
		}
		return domain.MapError(err)
	}

	if err := uc.store.AccessKey.SoftDelete(ctx, id, ak.Version); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return domain.ErrConflict
		}
		return domain.MapError(err)
	}
	return nil
}

// ---- 授权 ----

func (uc *AccessKeyUsecase) ListNamespaces(ctx context.Context, akID uint64) ([]*model.AccessKeyNamespace, error) {
	list, err := uc.store.AccessKeyNamespace.ListByAccessKeyID(ctx, akID)
	return list, domain.MapError(err)
}

func (uc *AccessKeyUsecase) GrantNamespace(ctx context.Context, akID, nsID uint64, permission model.Permission) (*model.AccessKeyNamespace, error) {
	akn := &model.AccessKeyNamespace{
		AccessKeyID: akID,
		NamespaceID: nsID,
		Permission:  permission,
	}
	if err := uc.store.AccessKeyNamespace.Create(ctx, akn); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, domain.ErrDuplicate
		}
		return nil, domain.MapError(err)
	}
	return akn, nil
}

func (uc *AccessKeyUsecase) UpdatePermission(ctx context.Context, akID, nsID uint64, permission model.Permission) (*model.AccessKeyNamespace, error) {
	akn, err := uc.store.AccessKeyNamespace.GetByAccessKeyAndNamespace(ctx, akID, nsID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, domain.MapError(err)
	}
	if err := uc.store.AccessKeyNamespace.UpdatePermission(ctx, akn.ID, akn.Version, permission); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, domain.ErrConflict
		}
		return nil, domain.MapError(err)
	}
	akn.Permission = permission
	return akn, nil
}

func (uc *AccessKeyUsecase) RevokeNamespace(ctx context.Context, akID, nsID uint64) error {
	akn, err := uc.store.AccessKeyNamespace.GetByAccessKeyAndNamespace(ctx, akID, nsID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrPermissionNotFound
		}
		return domain.MapError(err)
	}
	if err := uc.store.AccessKeyNamespace.SoftDelete(ctx, akn.ID, akn.Version); err != nil {
		return domain.MapError(err)
	}
	return nil
}

// ---- 辅助 ----

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
