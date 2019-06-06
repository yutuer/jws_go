package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// TBBattleStart : 组队BOSS战开始
// 组队BOSS战开始

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBBattleStart 组队BOSS战开始请求消息定义
type reqMsgTBBattleStart struct {
	Req
	TBBattleSTeamId string `codec:"b_bs_tid"` // 开始战斗的teamId
}

// rspMsgTBBattleStart 组队BOSS战开始回复消息定义
type rspMsgTBBattleStart struct {
	SyncResp
	TBBattleServUrl string `codec:"b_bs_su"` // 开始战斗的服务器url
	TBBattleSGlobalTeamId string `codec:"b_bs_gtid"` // 开始战斗的GloablteamId
}

// TBBattleStart 组队BOSS战开始: 组队BOSS战开始
func (p *Account) TBBattleStart(r servers.Request) *servers.Response {
	req := new(reqMsgTBBattleStart)
	rsp := new(rspMsgTBBattleStart)

	initReqRsp(
		"Attr/TBBattleStartRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBBattleStartHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// TBBattleEnd : 组队BOSS战结束
// 组队BOSS战结束

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBBattleEnd 组队BOSS战结束请求消息定义
type reqMsgTBBattleEnd struct {
	Req
	TBBattleETeamId string `codec:"b_bs_tid"` // 结束战斗的teamId
	TBBattleEGlobalTeamId string `codec:"b_be_gtid"` // 结束战斗的GloablteamId
}

// rspMsgTBBattleEnd 组队BOSS战结束回复消息定义
type rspMsgTBBattleEnd struct {
	SyncRespWithRewards
	TBBoxIsFull bool `codec:"b_bb_if"` // 仓库中宝箱是否满
}

// TBBattleEnd 组队BOSS战结束: 组队BOSS战结束
func (p *Account) TBBattleEnd(r servers.Request) *servers.Response {
	req := new(reqMsgTBBattleEnd)
	rsp := new(rspMsgTBBattleEnd)

	initReqRsp(
		"Attr/TBBattleEndRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBBattleEndHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// TBOpenStorage : 打开组队BOSS仓库
// 打开组队BOSS仓库

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBOpenStorage 打开组队BOSS仓库请求消息定义
type reqMsgTBOpenStorage struct {
	Req
}

// rspMsgTBOpenStorage 打开组队BOSS仓库回复消息定义
type rspMsgTBOpenStorage struct {
	SyncResp
	TBBoxInfo [][]byte `codec:"b_in"` // 仓库中宝箱信息
	TBBoxHCOpenTimes int64 `codec:"b_op_t"` // 仓库中宝箱用钻石已开次数
}

// TBOpenStorage 打开组队BOSS仓库: 打开组队BOSS仓库
func (p *Account) TBOpenStorage(r servers.Request) *servers.Response {
	req := new(reqMsgTBOpenStorage)
	rsp := new(rspMsgTBOpenStorage)

	initReqRsp(
		"Attr/TBOpenStorageRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBOpenStorageHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// TBBoxInfo 打开组队BOSS仓库
type TBBoxInfo struct {
	
	TBBoxId string `codec:"b_id"` // 仓库中宝箱id
	TBBoxPos int64 `codec:"b_po"` // 仓库中宝箱的位置
	TBBoxEndTime int64 `codec:"b_e_t"` // 仓库中宝箱到期截止时间
}

// TBOpenBox : 打开组队BOSS宝箱
// 打开组队BOSS宝箱

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBOpenBox 打开组队BOSS宝箱请求消息定义
type reqMsgTBOpenBox struct {
	Req
	TBBoxOpIndex int64 `codec:"op_b_in"` // 打开箱子的位置
	TBBoxOpType int64 `codec:"op_b_tp"` // 打开宝箱的方式 0等时间 1花钻石
}

// rspMsgTBOpenBox 打开组队BOSS宝箱回复消息定义
type rspMsgTBOpenBox struct {
	SyncRespWithRewards
	TBBoxInfo [][]byte `codec:"b_in"` // 仓库中宝箱信息
	TBBoxHCOpenTimes int64 `codec:"b_op_t"` // 仓库中宝箱用钻石已开次数
}

// TBOpenBox 打开组队BOSS宝箱: 打开组队BOSS宝箱
func (p *Account) TBOpenBox(r servers.Request) *servers.Response {
	req := new(reqMsgTBOpenBox)
	rsp := new(rspMsgTBOpenBox)

	initReqRsp(
		"Attr/TBOpenBoxRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBOpenBoxHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// TBDelBox : 删除组队BOSS宝箱
// 删除组队BOSS宝箱

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBDelBox 删除组队BOSS宝箱请求消息定义
type reqMsgTBDelBox struct {
	Req
	TBBoxPos int64 `codec:"de_b_in"` // 打开箱子的位置下标
}

// rspMsgTBDelBox 删除组队BOSS宝箱回复消息定义
type rspMsgTBDelBox struct {
	SyncResp
	TBBoxInfo [][]byte `codec:"b_in"` // 仓库中宝箱信息
	TBBoxHCOpenTimes int64 `codec:"b_op_t"` // 仓库中宝箱用钻石已开次数
}

// TBDelBox 删除组队BOSS宝箱: 删除组队BOSS宝箱
func (p *Account) TBDelBox(r servers.Request) *servers.Response {
	req := new(reqMsgTBDelBox)
	rsp := new(rspMsgTBDelBox)

	initReqRsp(
		"Attr/TBDelBoxRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBDelBoxHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


