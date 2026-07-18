package model

// AccessKeyStatus 密钥状态。
type AccessKeyStatus int8

const (
	AccessKeyStatusDisabled AccessKeyStatus = 0 // 禁用
	AccessKeyStatusEnabled  AccessKeyStatus = 1 // 启用
)

// AccessKey 访问密钥表，对应 access_keys。
type AccessKey struct {
	Base
	AccessKey   string           `gorm:"column:access_key;uniqueIndex:uk_access_key;size:64;not null"`
	SecretHash  string           `gorm:"column:secret_hash;size:128;not null" json:"-"`
	Status      AccessKeyStatus  `gorm:"column:status;not null;default:1"`
	Description string           `gorm:"column:description;size:255;not null;default:''"`
}

func (AccessKey) TableName() string { return "access_keys" }
