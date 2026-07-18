package model

import (
	"time"

	"gorm.io/gorm"

	"github.com/itgrape/pcc/server/pkg/idgen"
)

// ConfigHistory 配置历史版本表，对应 config_history。追加式写入，无 version/updated_at/deleted_at。
type ConfigHistory struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement:false"`
	ConfigID      uint64 `gorm:"column:config_id;not null"`
	NamespaceID   uint64 `gorm:"column:namespace_id;not null"`
	ConfigName    string `gorm:"column:config_name;size:255;not null"`
	Content       string `gorm:"column:content;type:longtext;not null"`
	ConfigVersion uint   `gorm:"column:config_version;not null"`
	Operator      string `gorm:"column:operator;size:64;not null;default:''"`
	ChangeReason  string `gorm:"column:change_reason;size:255;not null;default:''"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (ConfigHistory) TableName() string { return "config_history" }

// BeforeCreate 在插入前为空主键分配雪花 ID。
func (c *ConfigHistory) BeforeCreate(_ *gorm.DB) error {
	if c.ID == 0 {
		id, err := idgen.NextID()
		if err != nil {
			return err
		}
		c.ID = id
	}
	return nil
}
