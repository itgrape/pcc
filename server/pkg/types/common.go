package types

// Response 统一响应格式。
type Response struct {
	Code    int         `json:"code" default:"0"`
	Message string      `json:"message" default:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 分页响应结构。
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}
