package infra

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/itgrape/pcc/server/internal/config"
	"github.com/itgrape/pcc/server/internal/query"
	"github.com/itgrape/pcc/server/pkg/idgen"
)

func OpenDB(dbCfg config.DBConfig, nodeID int64) (*gorm.DB, error) {
	if err := idgen.Init(nodeID); err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dbCfg.DSN), &gorm.Config{
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
	sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	if dbCfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	}

	query.SetDefault(db)
	return db, nil
}
