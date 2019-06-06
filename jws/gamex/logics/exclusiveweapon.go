package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// ActivateExclusiveWeapon : 激活专属兵器
// 激活专属兵器

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgActivateExclusiveWeapon 激活专属兵器请求消息定义
type reqMsgActivateExclusiveWeapon struct {
	Req
	AvatarId int64 `codec:"avatar"` // 武将ID
}

// rspMsgActivateExclusiveWeapon 激活专属兵器回复消息定义
type rspMsgActivateExclusiveWeapon struct {
	SyncResp
}

// ActivateExclusiveWeapon 激活专属兵器: 激活专属兵器
func (p *Account) ActivateExclusiveWeapon(r servers.Request) *servers.Response {
	req := new(reqMsgActivateExclusiveWeapon)
	rsp := new(rspMsgActivateExclusiveWeapon)

	initReqRsp(
		"Attr/ActivateExclusiveWeaponRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ActivateExclusiveWeaponHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// EvolveExclusiveWeapon : 专属兵器升品
// 专属兵器升品

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgEvolveExclusiveWeapon 专属兵器升品请求消息定义
type reqMsgEvolveExclusiveWeapon struct {
	Req
	AvatarId int64 `codec:"avatar"` // 武将ID
}

// rspMsgEvolveExclusiveWeapon 专属兵器升品回复消息定义
type rspMsgEvolveExclusiveWeapon struct {
	SyncRespWithRewards
}

// EvolveExclusiveWeapon 专属兵器升品: 专属兵器升品
func (p *Account) EvolveExclusiveWeapon(r servers.Request) *servers.Response {
	req := new(reqMsgEvolveExclusiveWeapon)
	rsp := new(rspMsgEvolveExclusiveWeapon)

	initReqRsp(
		"Attr/EvolveExclusiveWeaponRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.EvolveExclusiveWeaponHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// PromoteExclusiveWeapon : 培养兵器升品
// 培养兵器升品

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgPromoteExclusiveWeapon 培养兵器升品请求消息定义
type reqMsgPromoteExclusiveWeapon struct {
	Req
	AvatarId     int64 `codec:"avatar"`   // 武将ID
	PromoteType  int64 `codec:"pro_type"` // 培养类型，0=培养 1=保存 2=取消
	PromoteByTen bool  `codec:"pro_ten"`  // true=培养十次
}

// rspMsgPromoteExclusiveWeapon 培养兵器升品回复消息定义
type rspMsgPromoteExclusiveWeapon struct {
	SyncResp
	PromoteByTen    bool  `codec:"pro_ten"`       // true=培养十次
	RealPromoteTime int64 `codec:"real_pro_time"` // 实际培养次数
}

// PromoteExclusiveWeapon 培养兵器升品: 培养兵器升品
func (p *Account) PromoteExclusiveWeapon(r servers.Request) *servers.Response {
	req := new(reqMsgPromoteExclusiveWeapon)
	rsp := new(rspMsgPromoteExclusiveWeapon)

	initReqRsp(
		"Attr/PromoteExclusiveWeaponRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.PromoteExclusiveWeaponHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// ResetExclusiveWeapon : 重置神兵
// 重置神兵

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgResetExclusiveWeapon 重置神兵请求消息定义
type reqMsgResetExclusiveWeapon struct {
	Req
	AvatarId int64 `codec:"avatar"` // 武将ID
}

// rspMsgResetExclusiveWeapon 重置神兵回复消息定义
type rspMsgResetExclusiveWeapon struct {
	SyncRespWithRewards
}

// ResetExclusiveWeapon 重置神兵: 重置神兵
func (p *Account) ResetExclusiveWeapon(r servers.Request) *servers.Response {
	req := new(reqMsgResetExclusiveWeapon)
	rsp := new(rspMsgResetExclusiveWeapon)

	initReqRsp(
		"Attr/ResetExclusiveWeaponRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ResetExclusiveWeaponHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
