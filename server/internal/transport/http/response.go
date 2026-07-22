package httpapi

import "github.com/itgrape/pcc/server/pkg/types"

func resOK(data any) types.Response {
	return types.Response{Code: 0, Message: "success", Data: data}
}

func resPage(list any, total int64, page, pageSize int) types.Response {
	return resOK(types.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}
