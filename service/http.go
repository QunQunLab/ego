package service

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/QunQunLab/ego/conf"
	"github.com/QunQunLab/ego/log"
)

const (
	EGOVersion = "EGO 0.0.1"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

var (
	ErrorController = fmt.Errorf("controller is not ControllerInterface")
)

// ControllerInfo holds information about the controller.
type ControllerInfo struct {
	controllerType reflect.Type
	httpMethod     string
	method         string
	pattern        string
}

// HttpService default http service
type HttpService struct {
	pool sync.Pool
	ctx  *Context

	//key:controller/method: val:controllerInfo
	routMap map[string]interface{}
}

func (s *HttpService) Name() string {
	return "DefaultHttpService"
}

func (s *HttpService) Init() error {
	return nil
}

func (s *HttpService) Register(c interface{}) {
	reflectVal := reflect.ValueOf(c)
	rt := reflectVal.Type()
	ct := reflect.Indirect(reflectVal).Type()
	controllerName := strings.TrimSuffix(ct.Name(), "Controller")
	for i := 0; i < rt.NumMethod(); i++ {
		route := &ControllerInfo{}
		route.controllerType = ct
		route.httpMethod = "GET" // todo
		route.method = rt.Method(i).Name
		pattern := path.Join("/", strings.ToLower(controllerName), strings.ToLower(rt.Method(i).Name))
		route.pattern = pattern
		s.routMap[pattern] = route
	}
}

func (s *HttpService) Start() error {
	section := conf.Get("http_conf")
	port, err := section.Uint("port")
	if err != nil {
		log.Warn("HTTP_CONF:PORT config is undefined. Using port :8080 by default")
		port = 8080
	}
	address := fmt.Sprintf(":%d", port)
	log.Info("Listening and serving HTTP on %s", address)
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	go http.Serve(l, s)

	return nil
}

func (s *HttpService) RunMode() string {
	return HttpMode
}

// ServeHTTP conforms to the http.Handler interface.
func (s *HttpService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := s.pool.Get().(*Context)
	defer s.pool.Put(c)

	c.ResponseWriter = w
	c.Request = req
	c.S = time.Now()

	s.handleHTTPRequest(c)
}

func (s *HttpService) handleHTTPRequest(ctx *Context) {
	defer func() {
		if err := recover(); err != nil {
			var errMsg string
			switch e := err.(type) {
			case error:
				errMsg = e.Error()

			default:
				return
			}

			log.Error("handleHTTPRequest err:%v", errMsg)
			http.Error(ctx.ResponseWriter, errMsg, http.StatusInternalServerError)
		}
	}()

	// cors domain setting
	ctx.ResponseWriter.Header().Set("Server", EGOVersion)
	if ref := ctx.Request.Referer(); ref != "" {
		if u, err := url.Parse(ref); nil == err {
			corsDomain := conf.GetKey("cors_domain")
			if corsDomain != "" {
				if "*" == corsDomain || strings.Contains(","+corsDomain+",", ","+u.Host+",") {
					ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", u.Scheme+"://"+u.Host)
					ctx.ResponseWriter.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}
		}
	}

	//httpMethod := c.Request.Method
	urlPath := strings.ToLower(ctx.Request.URL.Path)
	obj, ok := s.routMap[urlPath]
	if !ok {
		log.Error("the uri:%v not find.", urlPath)
		//if 50x error has been removed from errorMap
		serveError(ctx, http.StatusNotFound, default404Body)
		return
	}

	if c, ok := obj.(*ControllerInfo); ok {
		var execController ControllerInterface
		vc := reflect.New(c.controllerType)
		execController, ok = vc.Interface().(ControllerInterface)
		if !ok {
			panic(ErrorController)
		}

		defer func() {
			if err := recover(); err != nil {
				execController.RenderError(err)
			}
		}()

		ctx.ReqMethod = c.pattern

		// 1.0 prepare
		execController.Prepare(ctx)

		// 1.1 before start execute logical
		execController.BeforeProcess()

		// 2.0 controller method
		method := vc.MethodByName(c.method)
		in := make([]reflect.Value, 0)
		method.Call(in)
	}
}

// Stop stop service
func (s *HttpService) Stop(w *sync.WaitGroup) {
	w.Done()
}

func serveError(ctx *Context, code int, defaultMessage []byte) {
	var mimeJson = []string{"application/json"}
	ctx.ResponseWriter.Header()["Content-Type"] = mimeJson
	_, err := ctx.ResponseWriter.Write(defaultMessage)
	if err != nil {
		log.Error("cannot write message to writer during serve error: %v", err)
	}
	return
}

// NewHttpService new default tcp service
func NewHttpService() *HttpService {
	service := &HttpService{
		routMap: map[string]interface{}{},
	}
	service.pool.New = func() interface{} {
		return &Context{}
	}
	return service
}
