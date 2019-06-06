package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) checkGuildForGVG(rsp respInterface, checkRejoin bool) *servers.Response {
	if errCode := p.CheckGuildStatus(checkRejoin); errCode > 0 {
		gvg.GetModule(p.AccountID.ShardId).CommandExecAsync(gvg.GVGCmd{
			Typ:  gvg.Cmd_Typ_Remove_Player,
			AcID: p.AccountID.String(),
		})
		return rpcWarn(rsp, uint32(errCode))
	}
	return nil
}

// checkRejoin = false: 忽略检查加入公会的CD(同一天换工会)
func (p *Account) CheckGuildStatus(checkRejoin bool) uint32 {
	if !p.GuildProfile.InGuild() {
		return uint32(errCode.GuildPlayerNotIn)
	}
	// 2.4.0后不再检查该选项
	//if checkRejoin && p.GuildProfile.IsLimitedByRejoinGuild(p.GetProfileNowTime()) {
	//	return uint32(errCode.ChangeGuildInSameDay)
	//}
	return 0
}

// GVGEnterCity : GVG军团战进入城市协议
// GVG军团战进入某个城市

// reqMsgGVGEnterCity GVG军团战进入城市协议请求消息定义
type reqMsgGVGEnterCity struct {
	Req
	CityID int64 `codec:"city_id"` // 城市ID
}

// rspMsgGVGEnterCity GVG军团战进入城市协议回复消息定义
type rspMsgGVGEnterCity struct {
	SyncResp
}

// GVGEnterCity GVG军团战进入城市协议: GVG军团战进入某个城市
func (p *Account) GVGEnterCity(r servers.Request) *servers.Response {
	req := new(reqMsgGVGEnterCity)
	rsp := new(rspMsgGVGEnterCity)

	initReqRsp(
		"Attr/GVGEnterCityRsp",
		r.RawBytes,
		req, rsp, p)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, true); errRsp != nil {
		return errRsp
	}

	info := [gvg.GVG_AVATAR_COUNT]helper.AvatarState{}
	playerDetail := [gvg.GVG_AVATAR_COUNT]*helper.Avatar2Client{}
	heroTeam := p.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_GVG)
	if len(heroTeam) != 3 {
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	for i, avatar := range heroTeam {
		info[i].HP = 1
		info[i].MP = 0.5
		info[i].WS = 0
		info[i].Avatar = int(avatar)
		a := &helper.Avatar2Client{}
		err := account.FromAccount(a, p.Account, avatar)
		if err != nil {
			logs.Warn("account.FromAccount Error by %v", err)
			playerDetail = [gvg.GVG_AVATAR_COUNT]*helper.Avatar2Client{}
			break
		}
		playerDetail[i] = a
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:          gvg.Cmd_Typ_EnterCity,
		AcID:         p.AccountID.String(),
		Name:         p.Profile.Name,
		GuID:         p.GuildProfile.GuildUUID,
		GuName:       p.GuildProfile.GuildName,
		Title:        p.Profile.GetTitle().TitleTakeOn,
		Players:      info,
		CityID:       int(req.CityID),
		DestinySkill: [helper.DestinyGeneralSkillMax]int{-1, -1, -1},
		DetailInfo:   playerDetail,
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	p.Tmp.GVGCity = int(req.CityID)

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GVGMatchEnemy : GVG军团战开始匹配
// GVG军团战开始匹配

// reqMsgGVGMatchEnemy GVG军团战开始匹配请求消息定义
type reqMsgGVGMatchEnemy struct {
	Req
	CityID int64 `codec:"city_id"` // 城市ID
}

// rspMsgGVGMatchEnemy GVG军团战开始匹配回复消息定义
type rspMsgGVGMatchEnemy struct {
	SyncResp
}

// GVGMatchEnemy GVG军团战开始匹配: GVG军团战开始匹配
func (p *Account) GVGMatchEnemy(r servers.Request) *servers.Response {
	req := new(reqMsgGVGMatchEnemy)
	rsp := new(rspMsgGVGMatchEnemy)

	initReqRsp(
		"Attr/GVGMatchEnemyRsp",
		r.RawBytes,
		req, rsp, p)

	// 清除信息
	p.Tmp.CleanGVGData()
	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, true); errRsp != nil {
		return errRsp
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_PrepareFight,
		CityID: int(req.CityID),
		AcID:   p.AccountID.String(),
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GVGMatchQuery : GVG军团战轮询结果
// GVG军团战轮询匹配结果

// reqMsgGVGMatchQuery GVG军团战轮询结果请求消息定义
type reqMsgGVGMatchQuery struct {
	Req
	CityID int64 `codec:"city_id"` // 城市ID
}

// rspMsgGVGMatchQuery GVG军团战轮询结果回复消息定义
type rspMsgGVGMatchQuery struct {
	SyncResp

	AvatarHP     []float32 `codec:"avatar_hp"`
	AvatarMP     []float32 `codec:"avatar_mp"`
	AvatarWS     []float32 `codec:"avatar_ws"`
	DestinySkill []int64   `codec:"destiny_skill"`

	EnemyInfo     [][]byte  `codec:"enemy_info"`
	EAvatarHP     []float32 `codec:"eavatar_hp"`
	EAvatarMP     []float32 `codec:"eavatar_mp"`
	EAvatarWS     []float32 `codec:"eavatar_ws"`
	EDestinySkill []int64   `codec:"edestiny_skill`
	IsRobot       bool      `codec:"is_robot"`
	URL           string    `codec:"url"`
	RoomID        string    `codec:"room_id"`
}

// foolish
// GVGMatchQuery GVG军团战轮询结果: GVG军团战轮询匹配结果，若Rsp中is_success返回为true，证明匹配成功
func (p *Account) GVGMatchQuery(r servers.Request) *servers.Response {
	req := new(reqMsgGVGMatchQuery)
	rsp := new(rspMsgGVGMatchQuery)

	initReqRsp(
		"Attr/GVGMatchQueryRsp",
		r.RawBytes,
		req, rsp, p)
	const (
		_ = iota
		Err_LoadAccount
		Err_ParseAccountID
		Err_LoadAvatar
	)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, true); errRsp != nil {
		return errRsp
	}

	gvgData := p.Tmp.GetGVGData()
	if gvgData.GvgEnemyAcID == "" {
		return rpcSuccess(rsp)
	}
	rsp.EnemyInfo = make([][]byte, helper.GVG_AVATAR_COUNT, helper.GVG_AVATAR_COUNT)
	rsp.AvatarHP = make([]float32, helper.GVG_AVATAR_COUNT, helper.GVG_AVATAR_COUNT)
	rsp.AvatarMP = make([]float32, helper.GVG_AVATAR_COUNT, helper.GVG_AVATAR_COUNT)
	rsp.AvatarWS = make([]float32, helper.GVG_AVATAR_COUNT, helper.GVG_AVATAR_COUNT)
	rsp.EAvatarHP = make([]float32, helper.GVG_AVATAR_COUNT, helper.GVG_AVATAR_COUNT)
	rsp.EAvatarMP = make([]float32, helper.GVG_AVATAR_COUNT, helper.GVG_AVATAR_COUNT)
	rsp.EAvatarWS = make([]float32, helper.GVG_AVATAR_COUNT, helper.GVG_AVATAR_COUNT)
	rsp.DestinySkill = make([]int64, 0, helper.DestinyGeneralSkillMax)
	rsp.EDestinySkill = make([]int64, 0, helper.DestinyGeneralSkillMax)
	for i, item := range gvgData.GvgEAvatarState {
		rsp.EAvatarHP[i] = item.HP
		rsp.EAvatarMP[i] = item.MP
		rsp.EAvatarWS[i] = item.WS
	}
	for i, item := range gvgData.GvgAvatarState {
		rsp.AvatarHP[i] = item.HP
		rsp.AvatarMP[i] = item.MP
		rsp.AvatarWS[i] = item.WS
	}

	for _, item := range gvgData.GvgDestinySkill {
		if item != -1 {
			rsp.DestinySkill = append(rsp.DestinySkill, int64(item))
		}
	}
	for _, item := range gvgData.GvgEDestinySkill {
		if item != -1 {
			rsp.EDestinySkill = append(rsp.DestinySkill, int64(item))
		}
	}

	var enemyData [helper.GVG_AVATAR_COUNT]*helper.Avatar2Client
	if gvgData.GvgEnemyAcID == "0:0:RobotID" {
		for i, _ := range gvgData.GvgEAvatarState {
			droid := gamedata.GetDroidForGVG(uint32(p.Profile.GetCorp().Level))
			a := &helper.Avatar2Client{}
			// 主将ID为0、1、2的机器人
			account.FromAccountByDroid(a, droid, i)
			enemyData[i] = a
		}
		rsp.IsRobot = true
	} else {
		if gvgData.GvgEnemyData != [helper.GVG_AVATAR_COUNT]*helper.Avatar2Client{} {
			// 有可能阵容顺序发生变化，按照指定顺序返给客户端
			for i, item := range gvgData.GvgEAvatarState {
				for _, data := range gvgData.GvgEnemyData {
					if data.AvatarId == item.Avatar {
						enemyData[i] = data
						continue
					}
				}
			}

		} else {
			dbAccountID, err := db.ParseAccount(gvgData.GvgEnemyAcID)
			if err != nil {
				return rpcErrorWithMsg(rsp, Err_ParseAccountID, err.Error())
			}

			enemy_account, err := account.LoadPvPAccount(dbAccountID)
			if err != nil {
				return rpcErrorWithMsg(rsp, Err_LoadAccount, err.Error())
			}
			for i, item := range gvgData.GvgEAvatarState {
				a := &helper.Avatar2Client{}
				err := account.FromAccount(a, enemy_account, item.Avatar)
				if err != nil {
					return rpcErrorWithMsg(rsp, Err_LoadAvatar,
						fmt.Sprintf("load Account Error by %s, ID: %d", item.Avatar, err.Error()))
				}
				enemyData[i] = a
			}
		}
	}

	avatarId := []int{gvgData.GvgAvatarState[0].Avatar, gvgData.GvgAvatarState[1].Avatar, gvgData.GvgAvatarState[2].Avatar}
	eavatarId := []int{gvgData.GvgEAvatarState[0].Avatar, gvgData.GvgEAvatarState[1].Avatar, gvgData.GvgEAvatarState[2].Avatar}
	logiclog.LogGvGstartFight(p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, int(req.CityID), p.Profile.GetData().CorpCurrGS, avatarId,
		gvgData.GvgEnemyAcID, int(req.CityID), enemyData[0].CorpGs, enemyData[0].CorpLv, eavatarId,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	for i, data := range enemyData {
		rsp.EnemyInfo[i] = encode(data)
	}
	rsp.RoomID = gvgData.RoomID
	rsp.URL = gvgData.Url
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GVGEndFight : GVG军团战战斗结束提交战果
// GVG军团战提交战果

// reqMsgGVGEndFight GVG军团战战斗结束提交战果请求消息定义
type reqMsgGVGEndFight struct {
	ReqWithAnticheat
	CityID       int64     `codec:"city_id"` // 城市ID
	IsWin        int64     `codec:"is_win"`  // 是否胜利
	AvatarHP     []float32 `codec:"avatar_hp"`
	AvatarMP     []float32 `codec:"avatar_mp"`
	AvatarWS     []float32 `codec:"avatar_ws"`
	AvatarID     []int64   `codec:"avatar_id"`
	DestinySkill []int64   `codec:"destiny_skill"`
	RoomID       string    `codec:"room_id"`
}

// rspMsgGVGEndFight GVG军团战战斗结束提交战果回复消息定义
type rspMsgGVGEndFight struct {
	SyncRespWithRewardsAnticheat
	Score     int64 `codec:"score"`      // 总分
	WinStreak int64 `codec:"win_streak"` // 连胜次数
}

// GVGEndFight GVG军团战战斗结束提交战果: GVG军团战提交战果
func (p *Account) GVGEndFight(r servers.Request) *servers.Response {
	req := new(reqMsgGVGEndFight)
	rsp := new(rspMsgGVGEndFight)

	initReqRsp(
		"Attr/GVGEndFightRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	// 反作弊检查
	avatar := make([]int, 0, 3)
	for _, id := range req.AvatarID {
		avatar = append(avatar, int(id))
	}
	if cheatRsp := p.AntiCheatCheck(&rsp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0, account.Anticheat_Typ_GVG); cheatRsp != nil {
		return cheatRsp
	}

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, true); errRsp != nil {
		return errRsp
	}
	info := [gvg.GVG_AVATAR_COUNT]helper.AvatarState{}
	if req.IsWin > 0 {
		for i, avatar := range req.AvatarID {
			info[i].HP = req.AvatarHP[i]
			info[i].MP = req.AvatarMP[i]
			info[i].WS = req.AvatarWS[i]
			info[i].Avatar = int(avatar)
		}
	} else {
		for i, avatar := range req.AvatarID {
			info[i].HP = 1
			info[i].MP = 0.5
			info[i].WS = 0
			info[i].Avatar = int(avatar)
		}
	}

	destinySkill := [helper.DestinyGeneralSkillMax]int{-1, -1, -1}
	if req.IsWin > 0 {
		for i, item := range req.DestinySkill {
			destinySkill[i] = int(item)
		}
	}
	var guildMember int
	guildData, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if guildData == nil {
		guildMember = 0
		logs.Warn("gvg_logiclgo err by %v", code.ErrMsg)
	} else {
		guildMember = guildData.Base.MemNum
	}
	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		AcID:         p.AccountID.String(),
		Typ:          gvg.Cmd_Typ_EndFight,
		IsWin:        req.IsWin > 0,
		CityID:       int(req.CityID),
		Players:      info,
		DestinySkill: destinySkill,
		GuID:         p.GuildProfile.GuildUUID,
		GuMember:     guildMember,
		Name:         p.Profile.Name,
		GuName:       p.GuildProfile.GuildName,
		Title:        p.Profile.GetTitle().TitleTakeOn,
		RoomID:       req.RoomID,
	})
	logs.Debug("Logic PlayerInfo: %v", info)
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	if ret.WinStreakCount >= int(gamedata.GVGWinStreakNotice().GetLampValueIP1()) {
		logs.Trace("GVG SysNotice")
		sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_GVG_MARQUEE_CONSECUTIVE_VICTORY).
			AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
			AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", ret.WinStreakCount)).
			Send()
	}
	rsp.Score = int64(ret.Score)
	rsp.WinStreak = int64(ret.WinStreakCount)

	// logic imp end
	avatarId := []int{info[0].Avatar, info[1].Avatar, info[2].Avatar}
	logiclog.LogGvGFinishFight(p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, int(req.CityID), p.Profile.GetData().CorpCurrGS, avatarId, int(ret.WinStreakCount),
		int32(req.IsWin),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	// 清除信息
	p.Tmp.CleanGVGData()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GVGCancelMatch : GVG军团战取消匹配
//

// reqMsgGVGCancelMatch GVG军团战取消匹配请求消息定义
type reqMsgGVGCancelMatch struct {
	Req
	CityID int64 `codec:"city_id"` // 城市ID
}

// rspMsgGVGCancelMatch GVG军团战取消匹配回复消息定义
type rspMsgGVGCancelMatch struct {
	SyncResp
}

// GVGCancelMatch GVG军团战取消匹配:
func (p *Account) GVGCancelMatch(r servers.Request) *servers.Response {
	req := new(reqMsgGVGCancelMatch)
	rsp := new(rspMsgGVGCancelMatch)

	initReqRsp(
		"Attr/GVGCancelMatchRsp",
		r.RawBytes,
		req, rsp, p)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, true); errRsp != nil {
		return errRsp
	}
	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		AcID:   p.AccountID.String(),
		Typ:    gvg.Cmd_Typ_CancelMatch,
		CityID: int(req.CityID),
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	// 清除信息
	p.Tmp.CleanGVGData()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GVGLeaveCity : GVG军团战玩家离开
//
// reqMsgGVGLeaveCity GVG军团战玩家离开请求消息定义
type reqMsgGVGLeaveCity struct {
	Req
	CityID int64 `codec:"city_id"` // 城市ID
}

// rspMsgGVGLeaveCity GVG军团战玩家离开回复消息定义
type rspMsgGVGLeaveCity struct {
	SyncResp
}

// GVGLeaveCity GVG军团战玩家离开:
func (p *Account) GVGLeaveCity(r servers.Request) *servers.Response {
	req := new(reqMsgGVGLeaveCity)
	rsp := new(rspMsgGVGLeaveCity)

	initReqRsp(
		"Attr/GVGLeaveCityRsp",
		r.RawBytes,
		req, rsp, p)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, true); errRsp != nil {
		return errRsp
	}
	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_LeaveCity,
		AcID:   p.AccountID.String(),
		CityID: int(req.CityID),
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GVGGetGuildInfo : GVG军团战查询某个城市所有工会的分数
//

// reqMsgGVGGetGuildInfo GVG军团战查询某个城市所有工会的分数请求消息定义
type reqMsgGVGGetGuildInfo struct {
	Req
	CityID int64 `codec:"city_id"` // 城市ID
}

// rspMsgGVGGetGuildInfo GVG军团战查询某个城市所有工会的分数回复消息定义
type rspMsgGVGGetGuildInfo struct {
	SyncResp
	GuildName  []string `codec:"guild_name"`  // 工会成员的名字
	GuildScore []int64  `codec:"guild_score"` // 工会成员的分数
	NowMatch   int64    `codec:"now_match"`   //当前正在匹配的敌人
}

// GVGGetGuildInfo GVG军团战查询某个城市所有工会的分数:
func (p *Account) GVGGetGuildInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGVGGetGuildInfo)
	rsp := new(rspMsgGVGGetGuildInfo)

	initReqRsp(
		"Attr/GVGGetGuildInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, false); errRsp != nil {
		return errRsp
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_Get_GuildRank,
		AcID:   p.AccountID.String(),
		CityID: int(req.CityID),
		GuID:   p.GuildProfile.GuildUUID,
		GuName: p.GuildProfile.GuildName,
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	if ret.SortItem == nil {
		return rpcSuccess(rsp)
	}
	rsp.NowMatch = int64(ret.NowMatch)
	l := len(ret.SortItem)
	rsp.GuildName = make([]string, 0, l)
	rsp.GuildScore = make([]int64, 0, l)
	for _, item := range ret.SortItem {
		rsp.GuildName = append(rsp.GuildName, item.StrKey)
		rsp.GuildScore = append(rsp.GuildScore, item.IntVal)
	}
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GVGGetGuildMemberInfo : GVG军团战查询某个城市的自己所在工会的工会成员信息
//

// reqMsgGVGGetGuildMemberInfo GVG军团战查询某个城市的自己所在工会的工会成员信息请求消息定义
type reqMsgGVGGetGuildMemberInfo struct {
	Req
	CityID int64 `codec:"city_id"` // 城市ID
}

// rspMsgGVGGetGuildMemberInfo GVG军团战查询某个城市的自己所在工会的工会成员信息回复消息定义
type rspMsgGVGGetGuildMemberInfo struct {
	SyncResp
	MemberName  []string `codec:"member_name"`  // 工会成员的名字
	MemberScore []int64  `codec:"member_score"` // 工会成员的分数
	NowMatch    int64    `codec:"now_match"`    //当前正在匹配的敌人
}

// GVGGetGuildMemberInfo GVG军团战查询某个城市的自己所在工会的工会成员信息:
func (p *Account) GVGGetGuildMemberInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGVGGetGuildMemberInfo)
	rsp := new(rspMsgGVGGetGuildMemberInfo)

	initReqRsp(
		"Attr/GVGGetGuildMemberInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, false); errRsp != nil {
		return errRsp
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_Get_SelfGuildInfo,
		AcID:   p.AccountID.String(),
		CityID: int(req.CityID),
		GuID:   p.GuildProfile.GuildUUID,
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	if ret.SortItem == nil {
		return rpcSuccess(rsp)
	}
	rsp.NowMatch = int64(ret.NowMatch)
	l := len(ret.SortItem)
	rsp.MemberName = make([]string, 0, l)
	rsp.MemberScore = make([]int64, 0, l)
	for _, item := range ret.SortItem {
		rsp.MemberName = append(rsp.MemberName, item.StrKey)
		rsp.MemberScore = append(rsp.MemberScore, item.IntVal)
	}

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GVGGetPlayerInfo : GVG军团战查询某个城市所有存在玩家的信息
//

// reqMsgGVGGetPlayerInfo GVG军团战查询某个城市所有存在玩家的信息请求消息定义
type reqMsgGVGGetPlayerInfo struct {
	Req
	CityID int64 `codec:"city_id"` // 城市ID
}

// rspMsgGVGGetPlayerInfo GVG军团战查询某个城市所有存在玩家的信息回复消息定义
type rspMsgGVGGetPlayerInfo struct {
	SyncResp
	PlayerName  []string `codec:"player_name"`  // 玩家的名字
	PlayerGuild []string `codec:"player_guild"` // 玩家的分数
	NowMatch    int64    `codec:"now_match"`    //当前正在匹配的敌人
}

// GVGGetPlayerInfo GVG军团战查询某个城市所有存在玩家的信息:
func (p *Account) GVGGetPlayerInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGVGGetPlayerInfo)
	rsp := new(rspMsgGVGGetPlayerInfo)

	initReqRsp(
		"Attr/GVGGetPlayerInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, false); errRsp != nil {
		return errRsp
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_Get_PlayerInfo,
		AcID:   p.AccountID.String(),
		CityID: int(req.CityID),
		GuID:   p.GuildProfile.GuildUUID,
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	if ret.SortItem == nil {
		return rpcSuccess(rsp)
	}
	rsp.NowMatch = int64(ret.NowMatch)
	l := len(ret.SortItem)
	rsp.PlayerName = make([]string, 0, l)
	rsp.PlayerGuild = make([]string, 0, l)
	for _, item := range ret.SortItem {
		rsp.PlayerName = append(rsp.PlayerName, item.StrKey)
		rsp.PlayerGuild = append(rsp.PlayerGuild, item.StrVal)
	}

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetGVGCityData : 获取当前工会成员的积分排名
// 请求获取成员积分排名

// reqMsgGetGVGCityData 获取当前工会成员的积分排名请求消息定义
type reqMsgGetGVGCityData struct {
	Req
	CityID int64 `codec:"cityid"` // 选择城池id
}

// rspMsgGetGVGCityData 获取当前工会成员的积分排名回复消息定义
type rspMsgGetGVGCityData struct {
	SyncResp
	Names          []string `codec:"names"`          // 军团名字数组
	Scores         []int64  `codec:"scores"`         // 军团积分
	SelfGuildRank  int64    `codec:"selfguildrank"`  // 自己军团的排名
	SelfGuildScore int64    `codec:"selfguildscore"` // 自己的积分
}

// GetGVGCityData 获取当前工会成员的积分排名: 请求获取成员积分排名
func (p *Account) GetGVGCityData(r servers.Request) *servers.Response {
	req := new(reqMsgGetGVGCityData)
	rsp := new(rspMsgGetGVGCityData)

	initReqRsp(
		"Attr/GetGVGCityDataRsp",
		r.RawBytes,
		req, rsp, p)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, false); errRsp != nil {
		return errRsp
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_Get_GuildRank,
		AcID:   p.AccountID.String(),
		CityID: int(req.CityID),
		GuID:   p.GuildProfile.GuildUUID,
		GuName: p.GuildProfile.GuildName,
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	if ret.SortItem == nil {
		return rpcSuccess(rsp)
	}
	l := len(ret.SortItem)
	rsp.Names = make([]string, 0, l)
	rsp.Scores = make([]int64, 0, l)
	for _, item := range ret.SortItem {
		rsp.Names = append(rsp.Names, item.StrKey)
		rsp.Scores = append(rsp.Scores, item.IntVal)
	}
	rsp.SelfGuildRank = int64(ret.Rank)
	rsp.SelfGuildScore = int64(ret.Score)

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetGVGGuildScore : 获取当前工会成员的积分排名
// 请求获取成员积分排名

// reqMsgGetGVGGuildScore 获取当前工会成员的积分排名请求消息定义
type reqMsgGetGVGGuildScore struct {
	Req
}

// rspMsgGetGVGGuildScore 获取当前工会成员的积分排名回复消息定义
type rspMsgGetGVGGuildScore struct {
	SyncResp
	Names       []string `codec:"names"`       // 工会成员名字数组
	Scores      []int64  `codec:"scores"`      // 工会成员积分
	SelfRank    int64    `codec:"selfrank"`    // 自己的排名
	SelfName    string   `codec:"selfname"`    // 自己的名字
	SelfScore   int64    `codec:"selfscore"`   // 自己的积分
	SelfJoinGVG int64    `codec:"selfjoingvg"` // 自己是否参战
}

// GetGVGGuildScore 获取当前工会成员的积分排名: 请求获取成员积分排名
func (p *Account) GetGVGGuildScore(r servers.Request) *servers.Response {
	req := new(reqMsgGetGVGGuildScore)
	rsp := new(rspMsgGetGVGGuildScore)

	initReqRsp(
		"Attr/GetGVGGuildScoreRsp",
		r.RawBytes,
		req, rsp, p)

	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, false); errRsp != nil {
		return errRsp
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_Get_SelfGuildAllInfo,
		AcID:   p.AccountID.String(),
		GuID:   p.GuildProfile.GuildUUID,
		GuName: p.GuildProfile.GuildName,
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	if ret.SortItem == nil {
		return rpcSuccess(rsp)
	}
	l := len(ret.SortItem)
	rsp.Names = make([]string, 0, l)
	rsp.Scores = make([]int64, 0, l)
	for i, item := range ret.SortItem {
		rsp.Names = append(rsp.Names, item.StrKey)
		rsp.Scores = append(rsp.Scores, item.IntVal)
		if item.StrKey == p.Profile.Name {
			rsp.SelfName = item.StrKey
			rsp.SelfScore = item.IntVal
			rsp.SelfRank = int64(i) + 1
			rsp.SelfJoinGVG = 1
		}
	}

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetGVGGuildTotalScore : 获取所有军团所有玩家的总积分排名
// 请求获取成员积分排名

// reqMsgGetGVGGuildTotalScore 获取所有军团所有玩家的总积分排名请求消息定义
type reqMsgGetGVGGuildTotalScore struct {
	Req
}

// rspMsgGetGVGGuildTotalScore 获取所有军团所有玩家的总积分排名回复消息定义
type rspMsgGetGVGGuildTotalScore struct {
	SyncResp
	Names       []string `codec:"names"`       // 玩家名字数组
	Scores      []int64  `codec:"scores"`      // 玩家积分
	SelfRank    int64    `codec:"selfrank"`    // 自己排名
	SelfName    string   `codec:"selfname"`    // 自己的名字
	SelfScore   int64    `codec:"selfscore"`   // 自己的积分
	SelfJoinGVG int64    `codec:"selfjoingvg"` // 自己是否参战
}

// GetGVGGuildTotalScore 获取所有军团所有玩家的总积分排名: 请求获取成员积分排名
func (p *Account) GetGVGGuildTotalScore(r servers.Request) *servers.Response {
	req := new(reqMsgGetGVGGuildTotalScore)
	rsp := new(rspMsgGetGVGGuildTotalScore)

	initReqRsp(
		"Attr/GetGVGGuildTotalScoreRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, false); errRsp != nil {
		return errRsp
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_Get_PlayerWorldInfo,
		AcID:   p.AccountID.String(),
		Name:   p.Profile.Name,
		GuID:   p.GuildProfile.GuildUUID,
		GuName: p.GuildProfile.GuildName,
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	if ret.SortItem == nil {
		return rpcSuccess(rsp)
	}
	if ret.Rank != 0 {
		rsp.SelfJoinGVG = 1
		rsp.SelfScore = int64(ret.Score)
		rsp.SelfRank = int64(ret.Rank)
	}
	rsp.SelfName = p.Profile.Name

	l := len(ret.SortItem)
	rsp.Names = make([]string, 0, l)
	rsp.Scores = make([]int64, 0, l)
	for _, item := range ret.SortItem {
		rsp.Names = append(rsp.Names, item.StrKey)
		rsp.Scores = append(rsp.Scores, item.IntVal)
	}

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetGVGGuildRank : 获取所有军团的积分排名
// 请求获取所有军团的积分排名

// reqMsgGetGVGGuildRank 获取所有军团的积分排名请求消息定义
type reqMsgGetGVGGuildRank struct {
	Req
}

// rspMsgGetGVGGuildRank 获取所有军团的积分排名回复消息定义
type rspMsgGetGVGGuildRank struct {
	SyncResp
	Names            []string `codec:"names"`            // 军团名字数组
	Scores           []int64  `codec:"scores"`           // 军团总积分
	SelfGuildRank    int64    `codec:"selfguildrank"`    // 本军团排名
	SelfGuildName    string   `codec:"selfguildname"`    // 本军团名字
	SelfGuildScore   int64    `codec:"selfguildscore"`   // 本军团积分
	SelfGuildJoinGVG int64    `codec:"selfguildjoingvg"` // 自己是否参战
}

// GetGVGGuildRank 获取所有军团的积分排名: 请求获取所有军团的积分排名
func (p *Account) GetGVGGuildRank(r servers.Request) *servers.Response {
	req := new(reqMsgGetGVGGuildRank)
	rsp := new(rspMsgGetGVGGuildRank)

	initReqRsp(
		"Attr/GetGVGGuildRankRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	// 检查公会状态(离开和换公会CD)
	if errRsp := p.checkGuildForGVG(rsp, false); errRsp != nil {
		return errRsp
	}

	ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
		Typ:    gvg.Cmd_Typ_Get_GuildWorldRank,
		AcID:   p.AccountID.String(),
		GuID:   p.GuildProfile.GuildUUID,
		GuName: p.GuildProfile.GuildName,
	})
	if ret.ErrCode != nil {
		return rpcErrorWithMsg(rsp, ret.ErrCode.Code(), ret.ErrCode.Error())
	}
	if ret.SortItem == nil {
		return rpcSuccess(rsp)
	}
	if ret.Rank != 0 {
		rsp.SelfGuildJoinGVG = 1
		rsp.SelfGuildScore = int64(ret.Score)
		rsp.SelfGuildRank = int64(ret.Rank)
	}
	rsp.SelfGuildName = p.GuildProfile.GuildName
	l := len(ret.SortItem)
	rsp.Names = make([]string, 0, l)
	rsp.Scores = make([]int64, 0, l)
	for _, item := range ret.SortItem {
		rsp.Names = append(rsp.Names, item.StrKey)
		rsp.Scores = append(rsp.Scores, item.IntVal)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
