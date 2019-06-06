package safecache

import (
	"sync"

	"fmt"

	"github.com/astaxie/beego/cache"
)

var (
	mutex   sync.Mutex
	regName map[string]struct{}
)

func init() {
	regName = make(map[string]struct{}, 10)
}

func NewSafeCache(adapterName, config string) (adapter cache.Cache, err error) {
	mutex.Lock()
	if _, ok := regName[adapterName]; ok {
		panic(fmt.Errorf("NewSafeCache name duplicate"))
	}
	regName[adapterName] = struct{}{}
	cache.Register(adapterName, cache.NewMemoryCache)
	ad, err := cache.NewCache(adapterName, config)
	mutex.Unlock()
	return ad, err
}
