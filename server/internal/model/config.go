package model

// Config 配置主表，对应 configs。
type Config struct {
	Base
	NamespaceID uint64 `gorm:"column:namespace_id;uniqueIndex:uk_namespace_config;not null"`
	ConfigName  string `gorm:"column:config_name;uniqueIndex:uk_namespace_config;size:255;not null"`
	Format      string `gorm:"column:format;size:64;not null;default:'json'"`
	Content     string `gorm:"column:content;type:longtext;not null"`
	Description string `gorm:"column:description;size:255;not null;default:''"`
	CreatedBy   string `gorm:"column:created_by;size:64;not null;default:''"`
	UpdatedBy   string `gorm:"column:updated_by;size:64;not null;default:''"`
}

func (Config) TableName() string { return "configs" }
