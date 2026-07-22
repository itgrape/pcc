package app

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/model"
	"github.com/itgrape/pcc/server/internal/repository"
)

type runtimeAccessKeyFake struct {
	key *model.AccessKey
	err error
}

func (f runtimeAccessKeyFake) GetByAccessKey(context.Context, string) (*model.AccessKey, error) {
	return f.key, f.err
}

type runtimeNamespaceFake struct {
	namespace *model.Namespace
	err       error
}

func (f runtimeNamespaceFake) GetByNamespace(context.Context, string) (*model.Namespace, error) {
	return f.namespace, f.err
}

type runtimePermissionFake struct {
	permission *model.AccessKeyNamespace
	err        error
}

func (f runtimePermissionFake) GetByAccessKeyAndNamespace(context.Context, uint64, uint64) (*model.AccessKeyNamespace, error) {
	return f.permission, f.err
}

type runtimeConfigFake struct {
	configs []*model.Config
	err     error
}

func (f runtimeConfigFake) GetByNamespaceAndName(context.Context, uint64, string) (*model.Config, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.configs[0], nil
}

func (f runtimeConfigFake) ListByNamespaceID(context.Context, uint64) ([]*model.Config, error) {
	return f.configs, f.err
}

func TestRuntimeConfigRequiresValidCredentialAndGrant(t *testing.T) {
	const rawSecret = "service-secret"
	hash, err := bcrypt.GenerateFromPassword([]byte(rawSecret), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}

	key := &model.AccessKey{
		Base:       model.Base{ID: 10},
		AccessKey:  "pcc_ak_service",
		SecretHash: string(hash),
		Status:     model.AccessKeyStatusEnabled,
	}
	namespace := &model.Namespace{
		Base:      model.Base{ID: 20},
		Namespace: "production",
	}
	grant := &model.AccessKeyNamespace{
		AccessKeyID: 10,
		NamespaceID: 20,
		Permission:  model.PermissionReadOnly,
	}
	config := &model.Config{
		Base:        model.Base{ID: 30},
		NamespaceID: 20,
		ConfigName:  "database",
		Content:     `{"host":"db.internal"}`,
	}

	newUsecase := func(permission runtimePermissionRepository) *RuntimeConfigUsecase {
		return &RuntimeConfigUsecase{
			accessKeys:  runtimeAccessKeyFake{key: key},
			namespaces:  runtimeNamespaceFake{namespace: namespace},
			permissions: permission,
			configs:     runtimeConfigFake{configs: []*model.Config{config}},
		}
	}

	t.Run("authorized read", func(t *testing.T) {
		got, getErr := newUsecase(runtimePermissionFake{permission: grant}).Get(
			context.Background(),
			key.AccessKey,
			"pcc_sk_"+rawSecret,
			namespace.Namespace,
			config.ConfigName,
		)
		if getErr != nil {
			t.Fatalf("Get() error = %v", getErr)
		}
		if got != config {
			t.Fatalf("Get() config = %#v, want %#v", got, config)
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		_, getErr := newUsecase(runtimePermissionFake{permission: grant}).Get(
			context.Background(),
			key.AccessKey,
			"pcc_sk_wrong",
			namespace.Namespace,
			config.ConfigName,
		)
		if !errors.Is(getErr, domain.ErrUnauthenticated) {
			t.Fatalf("Get() error = %v, want unauthenticated", getErr)
		}
	})

	t.Run("missing namespace grant", func(t *testing.T) {
		_, getErr := newUsecase(runtimePermissionFake{err: repository.ErrNotFound}).Get(
			context.Background(),
			key.AccessKey,
			"pcc_sk_"+rawSecret,
			namespace.Namespace,
			config.ConfigName,
		)
		if !errors.Is(getErr, domain.ErrForbidden) {
			t.Fatalf("Get() error = %v, want forbidden", getErr)
		}
	})
}
