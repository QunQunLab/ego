package service

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Context
type Context struct {
	http.ResponseWriter
	*http.Request

	ReqMethod string
	S         time.Time
}

func (ctx *Context) Reset(rw http.ResponseWriter, r *http.Request) {
	ctx.ResponseWriter = rw
	ctx.Request = r
}

type RequestWrap struct {
	Form url.Values
}

func (r *RequestWrap) FormValue(key string) string {
	if vs := r.Form[key]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

// ControllerInterface
type ControllerInterface interface {
	Prepare(ctx *Context)
	BeforeProcess()
	Render(...interface{})
	RenderError(interface{})
}

type Controller struct {
	Ctx   *Context
	RWrap *RequestWrap
}

func (c *Controller) Prepare(ctx *Context) {
	c.Ctx = ctx
	c.Ctx.S = time.Now()
	c.Ctx.ResponseWriter = ctx.ResponseWriter
	c.Ctx.Request = ctx.Request
	c.Ctx.Request.ParseMultipartForm(32 << 20) //32M

	c.RWrap = &RequestWrap{ctx.Request.Form}
}

func (c *Controller) BeforeProcess() {
}

func (c *Controller) Render(...interface{}) {
}

func (c *Controller) RenderError(interface{}) {
}

func (c *Controller) GetCookie(key string) string {
	cookie, err := c.Ctx.Request.Cookie(key)
	if err == nil {
		return cookie.Value
	}
	return ""
}

func (c *Controller) GetHeader(key string) string {
	return c.Ctx.Request.Header.Get(key)
}

func (c *Controller) _getFormValue(key string) string {
	val := c.RWrap.FormValue(key)
	return strings.Trim(val, " \r\t\v")
}

func (c *Controller) GetString(key string, defaultValues ...string) string {
	defaultValue := ""

	if len(defaultValues) > 0 {
		defaultValue = defaultValues[0]
	}

	ret := c._getFormValue(key)
	if ret == "" {
		ret = defaultValue
	}
	return ret
}

func (c *Controller) GetSlice(key string, separators ...string) []string {
	separator := ","
	if len(separators) > 0 {
		separator = separators[0]
	}

	value := c.GetString(key)
	if "" == value {
		return nil
	}

	var slice []string
	for _, part := range strings.Split(value, separator) {
		slice = append(slice, strings.Trim(part, " \r\t\v"))
	}
	return slice
}

func (c *Controller) GetSliceInt(key string, separators ...string) []int {
	slice := c.GetSlice(key, separators...)

	if nil == slice {
		return nil
	}

	var sliceInt []int
	for _, val := range slice {
		if val, err := strconv.Atoi(val); nil == err {
			sliceInt = append(sliceInt, val)
		}
	}

	return sliceInt
}

func (c *Controller) GetParams() map[string]string {
	if c.RWrap.Form == nil {
		return nil
	}

	params := map[string]string{}
	for k, v := range c.RWrap.Form {
		if len(v) > 0 {
			params[k] = strings.Trim(v[0], " \r\t\v")
		}
	}

	return params
}

func (c *Controller) GetArray(key string) []string {
	if c.RWrap.Form == nil {
		return nil
	}
	vs := c.RWrap.Form[key]
	return vs
}

func (c *Controller) GetInt(key string, defaultValues ...int) int {
	defaultValue := 0

	if len(defaultValues) > 0 {
		defaultValue = defaultValues[0]
	}

	ret, err := strconv.Atoi(c._getFormValue(key))
	if err != nil {
		ret = defaultValue
	}
	return ret
}

func (c *Controller) GetInt64(key string, defaultValues ...int64) int64 {
	defaultValue := int64(0)

	if len(defaultValues) > 0 {
		defaultValue = defaultValues[0]
	}

	ret, err := strconv.ParseInt(c._getFormValue(key), 10, 64)
	if err != nil {
		ret = defaultValue
	}
	return ret
}

func (c *Controller) GetBool(key string, defaultValues ...bool) bool {
	defaultValue := false

	if len(defaultValues) > 0 {
		defaultValue = defaultValues[0]
	}

	ret, err := strconv.ParseBool(c._getFormValue(key))
	if err != nil {
		ret = defaultValue
	}
	return ret
}

func (c *Controller) GetFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.Ctx.Request.FormFile(key)
}

func (c *Controller) GetIp() string {
	r := c.Ctx.Request

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" || ip == "127.0.0.1" {
		ip = r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.Header.Get("Host")
			if ip == "" {
				ip = r.RemoteAddr
			}
		}
	} else {
		//X-Forwarded-For 的格式 client1, proxy1, proxy2
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			ip = ips[0]
		}
	}

	//去除端口号
	ips := strings.Split(ip, ":")
	if len(ips) > 0 {
		ip = ips[0]
	}

	return ip
}

func (c *Controller) GetRequestUri() string {
	if nil != c.Ctx.Request {
		return fmt.Sprint(c.Ctx.Request.URL)
	}

	return ""
}

func (c *Controller) GetUA() string {
	if nil != c.Ctx.Request {
		return c.Ctx.Request.UserAgent()
	}
	return ""
}

func (c *Controller) SetCookie(key, val string, lifetime int) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    val,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   lifetime,
		Expires:  time.Now().Add(time.Second * time.Duration(lifetime)),
	}
	http.SetCookie(c.Ctx.ResponseWriter, cookie)
}

func (c *Controller) UnsetCookie(key string) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		MaxAge:   0,
		Expires:  time.Now().AddDate(-1, 0, 0),
	}
	http.SetCookie(c.Ctx.ResponseWriter, cookie)
}

func (c *Controller) SetHeader(key, val string) {
	c.Ctx.ResponseWriter.Header().Set(key, val)
}

func (c *Controller) SetHeaders(headers http.Header) {
	hs := c.Ctx.ResponseWriter.Header()
	for k, v := range headers {
		hs.Set(k, v[0])
	}
}
