package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// ShowBattleArmy : 传输战阵信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgShowBattleArmy 传输战阵信息请求消息定义
type reqMsgShowBattleArmy struct {
	Req
}

// rspMsgShowBattleArmy 传输战阵信息回复消息定义
type rspMsgShowBattleArmy struct {
	SyncResp
	BattleArmysInfo [][]byte `codec:"battle_army_info"` // 所有战阵信息
}

// ShowBattleArmy 传输战阵信息:
func (p *Account) ShowBattleArmy(r servers.Request) *servers.Response {
	req := new(reqMsgShowBattleArmy)
	rsp := new(rspMsgShowBattleArmy)

	initReqRsp(
		"Attr/ShowBattleArmyRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ShowBattleArmyHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// BattleArmyInfo 传输战阵信息
type BattleArmyInfo struct {
	BattleArmyID       int64 `codec:"battle_army_id"`       // 战阵ID
	BattleArmyAvatarID int64 `codec:"battle_army_avatarID"` // 战阵武将
	BattleArmyLev      int64 `codec:"battle_army_lev"`      // 战阵等级
}

// BattleArmyLevUp : 战阵升级／解锁协议
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBattleArmyLevUp 战阵升级／解锁协议请求消息定义
type reqMsgBattleArmyLevUp struct {
	Req
	BattleArmyID        int64 `codec:"battle_army_id"`       // 战阵ID
	CurrBattleArmyLevel int64 `codec:"curr_battle_army_lev"` // 升级前显示的战阵等级
}

// rspMsgBattleArmyLevUp 战阵升级／解锁协议回复消息定义
type rspMsgBattleArmyLevUp struct {
	SyncResp
	BattleArmyID        int64 `codec:"battle_army_id"`       // 战阵ID
	CurrBattleArmyLevel int64 `codec:"curr_battle_army_lev"` // 升级后显示的战阵等级
}

// BattleArmyLevUp 战阵升级／解锁协议:
func (p *Account) BattleArmyLevUp(r servers.Request) *servers.Response {
	req := new(reqMsgBattleArmyLevUp)
	rsp := new(rspMsgBattleArmyLevUp)

	initReqRsp(
		"Attr/BattleArmyLevUpRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.BattleArmyLevUpHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// BattleArmyChoiceAvatarID : 选择武将
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBattleArmyChoiceAvatarID 选择武将请求消息定义
type reqMsgBattleArmyChoiceAvatarID struct {
	Req
	BattleArmyID           int64 `codec:"battle_army_id"`            // 战阵ID
	CurrBattleArmyAvatarID int64 `codec:"curr_battle_army_avatarID"` // 选择的武将ID
}

// rspMsgBattleArmyChoiceAvatarID 选择武将回复消息定义
type rspMsgBattleArmyChoiceAvatarID struct {
	SyncResp
	BattleArmyID           int64 `codec:"battle_army_id"`            // 战阵ID
	CurrBattleArmyAvatarID int64 `codec:"curr_battle_army_avatarID"` // 选择武将后战阵武将ID
}

// BattleArmyChoiceAvatarID 选择武将:
func (p *Account) BattleArmyChoiceAvatarID(r servers.Request) *servers.Response {
	req := new(reqMsgBattleArmyChoiceAvatarID)
	rsp := new(rspMsgBattleArmyChoiceAvatarID)

	initReqRsp(
		"Attr/BattleArmyChoiceAvatarIDRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.BattleArmyChoiceAvatarIDHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
