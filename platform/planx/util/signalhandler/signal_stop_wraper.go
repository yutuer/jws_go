package signalhandler

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Stoper interface {
	Stop()
}

type Reloader interface {
	Reload()
}

type ReloaderFunc func()

func (r ReloaderFunc) Reload() {
	r()
}

type StoperFunc func()

func (s StoperFunc) Stop() {
	s()
}

var (
	signalLock sync.RWMutex
	stopers    []Stoper
	reloaders  []Reloader
	quit       chan os.Signal
)

func init() {
	stopers = make([]Stoper, 0, 10)
	reloaders = make([]Reloader, 0, 10)
	quit = make(chan os.Signal, 1)

}

func SignalReloadFunc(f func()) {
	SignalReloadHandler(ReloaderFunc(f))
}

func SignalReloadHandler(r Reloader) {
	signalLock.Lock()
	defer signalLock.Unlock()
	reloaders = append(reloaders, r)
}

func signalReloadHandle() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGKILL)
	for {
		select {
		case <-quit:
			return
		case <-ch:
			fmt.Printf("Got %v\n", "SIGUSR2")
			go func() {
				signalLock.RLock()
				defer signalLock.RUnlock()
				for _, v := range reloaders {
					v.Reload()
				}
			}()
		}
	}
}

func SignalKillHandler(s Stoper) {
	signalLock.Lock()
	defer signalLock.Unlock()
	stopers = append(stopers, s)
}

func SignalKillFunc(f func()) {
	SignalKillHandler(StoperFunc(f))
}

func SignalKillHandle() {
	go signalReloadHandle()

	stopch := make(chan os.Signal, 1)
	signal.Notify(stopch, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	<-stopch
	func() {
		signalLock.RLock()
		defer signalLock.RUnlock()
		i := 0
		for _, v := range stopers {
			v.Stop()
			i++
		}
		close(quit)
	}()
}
