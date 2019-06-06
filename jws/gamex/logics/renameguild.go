package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// RenameGuild : 军团重命名
// 只有军团长有此操作权限

// reqMsgRenameGuild 军团重命名请求消息定义
type reqMsgRenameGuild struct {
	Req
	NewName string `codec:"new_name"` // 军团新名字
}

// rspMsgRenameGuild 军团重命名回复消息定义
type rspMsgRenameGuild struct {
	SyncResp
}

// RenameGuild 军团重命名: 只有军团长有此操作权限
func (p *Account) RenameGuild(r servers.Request) *servers.Response {
	req := new(reqMsgRenameGuild)
	rsp := new(rspMsgRenameGuild)

	initReqRsp(
		"Attr/RenameGuildRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	if p.GuildProfile.GuildUUID == "" {
		return rpcWarn(rsp, errCode.GuildPlayerNotFound)
	}

	powerData := gamedata.GetGuildPosData(p.GuildProfile.GuildPosition)
	logs.Debug("rename guild config, %v", powerData)
	if powerData.GetReNamePower() <= 0 {
		logs.Debug("failed to rename guild, %d, %d", p.GuildProfile.GuildPosition, powerData.GetReNamePower())
		return rpcWarn(rsp, errCode.GuildPositionErr)
	}

	sid := p.AccountID.ShardId
	info, ret := guild.GetModule(sid).GetGuildInfo(p.GuildProfile.GuildUUID)
	if resp := guildErrRet(ret, rsp); resp != nil {
		return resp
	}
	cost := gamedata.GetRenameGuildCost(info.Base.RenameTimes + 1)
	if !p.Profile.GetHC().HasHC(cost) {
		return rpcWarn(rsp, errCode.MaterialNotEnough)
	}

	guildID := p.GuildProfile.GuildUUID
	newName := req.NewName
	ret = guild.GetModule(sid).RenameGuild(guildID, newName, p.Profile.Name)
	if resp := guildErrRet(ret, rsp); resp != nil {
		return resp
	}
	gvg.GetModule(sid).RenameGuild(guildID, newName)
	p.GuildProfile.UpdateName(req.NewName)

	data := &gamedata.CostData{}
	data.AddItem(helper.VI_Hc, uint32(cost))
	if !account.CostBySync(p.Account, data, rsp, "RenameGuild") {
		logs.Error("fatal error: no enough hc")
	}
	// logic imp end
	rsp.OnChangeGuildInfo()

	//通知其他需要刷新缓存的模块
	csrob.GetModule(sid).GuildMod.Rename(guildID)

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
