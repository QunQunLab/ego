package error

import (
	"fmt"
	"strings"
)

const (
	defaultLanguage = "cn"
)

// Errorf error format output
// Multi language support
// var ErrorPass = &Errorf{Code: 1, Msg: map[string]string{"cn": "密码错误", "en": "Invalid password"}}
type Errorf struct {
	Code int
	// string or map[string]string
	Msg  interface{}
	Fmt  []interface{}
	Data map[string]interface{}
}

func (e *Errorf) GetCode() int {
	return e.Code
}

func (e *Errorf) GetMsg(langs ...string) string {
	if len(e.Fmt) > 0 {
		if data, ok := e.Fmt[len(e.Fmt)-1].(map[string]interface{}); ok {
			e.Fmt = e.Fmt[0 : len(e.Fmt)-1]
			e.Data = data
		}
	}

	var (
		gMsg map[string]string

		msg  string
		ok   bool
		lang string
	)
	if msg, ok = e.Msg.(string); ok {
		return fmt.Sprintf(msg, e.Fmt...)
	} else if gMsg, ok = e.Msg.(map[string]string); ok {
		if len(langs) > 0 {
			lang = strings.ToLower(langs[0])
		}

		if lang != "" {
			if msg, ok = gMsg[lang]; ok {
				return fmt.Sprintf(msg, e.Fmt...)
			}
		}

		if msg, ok = gMsg[defaultLanguage]; ok {
			return fmt.Sprintf(msg, e.Fmt...)
		}

		for _, v := range gMsg {
			return fmt.Sprintf(v, e.Fmt...)
		}
	}

	return fmt.Sprint(e.Msg)
}

func (e *Errorf) GetData() map[string]interface{} {
	if e.Data == nil && len(e.Fmt) > 0 {
		if data, ok := e.Fmt[len(e.Fmt)-1].(map[string]interface{}); ok {
			return data
		}
	}

	return e.Data
}

func (e *Errorf) Error() string {
	return e.GetMsg()
}
