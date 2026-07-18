// Package infra 承载与外部系统对接的基础设施：数据库、密码哈希、令牌等。
package infra

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/itgrape/pcc/server/internal/query"
	"github.com/itgrape/pcc/server/pkg/idgen"
)

// DBConfig 数据库连接配置。
type DBConfig struct {
	DSN            string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLt      time.Duration
	SnowflakeNode  int64
}

// OpenDB 初始化 GORM（MariaDB）、连接池、雪花 ID 生成器与 gorm-gen 全局查询单例。
func OpenDB(cfg DBConfig) (*gorm.DB, error) {
	if err := idgen.Init(cfg.SnowflakeNode); err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		TranslateError:         true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("infra: open db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("infra: get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	if cfg.ConnMaxLt > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLt)
	}

	query.SetDefault(db)
	return db, nil
}
