package client

import (
	"vcs.taiyouxi.net/jws/crossservice/module"
)

//modules ..
type modules struct {
	ms map[string]module.Module
}

//newModules ..
func newModules() *modules {
	m := &modules{
		ms: make(map[string]module.Module),
	}
	return m
}

func (m *modules) loadModules() {
	for _, g := range module.LoadModulesList {
		m.ms[g.ModuleID()] = g.NewModule(0)
	}
}

func (m *modules) getMethod(moName string, meName string) module.Method {
	mo := m.ms[moName]
	if nil == mo {
		return nil
	}

	return mo.GetMethod(meName)
}
