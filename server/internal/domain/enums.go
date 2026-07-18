package domain

// SysEnums 系统枚举值，供前端下拉框使用。
type SysEnums struct {
	ConfigFormats []string     `json:"config_formats"`
	Permissions   []string     `json:"permissions"`
	KeyStatus     []EnumItem   `json:"key_status"`
}

// EnumItem 枚举项。
type EnumItem struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}
