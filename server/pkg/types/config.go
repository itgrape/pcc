package types

type Config struct {
	ID          string `json:"id"`
	NamespaceID string `json:"namespace_id"`
	ConfigName  string `json:"config_name"`
	Format      int8   `json:"format"`
	Content     string `json:"content,omitempty"`
	Description string `json:"description"`
	Version     uint   `json:"version"`
	CreatedBy   string `json:"created_by,omitempty"`
	UpdatedBy   string `json:"updated_by,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateConfigReq struct {
	ConfigName  string `json:"config_name" binding:"required,max=255"`
	Format      int8   `json:"format" binding:"required"`
	Content     string `json:"content" binding:"required"`
	Description string `json:"description,omitempty" binding:"max=255"`
}

type UpdateConfigReq struct {
	Content      string `json:"content" binding:"required"`
	Description  string `json:"description,omitempty" binding:"max=255"`
	ChangeReason string `json:"change_reason,omitempty" binding:"max=255"`
}

type ConfigHistory struct {
	ID            string `json:"id"`
	ConfigVersion uint   `json:"version"`
	Operator      string `json:"operator"`
	ChangeReason  string `json:"change_reason"`
	Content       string `json:"content,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type RollbackReq struct {
	ChangeReason string `json:"change_reason,omitempty" binding:"max=255"`
}
