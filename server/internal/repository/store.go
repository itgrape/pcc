package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/itgrape/pcc/server/internal/query"
)

// Store 聚合所有仓储并提供事务入口。
type Store struct {
	q *query.Query

	Namespace          NamespaceRepository
	Config             ConfigRepository
	ConfigHistory      ConfigHistoryRepository
	AccessKey          AccessKeyRepository
	AccessKeyNamespace AccessKeyNamespaceRepository
}

// New 基于 *gorm.DB 构造顶层 Store。
func New(db *gorm.DB) *Store {
	return newStore(query.Use(db))
}

func newStore(q *query.Query) *Store {
	return &Store{
		q:                  q,
		Namespace:          &namespaceRepository{q: q},
		Config:             &configRepository{q: q},
		ConfigHistory:      &configHistoryRepository{q: q},
		AccessKey:          &accessKeyRepository{q: q},
		AccessKeyNamespace: &accessKeyNamespaceRepository{q: q},
	}
}

// Tx 在单个数据库事务中执行 fn。
func (s *Store) Tx(ctx context.Context, fn func(tx *Store) error) error {
	return s.q.Transaction(func(txq *query.Query) error {
		return fn(newStore(txq))
	})
}
