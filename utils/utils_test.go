package utils

import (
	"testing"

	"github.com/QunQunLab/ego/error"
)

func TestInterceptor(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			switch e := err.(type) {
			case *error.Errorf:
				t.Log(e.GetMsg("cn"))
				t.Log(e.GetMsg("en"))

			default:
				t.Log("unknown error:", err)
			}
		}
	}()

	var ErrorPass = &error.Errorf{Code: 1, Msg: map[string]string{"cn": "密码错误", "en": "Invalid password"}}
	Interceptor(false, ErrorPass)
}

func TestInSlice(t *testing.T) {
	var s []string
	s = append(s, "a")
	s = append(s, "b")

	t.Log("a in slice:", InSlice("a", s))
	t.Log("ab in slice:", InSlice("ab", s))
}
