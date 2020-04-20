package service

import (
	"os"
	"os/signal"
	"sync"
)

type Handler func(os.Signal) bool

var (
	mu       sync.Mutex
	handlers = map[os.Signal][]Handler{}
	sigChan  = make(chan os.Signal)
)

func Register(sig os.Signal, handler Handler) {
	mu.Lock()
	defer mu.Unlock()
	if g, ok := handlers[sig]; ok {
		handlers[sig] = append(g, handler)
	} else {
		handlers[sig] = []Handler{handler}
	}
}

func Listen() {
	signals := make([]os.Signal, 0, len(handlers))
	for sig, _ := range handlers {
		signals = append(signals, sig)
	}
	signal.Notify(sigChan, signals...)
	for {
		select {
		case sig := <-sigChan:
			mu.Lock()
			g, ok := handlers[sig]
			mu.Unlock()
			if ok {
				for _, h := range g {
					if h(sig) {
						return
					}
				}
			}
		}
	}
}

func Wait(sig os.Signal) {
	Register(sig, func(os.Signal) bool {
		return true
	})
	Listen()
}
