package types

type AccessKey struct {
	ID          string `json:"id"`
	AccessKey   string `json:"access_key"`
	Status      int    `json:"status"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateAccessKeyReq struct {
	Description string `json:"description,omitempty" binding:"max=255"`
}

type AccessKeyCreated struct {
	ID          string `json:"id"`
	AccessKey   string `json:"access_key"`
	Secret      string `json:"secret"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

type UpdateAccessKeyReq struct {
	Description *string `json:"description,omitempty" binding:"max=255"`
	Status      *int    `json:"status,omitempty"`
}

type AccessKeyNamespace struct {
	AccessKeyID string `json:"access_key_id"`
	NamespaceID string `json:"namespace_id"`
	Permission  int    `json:"permission"`
	CreatedAt   string `json:"created_at"`
}

type GrantNamespaceReq struct {
	NamespaceID string `json:"namespace_id" binding:"required"`
	Permission  int    `json:"permission" binding:"required"`
}

type UpdatePermissionReq struct {
	Permission int `json:"permission" binding:"required"`
}
