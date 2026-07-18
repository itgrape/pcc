package app

import (
	"context"
	"errors"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/internal/repository"
)

// NamespaceUsecase 命名空间用例。
type NamespaceUsecase struct {
	nsRepo  repository.NamespaceRepository
	cfgRepo repository.ConfigRepository
	aknRepo repository.AccessKeyNamespaceRepository
}

func NewNamespaceUsecase(nsRepo repository.NamespaceRepository, cfgRepo repository.ConfigRepository, aknRepo repository.AccessKeyNamespaceRepository) *NamespaceUsecase {
	return &NamespaceUsecase{nsRepo: nsRepo, cfgRepo: cfgRepo, aknRepo: aknRepo}
}

func (uc *NamespaceUsecase) List(ctx context.Context, page domain.PageReq) ([]*model.Namespace, int64, error) {
	offset := (page.Page - 1) * page.PageSize
	list, total, err := uc.nsRepo.List(ctx, offset, page.PageSize)
	return list, total, domain.MapError(err)
}

func (uc *NamespaceUsecase) GetByID(ctx context.Context, id uint64) (*model.Namespace, error) {
	ns, err := uc.nsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNamespaceNotFound
		}
		return nil, domain.MapError(err)
	}
	return ns, nil
}

func (uc *NamespaceUsecase) Create(ctx context.Context, namespace, description string) (*model.Namespace, error) {
	ns := &model.Namespace{Namespace: namespace, Description: description}
	if err := uc.nsRepo.Create(ctx, ns); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, domain.ErrDuplicate
		}
		return nil, domain.MapError(err)
	}
	return ns, nil
}

func (uc *NamespaceUsecase) Update(ctx context.Context, id uint64, description string) (*model.Namespace, error) {
	ns, err := uc.nsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNamespaceNotFound
		}
		return nil, domain.MapError(err)
	}
	ns.Description = description
	if err := uc.nsRepo.Update(ctx, ns); err != nil {
		return nil, domain.MapError(err)
	}
	return ns, nil
}

func (uc *NamespaceUsecase) Delete(ctx context.Context, id uint64) error {
	// 先获取命名空间及其版本号
	ns, err := uc.nsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNamespaceNotFound
		}
		return domain.MapError(err)
	}
	// 检查命名空间下是否有配置
	cfgs, err := uc.cfgRepo.ListByNamespaceID(ctx, id)
	if err != nil {
		return domain.MapError(err)
	}
	if len(cfgs) > 0 {
		return domain.ErrNamespaceNotEmpty
	}
	// 检查是否有授权关系
	aks, err := uc.aknRepo.ListByNamespaceID(ctx, id)
	if err != nil {
		return domain.MapError(err)
	}
	if len(aks) > 0 {
		return domain.ErrNamespaceNotEmpty
	}

	if err := uc.nsRepo.SoftDelete(ctx, id, ns.Version); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNamespaceNotFound
		}
		return domain.MapError(err)
	}
	return nil
}
