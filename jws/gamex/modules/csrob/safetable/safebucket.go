package safetable

import "sync"

type safeBucket struct {
	sync.RWMutex
	count uint32
	m     map[string]interface{}
}

type elem struct {
	key string
	val interface{}
}

func newSafeBucket() *safeBucket {
	b := &safeBucket{}

	b.count = 0
	b.m = make(map[string]interface{})

	return b
}

func (b *safeBucket) get(key string) interface{} {
	b.RLock()
	defer b.RUnlock()

	return b.m[key]
}

func (b *safeBucket) set(key string, val interface{}) {
	b.Lock()
	defer b.Unlock()

	b.m[key] = val
}

func (b *safeBucket) del(key string) {
	b.Lock()
	defer b.Unlock()

	if nil != b.m[key] {
		delete(b.m, key)
	}
}
