package types

type Namespace struct {
	ID          string `json:"id"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateNamespaceReq struct {
	Namespace   string `json:"namespace" binding:"required,max=64"`
	Description string `json:"description,omitempty" binding:"max=255"`
}

type UpdateNamespaceReq struct {
	Description string `json:"description" binding:"required,max=255"`
}
