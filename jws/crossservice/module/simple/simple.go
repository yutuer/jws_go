package simple

import (
	"vcs.taiyouxi.net/jws/crossservice/module"
)

//..
const (
	ModuleID               = "simple"
	MethodSimpleSyncID     = "simplesync"
	MethodSimpleAsyncID    = "simpleasync"
	MethodSimpleTransferID = "simpletransfer"
)

func init() {
	module.RegModule(&Generator{})
}

//Simple ..
type Simple struct {
	module.BaseModule
}

//Generator ..
type Generator struct {
}

//ModuleID ..
func (g *Generator) ModuleID() string {
	return ModuleID
}

//NewModule ..
func (g *Generator) NewModule(group uint32) module.Module {
	moduleSimple := &Simple{
		BaseModule: module.BaseModule{
			GroupID: group,
			Module:  ModuleID,
			Methods: map[string]module.Method{},
			Static:  true,
		},
	}
	moduleSimple.Methods[MethodSimpleSyncID] = newMethodSimpleSync(moduleSimple)
	moduleSimple.Methods[MethodSimpleAsyncID] = newMethodSimpleAsync(moduleSimple)
	moduleSimple.Methods[MethodSimpleTransferID] = newMethodSimpleTransfer(moduleSimple)

	return moduleSimple
}
