package app

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/internal/repository"
)

type runtimeAccessKeyRepository interface {
	GetByAccessKey(ctx context.Context, accessKey string) (*model.AccessKey, error)
}

type runtimeNamespaceRepository interface {
	GetByNamespace(ctx context.Context, namespace string) (*model.Namespace, error)
}

type runtimePermissionRepository interface {
	GetByAccessKeyAndNamespace(ctx context.Context, accessKeyID, namespaceID uint64) (*model.AccessKeyNamespace, error)
}

type runtimeConfigRepository interface {
	GetByNamespaceAndName(ctx context.Context, namespaceID uint64, name string) (*model.Config, error)
	ListByNamespaceID(ctx context.Context, namespaceID uint64) ([]*model.Config, error)
}

// RuntimeConfigUsecase authenticates service credentials and enforces
// namespace grants before returning configuration data.
type RuntimeConfigUsecase struct {
	accessKeys  runtimeAccessKeyRepository
	namespaces  runtimeNamespaceRepository
	permissions runtimePermissionRepository
	configs     runtimeConfigRepository
}

func NewRuntimeConfigUsecase(
	accessKeys repository.AccessKeyRepository,
	namespaces repository.NamespaceRepository,
	permissions repository.AccessKeyNamespaceRepository,
	configs repository.ConfigRepository,
) *RuntimeConfigUsecase {
	return &RuntimeConfigUsecase{
		accessKeys:  accessKeys,
		namespaces:  namespaces,
		permissions: permissions,
		configs:     configs,
	}
}

func (uc *RuntimeConfigUsecase) Get(
	ctx context.Context,
	accessKey string,
	secret string,
	namespace string,
	configName string,
) (*model.Config, error) {
	namespaceID, err := uc.authorizeNamespace(ctx, accessKey, secret, namespace)
	if err != nil {
		return nil, err
	}

	config, err := uc.configs.GetByNamespaceAndName(ctx, namespaceID, configName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, domain.MapError(err)
	}
	return config, nil
}

func (uc *RuntimeConfigUsecase) List(
	ctx context.Context,
	accessKey string,
	secret string,
	namespace string,
) ([]*model.Config, error) {
	namespaceID, err := uc.authorizeNamespace(ctx, accessKey, secret, namespace)
	if err != nil {
		return nil, err
	}

	configs, err := uc.configs.ListByNamespaceID(ctx, namespaceID)
	return configs, domain.MapError(err)
}

func (uc *RuntimeConfigUsecase) authorizeNamespace(
	ctx context.Context,
	accessKey string,
	secret string,
	namespace string,
) (uint64, error) {
	if accessKey == "" || secret == "" || namespace == "" {
		return 0, domain.ErrUnauthenticated
	}

	key, err := uc.accessKeys.GetByAccessKey(ctx, accessKey)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, domain.ErrUnauthenticated
		}
		return 0, domain.MapError(err)
	}
	if key.Status != model.AccessKeyStatusEnabled {
		return 0, domain.ErrUnauthenticated
	}

	rawSecret := strings.TrimPrefix(secret, "pcc_sk_")
	if rawSecret == secret || bcrypt.CompareHashAndPassword([]byte(key.SecretHash), []byte(rawSecret)) != nil {
		return 0, domain.ErrUnauthenticated
	}

	ns, err := uc.namespaces.GetByNamespace(ctx, namespace)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, domain.ErrForbidden
		}
		return 0, domain.MapError(err)
	}

	grant, err := uc.permissions.GetByAccessKeyAndNamespace(ctx, key.ID, ns.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, domain.ErrForbidden
		}
		return 0, domain.MapError(err)
	}
	if grant.Permission != model.PermissionReadOnly && grant.Permission != model.PermissionReadWrite {
		return 0, domain.ErrForbidden
	}

	return ns.ID, nil
}
