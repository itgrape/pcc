package domain

// PageReq 通用分页请求参数。
type PageReq struct {
	Page     int `json:"page" default:"1" minimum:"1"`
	PageSize int `json:"page_size" default:"20" maximum:"100" minimum:"1"`
	Keyword  string `json:"keyword,omitempty" doc:"模糊搜索关键词"`
}
