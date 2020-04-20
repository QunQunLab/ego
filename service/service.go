package service

import (
	"os"
	"sync"

	"github.com/QunQunLab/ego/log"
)

// Service service interface
type Service interface {
	// Name the name of the service
	Name() string

	// Init init service
	Init() error

	// Register register service handler
	Register(h interface{})

	// Start start a service no blocking
	Start() error

	// Stop wait for all job done
	// then call sync.WaitGroup.Done
	Stop(*sync.WaitGroup)
}

// Run start services
func Run(services []Service) {
	for _, s := range services {
		err := s.Init()
		if err != nil {
			log.Fatal("%v init err:%v", s.Name(), err)
		}
		log.Info("%v init ok", s.Name())
	}

	for _, s := range services {
		err := s.Start()
		if err != nil {
			log.Fatal("%v started err:%v", s.Name(), err)
		}
		log.Info("%v started", s.Name())
	}

	// wait signal
	// Ctrl+C or kill -p
	Wait(os.Interrupt)

	var wg sync.WaitGroup
	for _, s := range services {
		wg.Add(1)
		s.Stop(&wg)
		log.Info("%v stopped", s.Name())
	}
	wg.Wait()
	log.Info("all services exit")
}
