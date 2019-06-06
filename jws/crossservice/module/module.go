package module

import (
	"hash/crc32"
)

//Generator ..
type Generator interface {
	ModuleID() string
	NewModule(uint32) Module
}

//Module ..
type Module interface {
	ModuleID() string
	GetMethods() map[string]Method
	GetMethod(string) Method
	Hash(string) uint32
	HashMask() uint32
	GetGroupID() uint32

	SetFuncPush(f FuncPush)
	Push(uint32, string, string, Param) error

	IsStatic() bool
	Start()
	AfterStart()
	Stop()
	BeforeStop()
}

//Method ..
type Method interface {
	MethodID() string
	ModuleAt() Module
	NewParam() Param
	NewRet() Ret
	Do(Transaction, Param) (uint32, Ret)
}

//FuncPush ..
type FuncPush func(uint32, string, string, Param) error

//Param ..
type Param interface {
}

//Ret ..
type Ret interface {
}

//Transaction ..
type Transaction struct {
	GroupID    uint32
	HashSource string
}

//BaseModule ..
type BaseModule struct {
	Module   string
	Methods  map[string]Method
	funcPush FuncPush
	Static   bool
	GroupID  uint32
}

//ModuleID ..
func (m *BaseModule) ModuleID() string {
	return m.Module
}

//GetMethods ..
func (m *BaseModule) GetMethods() map[string]Method {
	return m.Methods
}

//GetMethod ..
func (m *BaseModule) GetMethod(method string) Method {
	return m.Methods[method]
}

//HashMask ..
func (m *BaseModule) HashMask() uint32 {
	return 4
}

//GetGroupID ..
func (m *BaseModule) GetGroupID() uint32 {
	return m.GroupID
}

//Hash ..
func (m *BaseModule) Hash(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}

//SetFuncPush ..
func (m *BaseModule) SetFuncPush(f FuncPush) {
	m.funcPush = f
}

//Push ..
func (m *BaseModule) Push(sid uint32, mo, me string, p Param) error {
	if nil == m.funcPush {
		return nil
	}
	return m.funcPush(sid, mo, me, p)
}

//IsStatic ..
func (m *BaseModule) IsStatic() bool {
	return m.Static
}

//Start ..
func (m *BaseModule) Start() {
	return
}

//AfterStart ..
func (m *BaseModule) AfterStart() {
	return
}

//Stop ..
func (m *BaseModule) Stop() {
	return
}

//BeforeStop ..
func (m *BaseModule) BeforeStop() {
	return
}

//BaseMethod ..
type BaseMethod struct {
	Method string
	Module Module
}

//MethodID ..
func (m *BaseMethod) MethodID() string {
	return m.Method
}

//ModuleAt ..
func (m *BaseMethod) ModuleAt() Module {
	return m.Module
}

//NewParam ..
func (m *BaseMethod) NewParam() Param {
	return &BaseParam{}
}

//NewRet ..
func (m *BaseMethod) NewRet() Ret {
	return &BaseRet{}
}

//Do ..
func (m *BaseMethod) Do(Transaction, Param) (uint32, Ret) {
	return 0, nil
}

//BaseParam ..
type BaseParam struct {
}

//BaseRet ..
type BaseRet struct {
}
