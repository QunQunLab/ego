package service

import (
	"fmt"
	"sync"

	"github.com/QunQunLab/ego/conf"
	"github.com/QunQunLab/ego/log"
)

// RpcService default rpc service
type RpcService struct {
	pool sync.Pool
	ctx  *Context
}

// Name the name of the service
func (s *RpcService) Name() string {
	return "DefaultRPCService"
}

// Init init service
func (s *RpcService) Init() error {
	return nil
}

// Register register service handler
func (s *RpcService) Register(h interface{}) {
}

// Start start a service no blocking
func (s *RpcService) Start() error {
	section := conf.Get("rpc_conf")
	port, err := section.Uint("port")
	if err != nil {
		log.Warn("RPC_CONF:PORT config is undefined. Using port :8081 by default")
		port = 8081
	}
	address := fmt.Sprintf(":%d", port)
	log.Info("Listening and serving RPC on %s", address)
	return nil
}

func (s *RpcService) RunMode() string {
	return RPCMode
}

// Stop wait for all job done
// then call sync.WaitGroup.Done
func (s *RpcService) Stop(w *sync.WaitGroup) {
	w.Done()
}

// NewRpcService new default rpc service
func NewRpcService() *RpcService {
	service := &RpcService{}
	service.pool.New = func() interface{} {
		return &Context{}
	}
	return service
}
