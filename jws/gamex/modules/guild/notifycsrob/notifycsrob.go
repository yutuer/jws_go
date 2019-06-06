package notifycsrob

import "sync"

//notifyRefresh ..
type notifyRefresh func(guid string)

var callbacks []notifyRefresh
var lock sync.Mutex

func init() {
	callbacks = []notifyRefresh{}
}

//RegRefreshCallback ..
func RegRefreshCallback(callback notifyRefresh) {
	lock.Lock()
	defer lock.Unlock()

	callbacks = append(callbacks, callback)
}

//Call ..
func Call(guid string) {
	lock.Lock()
	for _, notify := range callbacks {
		notify(guid)
	}
	defer lock.Unlock()
}
