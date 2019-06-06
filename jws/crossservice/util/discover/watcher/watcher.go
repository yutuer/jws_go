package watcher

import (
	"vcs.taiyouxi.net/jws/crossservice/util/discover"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Watcher ..
type Watcher struct {
	project string

	manInfo *ManInfo
	filters []Filter

	stopNotify chan struct{}

	OnServiceAdd    func(discover.Service)
	OnServiceUpdate func(discover.Service)
	OnServiceDel    func(discover.Service)
}

//NewWatcher ..
func NewWatcher() *Watcher {
	w := &Watcher{}
	w.manInfo = NewManInfo()
	w.stopNotify = make(chan struct{}, 5)
	return w
}

//SetProject ..
func (w *Watcher) SetProject(p string) {
	w.project = p
}

//SetFilter ..
func (w *Watcher) SetFilter(fs []Filter) {
	w.filters = fs
}

//Start ..
func (w *Watcher) Start() {
	hm := map[string]discover.HandleWatchEvent{
		discover.MakeServicePathAsProject(w.project): w.handleEvent,
	}
	go func() {
		defer logs.PanicCatcherWithInfo("Watching Discover...")
		if err := discover.StartWatcher(hm, w.stopNotify); nil != err {
			logs.Error("StartWatcher Error, %v", err)
		}
	}()
}

//Close ..
func (w *Watcher) Close() {
	w.stopNotify <- struct{}{}
}

func (w *Watcher) handleEvent(et discover.EventType, path string, service *discover.Service) {
	// logs.Debug("Watcher handleEvent %v: %s -> %v", et, path, service)
	switch et {
	case discover.EventPut:
		if nil == service {
			break
		}
		needFilter := false
		for _, filter := range w.filters {
			if false == filter.Check(*service) {
				needFilter = true
				break
			}
		}
		if w.manInfo.CheckAsPath(path) {
			w.manInfo.UpdateInfo(path, *service)
			if !needFilter && nil != w.OnServiceUpdate {
				w.OnServiceUpdate(*service)
			}
		} else {
			w.manInfo.AddInfo(path, *service)
			if !needFilter && nil != w.OnServiceAdd {
				w.OnServiceAdd(*service)
			}
		}
	case discover.EventDel:
		if false == w.manInfo.CheckAsPath(path) {
			break
		}
		info := w.manInfo.GetAsPath(path)
		w.manInfo.DelInfo(path)
		if nil != w.OnServiceDel {
			w.OnServiceDel(info.Service)
		}
	case discover.EventUnknown:
		logs.Debug("Watcher Got Unknown Event, path [%s]", path)
	}
}
