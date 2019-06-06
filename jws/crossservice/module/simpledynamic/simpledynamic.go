package simpledynamic

import (
	"math/rand"
	"time"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//..
const (
	ModuleID          = "simpledynamic"
	MethodGetStatusID = "getstatus"
)

//Generator ..
type Generator struct {
}

//ModuleID ..
func (g *Generator) ModuleID() string {
	return ModuleID
}

//NewModule ..
func (g *Generator) NewModule(group uint32) module.Module {
	moduleSimpleDynamic := &SimpleDynamic{
		BaseModule: module.BaseModule{
			GroupID: group,
			Module:  ModuleID,
			Methods: map[string]module.Method{},
			Static:  false,
		},
		status:   0,
		isClosed: false,
	}
	moduleSimpleDynamic.Methods[MethodGetStatusID] = newMethodGetStatus(moduleSimpleDynamic)

	return moduleSimpleDynamic
}

func init() {
	module.RegModule(&Generator{})
}

//SimpleDynamic ..
type SimpleDynamic struct {
	module.BaseModule

	status   uint32
	isClosed bool
}

//Start ..
func (s *SimpleDynamic) Start() {
	go func() {
		defer logs.PanicCatcherWithInfo("CrossService Server, SimpleDynamic run")
		rander := rand.New(rand.NewSource(time.Now().Unix()))
		statusAfter := time.After(time.Millisecond)
		for !s.isClosed {
			select {
			case <-statusAfter:
				s.status++
				statusAfter = time.After(time.Duration(rander.Int()%1000) * time.Millisecond)
			}
		}
	}()
}

//Stop ..
func (s *SimpleDynamic) Stop() {
	s.isClosed = true
}
