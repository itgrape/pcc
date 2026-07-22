package types

type UserInfo struct {
	Owner       string            `json:"owner"`
	Username    string            `json:"name"`
	DisplayName string            `json:"displayName"`
	Email       string            `json:"email"`
	Phone       string            `json:"phone"`
	Avatar      string            `json:"avatar"`
	Roles       []string          `json:"roles"`
	Groups      []string          `json:"groups"`
	Properties  map[string]string `json:"properties"`
}
