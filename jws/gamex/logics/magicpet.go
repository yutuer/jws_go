package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// SetStateOfShowMagicPet : 设置灵宠显示状态
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgSetStateOfShowMagicPet 设置灵宠显示状态请求消息定义
type reqMsgSetStateOfShowMagicPet struct {
	Req
	ReqStateOfShowMagicPet bool `codec:"req_state_of_show_magic_pet"` // 客户端发送的灵宠形象状态
}

// rspMsgSetStateOfShowMagicPet 设置灵宠显示状态回复消息定义
type rspMsgSetStateOfShowMagicPet struct {
	SyncResp
	ResStateOfShowMagicPet bool `codec:"res_state_of_show_magic_pet"` // 服务器返回的灵宠形象状态
}

// SetStateOfShowMagicPet 设置灵宠显示状态:
func (p *Account) SetStateOfShowMagicPet(r servers.Request) *servers.Response {
	req := new(reqMsgSetStateOfShowMagicPet)
	rsp := new(rspMsgSetStateOfShowMagicPet)

	initReqRsp(
		"Attr/SetStateOfShowMagicPetRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.SetStateOfShowMagicPetHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// ShowMagicPet : 传输英雄灵宠信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgShowMagicPet 传输英雄灵宠信息请求消息定义
type reqMsgShowMagicPet struct {
	Req
}

// rspMsgShowMagicPet 传输英雄灵宠信息回复消息定义
type rspMsgShowMagicPet struct {
	SyncResp
	HeroMagicPetsInfo [][]byte `codec:"hero_magic_pet_info"` // 所有英雄的灵宠信息
}

// ShowMagicPet 传输英雄灵宠信息:
func (p *Account) ShowMagicPet(r servers.Request) *servers.Response {
	req := new(reqMsgShowMagicPet)
	rsp := new(rspMsgShowMagicPet)

	initReqRsp(
		"Attr/ShowMagicPetRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ShowMagicPetHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// HeroMagicPetInfo 传输英雄灵宠信息
type HeroMagicPetInfo struct {
	HeroID             int64    `codec:"hero_id"`              // 武将ID
	PetLev             int64    `codec:"pet_lev"`              // 灵宠等级
	PetStar            int64    `codec:"pet_star"`             // 灵宠星级
	PetAptitudes       [][]byte `codec:"pet_aptitude"`         // 灵宠资质
	CasualPetAptitude  [][]byte `codec:"casual_pet_aptitude"`  // 临时资质
	PetCompreTalent    int64    `codec:"pet_compre_talent"`    // 综合资质
	CasualCompreTalent int64    `codec:"casual_compre_talent"` // 临时综合
	ShowMagicPet       bool     `codec:"show_magic_pet"`       // 显示灵宠形象
}

// MagicPetAptitude 传输英雄灵宠信息
type MagicPetAptitude struct {
	PetAptitudeType  int64 `codec:"pet_aptitude_type"`  // 灵宠资质类型
	PetAptitudeValue int64 `codec:"pet_aptitude_value"` // 灵宠资质数值
}

// MagicPetLevUp : 英雄灵宠升级
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgMagicPetLevUp 英雄灵宠升级请求消息定义
type reqMsgMagicPetLevUp struct {
	Req
	HeroID int64 `codec:"hero_id"` // 武将ID
}

// rspMsgMagicPetLevUp 英雄灵宠升级回复消息定义
type rspMsgMagicPetLevUp struct {
	SyncResp
	HeroID int64 `codec:"hero_id"` // 武将ID
	Level  int64 `codec:"level"`   // 灵宠升级后等级
}

// MagicPetLevUp 英雄灵宠升级:
func (p *Account) MagicPetLevUp(r servers.Request) *servers.Response {
	req := new(reqMsgMagicPetLevUp)
	rsp := new(rspMsgMagicPetLevUp)

	initReqRsp(
		"Attr/MagicPetLevUpRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.MagicPetLevUpHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// MagicPetStarUp : 英雄灵宠升星
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgMagicPetStarUp 英雄灵宠升星请求消息定义
type reqMsgMagicPetStarUp struct {
	Req
	HeroID  int64 `codec:"Hero_id"` // 武将ID
	Special bool  `codec:"special"` // 是否使用保星符
}

// rspMsgMagicPetStarUp 英雄灵宠升星回复消息定义
type rspMsgMagicPetStarUp struct {
	SyncResp
	HeroID int64 `codec:"hero_id"` // 武将ID
	Star   int64 `codec:"star"`    // 灵宠升星后星级
}

// MagicPetStarUp 英雄灵宠升星:
func (p *Account) MagicPetStarUp(r servers.Request) *servers.Response {
	req := new(reqMsgMagicPetStarUp)
	rsp := new(rspMsgMagicPetStarUp)

	initReqRsp(
		"Attr/MagicPetStarUpRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.MagicPetStarUpHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// MagicPetChangeTalent : 英雄灵宠洗练
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgMagicPetChangeTalent 英雄灵宠洗练请求消息定义
type reqMsgMagicPetChangeTalent struct {
	Req
	HeroID  int64 `codec:"hero_id"` // 武将ID
	Special bool  `codec:"special"` // 是否使用道具
}

// rspMsgMagicPetChangeTalent 英雄灵宠洗练回复消息定义
type rspMsgMagicPetChangeTalent struct {
	SyncResp
	HeroID             int64    `codec:"hero_id"`              // 武将ID
	CompreTalent       int64    `codec:"compre_talent"`        // 综合资质
	CasualCompreTalent int64    `codec:"casual_compre_talent"` // 临时综合
	Talents            [][]byte `codec:"talents"`              // 灵宠资质
	CasualTalents      [][]byte `codec:"casual_talnets"`       // 临时资质
}

// MagicPetChangeTalent 英雄灵宠洗练:
func (p *Account) MagicPetChangeTalent(r servers.Request) *servers.Response {
	req := new(reqMsgMagicPetChangeTalent)
	rsp := new(rspMsgMagicPetChangeTalent)

	initReqRsp(
		"Attr/MagicPetChangeTalentRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.MagicPetChangeTalentHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// MagicPetSaveTalent : 英雄灵宠洗练保存
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgMagicPetSaveTalent 英雄灵宠洗练保存请求消息定义
type reqMsgMagicPetSaveTalent struct {
	Req
	HeroID int64 `codec:"hero_id"` // 武将ID
}

// rspMsgMagicPetSaveTalent 英雄灵宠洗练保存回复消息定义
type rspMsgMagicPetSaveTalent struct {
	SyncResp
	HeroID             int64    `codec:"hero_id"`              // 武将ID
	CompreTalent       int64    `codec:"compre_talent"`        // 综合资质
	CasualCompreTalent int64    `codec:"casual_compre_talent"` // 临时综合
	Talents            [][]byte `codec:"talents"`              // 灵宠资质
	CasualTalents      [][]byte `codec:"casual_talnets"`       // 临时资质
}

// MagicPetSaveTalent 英雄灵宠洗练保存:
func (p *Account) MagicPetSaveTalent(r servers.Request) *servers.Response {
	req := new(reqMsgMagicPetSaveTalent)
	rsp := new(rspMsgMagicPetSaveTalent)

	initReqRsp(
		"Attr/MagicPetSaveTalentRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.MagicPetSaveTalentHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
