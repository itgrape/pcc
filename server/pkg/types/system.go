package types

type SysEnums struct {
	ConfigFormats []EnumItem `json:"config_formats"`
	Permissions   []EnumItem `json:"permissions"`
	KeyStatus     []EnumItem `json:"key_status"`
}

type EnumItem struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}
