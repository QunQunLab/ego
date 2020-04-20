package utils

import (
	"github.com/QunQunLab/ego/error"
)

func Interceptor(guard bool, err *error.Errorf, f ...interface{}) {
	if !guard {
		panic(
			&error.Errorf{Code: err.Code,
				Msg:  err.Msg,
				Fmt:  f,
				Data: nil},
		)
	}
}

// InSlice checks given string in string slice or not.
func InSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}
