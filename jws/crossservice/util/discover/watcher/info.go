package watcher

import (
	"vcs.taiyouxi.net/jws/crossservice/util/discover"
)

//ManInfo ..
type ManInfo struct {
	AsPath    map[string]Info
	AsService map[string]Info
	AsVersion map[string]Info
	AsIndex   map[string]Info
}

//Info ..
type Info struct {
	Path    string
	Service discover.Service
}

//NewManInfo ..
func NewManInfo() *ManInfo {
	m := &ManInfo{
		AsPath:    make(map[string]Info),
		AsService: make(map[string]Info),
		AsVersion: make(map[string]Info),
		AsIndex:   make(map[string]Info),
	}
	return m
}

//CheckAsPath ..
func (m *ManInfo) CheckAsPath(p string) bool {
	_, exist := m.AsPath[p]
	return exist
}

//GetAsPath ..
func (m *ManInfo) GetAsPath(p string) Info {
	return m.AsPath[p]
}

//AddInfo ..
func (m *ManInfo) AddInfo(p string, s discover.Service) {
	m.UpdateInfo(p, s)
}

//UpdateInfo ..
func (m *ManInfo) UpdateInfo(p string, s discover.Service) {
	info := Info{
		Path:    p,
		Service: s,
	}
	m.AsPath[info.Path] = info
	m.AsService[info.Service.Service] = info
	m.AsVersion[info.Service.Version] = info
	m.AsIndex[info.Service.Index] = info
}

//DelInfo ..
func (m *ManInfo) DelInfo(p string) {
	info, exist := m.AsPath[p]
	if false == exist {
		return
	}
	delete(m.AsPath, info.Path)
	delete(m.AsService, info.Service.Service)
	delete(m.AsVersion, info.Service.Version)
	delete(m.AsIndex, info.Service.Index)
}
