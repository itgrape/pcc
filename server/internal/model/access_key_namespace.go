package model

// Permission 密钥命名空间权限。
type Permission int8

const (
	PermissionReadOnly  Permission = 1 // 只读
	PermissionReadWrite Permission = 2 // 读写
)

// AccessKeyNamespace 密钥-命名空间关联表，对应 access_key_namespaces。
type AccessKeyNamespace struct {
	Base
	AccessKeyID uint64     `gorm:"column:access_key_id;uniqueIndex:uk_key_namespace;not null"`
	NamespaceID uint64     `gorm:"column:namespace_id;uniqueIndex:uk_key_namespace;not null"`
	Permission  Permission `gorm:"column:permission;not null;default:1"`
}

func (AccessKeyNamespace) TableName() string { return "access_key_namespaces" }
