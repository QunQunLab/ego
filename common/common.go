package common

import (
	"github.com/QunQunLab/ego/error"
)

const (
	ErrOK                = 0
	ErrGeneralUnknown    = 100
	ErrGeneralForbidden  = 101
	ErrGeneralBadRequest = 102
)

// Define standard error response with errcode and errmsg
var (
	Success = error.Errorf{Code: ErrOK, Msg: map[string]string{"cn": "成功", "en": "success"}} // 200

	// 404
	Unknown = error.Errorf{Code: ErrGeneralUnknown, Msg: map[string]string{"cn": "未知错误:%v", "en": "ERR_UNKNOWN:%v"}}

	// 401
	Unauthorized = error.Errorf{Code: ErrGeneralUnknown, Msg: map[string]string{"cn": "未登录或登录已过期", "en": "Unauthorized request"}}

	// 403
	Forbidden = error.Errorf{Code: ErrGeneralForbidden, Msg: map[string]string{"cn": "禁止访问", "en": "You have no permission to access"}}

	// 400
	BadRequest = error.Errorf{Code: ErrGeneralBadRequest, Msg: map[string]string{"cn": "错误请求", "en": "You have send an error request"}}
)
