package model

// Namespace 命名空间表，对应 namespaces。
type Namespace struct {
	Base
	Namespace   string `gorm:"column:namespace;uniqueIndex:uk_namespace;size:64;not null"`
	Description string `gorm:"column:description;size:255;not null;default:''"`
}

func (Namespace) TableName() string { return "namespaces" }
