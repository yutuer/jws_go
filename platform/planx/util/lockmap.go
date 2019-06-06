package util

import (
	"sync"
)

type LockMap struct {
	mu       sync.RWMutex
	services map[string]interface{}
}

func NewLockMap(cap int) *LockMap {
	return &LockMap{
		services: make(map[string]interface{}, cap),
	}
}

func (s *LockMap) Set(name string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	//if _, ok := s.services[name]; ok {
	//logs.Critical("Service %s has been registered", name)
	//return
	//}
	s.services[name] = v
}

func (s *LockMap) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[name]; ok {
		delete(s.services, name)
	}
}

//func (s *LockMap) Update(name string, v interface{}) {
//s.mu.Lock()
//defer s.mu.Unlock()
//s.services[name] = v
//}

func (s *LockMap) Get(name string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.services[name]
	return v, ok
}

func (s *LockMap) GetLen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.services)
}
