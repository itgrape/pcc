package model

import (
	"time"

	"gorm.io/gorm"

	"github.com/itgrape/pcc/server/pkg/idgen"
)

// Base 实体表的公共字段。主键由雪花算法生成，禁用自增；version 用于乐观锁；deleted_at 承载软删除。
type Base struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement:false"`
	Version   uint           `gorm:"column:version;not null;default:0"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// BeforeCreate 在插入前为空主键分配雪花 ID。
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == 0 {
		id, err := idgen.NextID()
		if err != nil {
			return err
		}
		b.ID = id
	}
	return nil
}
