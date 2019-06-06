package logics

import (
	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/message"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ApplyGuildInventory : 工会仓库申请
// 工会仓库申请的协议
// reqMsgApplyGuildInventory 工会仓库申请请求消息定义
type reqMsgApplyGuildInventory struct {
	Req
	ItemId string `codec:"iid"` // 物品id
}

// rspMsgApplyGuildInventory 工会仓库申请回复消息定义
type rspMsgApplyGuildInventory struct {
	SyncResp
	ResCode int64 `codec:"rescode"` // 结果码
}

// ApplyGuildInventory 工会仓库申请: 工会仓库申请的协议
func (p *Account) ApplyGuildInventory(r servers.Request) *servers.Response {
	req := new(reqMsgApplyGuildInventory)
	rsp := new(rspMsgApplyGuildInventory)

	initReqRsp(
		"Attr/ApplyGuildInventoryRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Param
	)
	// 参数检查
	if nil == gamedata.GetGuildLostInventoryCfg(req.ItemId) {
		return rpcErrorWithMsg(rsp, Err_Param, "Err_Param")
	}

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	guildUuid := p.GuildProfile.GuildUUID

	cfg := gamedata.GetGuildLostInventoryCfg(req.ItemId)
	if !p.Profile.GetSC().HasSC(gamedata.SC_GB, int64(cfg.GetPrice())) {
		logs.Error("ApplyGuildInventory SC_GB not enough")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	var ret guild.GuildRet
	ret = guild.GetModule(p.AccountID.ShardId).ApplyGuildInventoryItem(
		guildUuid,
		p.AccountID.String(),
		req.ItemId)
	if rsp := guildErrRet(ret, rsp); rsp != nil {
		return rsp
	}

	cost := &gamedata.CostData{}
	cost.AddItem(gamedata.VI_GuildBoss, cfg.GetPrice())
	if !account.CostBySync(p.Account, cost, rsp, "ApplyGuildInventory") {
		logs.Error("ApplyGuildInventory CostBySync fail")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	rsp.ResCode = int64(ret.ErrCode)
	rsp.OnChangeGuildInventory()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// ApproveGuildInventory : 工会仓库批准，拒绝
// 工会仓库批准，拒绝的协议
// reqMsgApproveGuildInventory 工会仓库批准，拒绝请求消息定义
type reqMsgApproveGuildInventory struct {
	Req
	ItemId    string `codec:"iid"`  // 物品id
	AccountId string `codec:"acid"` // acid
	TimeStamp int64  `codec:"tss"`  // 申请的时间戳
	Oper      byte   `codec:"oper"` // 操作
}

// rspMsgApproveGuildInventory 工会仓库批准，拒绝回复消息定义
type rspMsgApproveGuildInventory struct {
	SyncResp
	ResCode int64    `codec:"rescode"` // 结果码
	Acids   []string `codec:"acids"`   // acids
	Names   []string `codec:"names"`   // 姓名
	Times   []int64  `codec:"tss"`     // 申请的时间戳
	Gss     []int64  `codec:"gss"`     // 战力
}

// ApproveGuildInventory 工会仓库批准，拒绝: 工会仓库批准，拒绝的协议
func (p *Account) ApproveGuildInventory(r servers.Request) *servers.Response {
	req := new(reqMsgApproveGuildInventory)
	rsp := new(rspMsgApproveGuildInventory)

	initReqRsp(
		"Attr/ApproveGuildInventoryRsp",
		r.RawBytes,
		req, rsp, p)

	// 自己是否已经不在公会中
	if !p.GuildProfile.InGuild() {
		rsp.ResCode = int64(guild_info.Inventory_Act_Leave)
		return rpcSuccess(rsp)
	}
	guildUuid := p.GuildProfile.GuildUUID
	aggree := false
	if req.Oper > 0 {
		aggree = true
	}
	var ret guild.GuildRet
	ret, rsp.Acids, rsp.Names, rsp.Times, rsp.Gss =
		guild.GetModule(p.AccountID.ShardId).ApproveGuildInventoryItem(
			guildUuid,
			p.AccountID.String(),
			req.AccountId,
			req.ItemId, req.TimeStamp, aggree, p.GetRand())
	if rsp := guildErrRet(ret, rsp); rsp != nil {
		return rsp
	}

	rsp.ResCode = int64(ret.ErrCode)
	rsp.OnChangeGuildInventory()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetGuildInventoryApplyList : 获取工会仓库申请列表
// 获取工会仓库申请列表的协议
// reqMsgGetGuildInventoryApplyList 获取工会仓库申请列表请求消息定义
type reqMsgGetGuildInventoryApplyList struct {
	Req
	ItemId string `codec:"iid"` // 物品id
}

// rspMsgGetGuildInventoryApplyList 获取工会仓库申请列表回复消息定义
type rspMsgGetGuildInventoryApplyList struct {
	SyncResp
	ResCode int64    `codec:"rescode"` // 结果码
	Acids   []string `codec:"acids"`   // acids
	Names   []string `codec:"names"`   // 姓名
	Times   []int64  `codec:"tss"`     // 申请的时间戳
	Gss     []int64  `codec:"gss"`     // 战力
}

// GetGuildInventoryApplyList 获取工会仓库申请列表: 获取工会仓库申请列表的协议
func (p *Account) GetGuildInventoryApplyList(r servers.Request) *servers.Response {
	req := new(reqMsgGetGuildInventoryApplyList)
	rsp := new(rspMsgGetGuildInventoryApplyList)

	initReqRsp(
		"Attr/GetGuildInventoryApplyListRsp",
		r.RawBytes,
		req, rsp, p)

	// 自己是否已经不在公会中
	if !p.GuildProfile.InGuild() {
		rsp.ResCode = int64(guild_info.Inventory_Act_Leave)
		return rpcSuccess(rsp)
	}
	guildUuid := p.GuildProfile.GuildUUID

	var ret guild.GuildRet
	ret, rsp.Acids, rsp.Names, rsp.Times, rsp.Gss =
		guild.GetModule(p.AccountID.ShardId).GetApplyListGuildInventoryItem(
			guildUuid,
			p.AccountID.String(),
			req.ItemId)
	if rsp := guildErrRet(ret, rsp); rsp != nil {
		return rsp
	}

	rsp.ResCode = int64(ret.ErrCode)
	rsp.OnChangeGuildInventory()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// ExchangeGuildInventory : 工会仓库兑换
// 工会仓库兑换的协议
// reqMsgExchangeGuildInventory 工会仓库兑换请求消息定义
type reqMsgExchangeGuildInventory struct {
	Req
	ItemId string `codec:"iid"` // 物品id
}

// rspMsgExchangeGuildInventory 工会仓库兑换回复消息定义
type rspMsgExchangeGuildInventory struct {
	SyncRespWithRewards
	ResCode int64 `codec:"rescode"` // 结果码
}

// ExchangeGuildInventory 工会仓库兑换: 工会仓库兑换的协议
func (p *Account) ExchangeGuildInventory(r servers.Request) *servers.Response {
	req := new(reqMsgExchangeGuildInventory)
	rsp := new(rspMsgExchangeGuildInventory)

	initReqRsp(
		"Attr/ExchangeGuildInventoryRsp",
		r.RawBytes,
		req, rsp, p)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	guildUuid := p.GuildProfile.GuildUUID

	cfg := gamedata.GetGuildLostInventoryCfg(req.ItemId)
	if !p.Profile.GetSC().HasSC(gamedata.SC_GB, int64(cfg.GetPrice())) {
		logs.Error("ExchangeGuildInventory SC_GB not enough")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	ret := guild.GetModule(p.AccountID.ShardId).ExchangeGuildInventoryItem(
		guildUuid, p.AccountID.String(), req.ItemId, p.GetRand())
	if rsp := guildErrRet(ret, rsp); rsp != nil {
		return rsp
	}

	cost := &gamedata.CostData{}
	cost.AddItem(gamedata.VI_GuildBoss, cfg.GetPrice())
	if !account.CostBySync(p.Account, cost, rsp, "ExchangeGuildInventory") {
		logs.Error("ExchangeGuildInventory CostBySync fail")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	rsp.ResCode = int64(ret.ErrCode)
	rsp.OnChangeGuildInventory()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetGuildInventoryLog : 查看公会仓库日志
// 查看公会仓库日志协议
// reqMsgGetGuildInventoryLog 查看公会仓库日志请求消息定义
type reqMsgGetGuildInventoryLog struct {
	Req
	IsOld bool `codec:"is_old"`
}

// rspMsgGetGuildInventoryLog 查看公会仓库日志回复消息定义
type rspMsgGetGuildInventoryLog struct {
	SyncResp
	Time         []int64  `codec:"t"`      // 时间
	AssignerName []string `codec:"asgnnm"` // 分配者姓名
	ShowItemId   []string `codec:"showim"` // 仓库物品id
	ReceiverName []string `codec:"revnm"`  // 接受者姓名
}

// GetGuildInventoryLog 查看公会仓库日志: 查看公会仓库日志协议
func (p *Account) GetGuildInventoryLog(r servers.Request) *servers.Response {
	req := new(reqMsgGetGuildInventoryLog)
	rsp := new(rspMsgGetGuildInventoryLog)

	initReqRsp(
		"Guild/GetGuildInventoryLogRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Load_Msg
	)

	warnCode := p.CheckGuildStatus(false)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	guildUuid := p.GuildProfile.GuildUUID
	var msgs []message.PlayerMsg
	var err error
	if req.IsOld {
		msgs, err = message.LoadPlayerMsgs(guildUuid,
			guild_info.GuildLostInventoryMsgTableKey, guild_info.GuildInventoryMsgCount)
	} else {
		msgs, err = message.LoadPlayerMsgs(guildUuid,
			guild_info.GuildInventoryMsgTableKey, guild_info.GuildInventoryMsgCount)
	}

	if err != nil {
		logs.Error("GetGuildInventoryLog LoadPlayerMsgs err %s", err.Error())
		return rpcErrorWithMsg(rsp, Err_Load_Msg, "Err_Load_Msg")
	}

	recs := make([]guild.AssignInventroyRecord, 0, guild_info.GuildInventoryMsgCount)
	for _, msg := range msgs {
		m := guild.AssignInventroyRecord{}
		err := json.Unmarshal([]byte(msg.Params[0]), &m)
		if err != nil {
			continue
		}
		recs = append(recs, m)
	}

	rsp.Time = make([]int64, 0, len(recs))
	rsp.AssignerName = make([]string, 0, len(recs))
	rsp.ShowItemId = make([]string, 0, len(recs))
	rsp.ReceiverName = make([]string, 0, len(recs))
	for _, r := range recs {
		rsp.Time = append(rsp.Time, r.Time)
		rsp.AssignerName = append(rsp.AssignerName, r.AssignerName)
		rsp.ShowItemId = append(rsp.ShowItemId, r.ItemId)
		rsp.ReceiverName = append(rsp.ReceiverName, r.ReceiverName)
	}

	return rpcSuccess(rsp)
}

// old老版本用
// AssignGuildInventory : 分配公会仓库物品
// 分配公会仓库物品协议
// reqMsgAssignGuildInventory 分配公会仓库物品请求消息定义
type reqMsgAssignGuildInventory struct {
	Req
	ItemId    string `codec:"iid"`  // 物品id
	AccountId string `codec:"acid"` // 公会成员acid
}

// rspMsgAssignGuildInventory 分配公会仓库物品回复消息定义
type rspMsgAssignGuildInventory struct {
	SyncResp
	ResCode int64 `codec:"rescode"` // 结果码, 0:成功,1:物品无,2:此人此物品剩余次数不足,3:职位变动,4:被分配玩家已被踢出军团,5:主分配者不在工会了
}

// AssignGuildInventory 分配公会仓库物品: 分配公会仓库物品协议
func (p *Account) AssignGuildInventory(r servers.Request) *servers.Response {
	req := new(reqMsgAssignGuildInventory)
	rsp := new(rspMsgAssignGuildInventory)

	initReqRsp(
		"Guild/AssignGuildInventoryRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Item_NotExist
	)
	if nil == gamedata.GetGuildInventoryCfg(req.ItemId) {
		return rpcErrorWithMsg(rsp, Err_Item_NotExist, "Err_Item_NotExist")
	}

	// 自己是否已经不在公会中
	if !p.GuildProfile.InGuild() {
		rsp.ResCode = int64(guild_info.Inventory_Act_Leave)
		return rpcSuccess(rsp)
	}
	guildUuid := p.GuildProfile.GuildUUID

	ret := guild.GetModule(p.AccountID.ShardId).AssignGuildInventoryItem(guildUuid, p.AccountID.String(),
		req.AccountId, req.ItemId, p.GetRand())
	if rsp := guildErrRet(ret, rsp); rsp != nil {
		return rsp
	}

	rsp.ResCode = int64(ret.ErrCode)
	rsp.OnChangeGuildInventory()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
