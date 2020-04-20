package service

import (
	"fmt"
	"net/http"
)

// Context
type Context struct {
	http.ResponseWriter
	*http.Request
}

func (ctx *Context) Reset(rw http.ResponseWriter, r *http.Request) {
	ctx.ResponseWriter = rw
	ctx.Request = r
}

// ControllerInterface
type ControllerInterface interface {
	Prepare(ctx *Context)
	BeforeStart()
	Render()
	RenderError(ctx *Context)
}

type Controller struct {
	ctx *Context
}

func (c *Controller) Prepare(ctx *Context) {
	c.ctx = ctx
	fmt.Println("controller prepare")
}

func (c *Controller) BeforeStart() {
	fmt.Println("controller before start")
}

func (c *Controller) Render() {
	fmt.Println("controller render")
}

func (c *Controller) RenderError(ctx *Context) {
	fmt.Println("controller render error")
}
