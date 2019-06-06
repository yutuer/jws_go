package logics

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//协议-----押运粮草:查看目标公会的信息和活动粮车
type reqMsgCSRobGuildInfo struct {
	Req
	GuildID string `codec:"guild_id"`
}

type rspMsgCSRobGuildInfo struct {
	Resp
	Type      uint32   `codec:"type"`       //活动类型
	GuildInfo []byte   `codec:"guild_info"` //目标公会信息 CSRobGuildInfo
	CarList   [][]byte `codec:"car_list"`   //目标公会粮车 []CSRobCarInfo
}

func (p *Account) CSRobGuildInfo(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobGuildInfo)
	rsp := new(rspMsgCSRobGuildInfo)

	initReqRsp(
		"Attr/CSRobGuildInfoRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Type = 0
	rsp.GuildInfo = []byte{}
	rsp.CarList = [][]byte{}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	rsp.Type, _ = gamedata.CSRobBattleIDAndHeroID(time.Now().Unix())

	var guild *csrob.Guild
	if req.GuildID == p.Account.GuildProfile.GuildUUID {
		guild = csrob.GetModule(p.AccountID.ShardId).GuildMod.GuildWithNew(p.Account.GuildProfile.GuildUUID, p.Account.GuildProfile.GuildName)
	} else {
		guild = csrob.GetModule(p.AccountID.ShardId).GuildMod.Guild(req.GuildID)
	}
	if nil == guild {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	info := guild.GetInfo()
	rsp.GuildInfo = encode(buildCSRobGuildInfo(info))

	//活动时间检查
	if false == csrobCheckTimeIn() {
		rsp.CarList = [][]byte{}
		return rpcSuccess(rsp)
	}

	list, err := guild.GetCarList()
	if nil != err {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	rsp.CarList = make([][]byte, 0, len(list))
	for _, car := range list {
		rsp.CarList = append(rsp.CarList, encode(buildCSRobCarInfo(&car)))
	}

	return rpcSuccess(rsp)
}

//协议-----押运粮草:查看本公会的仇敌列表和推荐列表
type reqMsgCSRobGuildList struct {
	Req
	CheckGuildID string `codec:"checkguid"` //触发查看的目标公会ID
}

type rspMsgCSRobGuildList struct {
	Resp
	Enemy [][]byte `codec:"enemy"` //公会仇敌列表 []CSRobGuildEnemy
	List  [][]byte `codec:"list"`  //公会推荐列表 []CSRobGuildInfo
}

func (p *Account) CSRobGuildList(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobGuildList)
	rsp := new(rspMsgCSRobGuildList)

	initReqRsp(
		"Attr/CSRobGuildListRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Enemy = [][]byte{}
	rsp.List = [][]byte{}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	guild := csrob.GetModule(p.AccountID.ShardId).GuildMod.GuildWithNew(p.Account.GuildProfile.GuildUUID, p.Account.GuildProfile.GuildName)
	if nil == guild {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	checkBestGrade := uint32(0)
	if "" != req.CheckGuildID {
		checkBestGrade = csrob.GetModule(p.AccountID.ShardId).GuildMod.GetGuildBestGrade(req.CheckGuildID)
	}

	list := guild.GetEnemies()
	logs.Trace("[CSRob] CSRobGuildList GetEnemies %+v", list)
	if 0 != len(list) {
		rsp.Enemy = make([][]byte, 0, len(list))
		for _, enemy := range list {
			if enemy.GuildID == req.CheckGuildID {
				enemy.BestGrade = checkBestGrade
			}
			rsp.Enemy = append(rsp.Enemy, encode(buildCSRobGuildEnemy(&enemy)))
		}
	}

	recommends := guild.GetList()
	if 0 != len(recommends) {
		rsp.List = make([][]byte, 0, len(recommends))
		for _, re := range recommends {
			if re.GuildID == req.CheckGuildID {
				re.BestGrade = checkBestGrade
			}
			rsp.List = append(rsp.List, encode(buildCSRobGuildInfo(&re)))
		}
	}

	return rpcSuccess(rsp)
}

//协议-----押运粮草:查看本公会的友方阵容列表
type reqMsgCSRobTeamsList struct {
	Req
}

type rspMsgCSRobTeamsList struct {
	Resp
	List [][]byte `codec:"list"` //公会阵容列表 []CSRobGuildTeam
}

func (p *Account) CSRobTeamsList(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobTeamsList)
	rsp := new(rspMsgCSRobTeamsList)

	initReqRsp(
		"Attr/CSRobTeamsListRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.List = [][]byte{}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	guild := csrob.GetModule(p.AccountID.ShardId).GuildMod.GuildWithNew(p.Account.GuildProfile.GuildUUID, p.Account.GuildProfile.GuildName)
	if nil == guild {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	pm := player_msg.GetModule(p.AccountID.ShardId)
	list := guild.GetTeams()
	if 0 != len(list) {
		rsp.List = make([][]byte, 0, len(list))
		for _, team := range list {
			if team.Acid == p.AccountID.String() || 0 == len(team.Hero) {
				continue
			}
			netTeam := buildCSRobGuildTeam(&team)
			netTeam.IsOnline = pm.IsOnline(team.Acid)
			rsp.List = append(rsp.List, encode(netTeam))
		}
	}

	return rpcSuccess(rsp)
}

//协议-----押运粮草:查看公会排行
type reqMsgCSRobGuildRank struct {
	Req
}

type rspMsgCSRobGuildRank struct {
	Resp
	Me   []byte   `codec:"me"`   //我自己的排名信息 CSRobGuildRankElem
	List [][]byte `codec:"list"` //公会排行列表 []CSRobGuildRankElem
}

//CSRobGuildRank ..
func (p *Account) CSRobGuildRank(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobGuildRank)
	rsp := new(rspMsgCSRobGuildRank)

	initReqRsp(
		"Attr/CSRobGuildRankRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.List = [][]byte{}
	rsp.Me = []byte{}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	guild := csrob.GetModule(p.AccountID.ShardId).GuildMod.GuildWithNew(p.Account.GuildProfile.GuildUUID, p.Account.GuildProfile.GuildName)
	if nil == guild {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	list := guild.GetRankList()
	if 0 != len(list) {
		rsp.List = make([][]byte, 0, len(list))
		for _, g := range list {
			rsp.List = append(rsp.List, encode(buildCSRobGuildRankElem(&g)))
		}
	}

	myRank := guild.GetMyRank()
	rsp.Me = encode(buildCSRobGuildRankElem(&myRank))

	return rpcSuccess(rsp)
}
