package logics

import (
	"fmt"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/message"
	"vcs.taiyouxi.net/jws/gamex/modules/team_pvp"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 打开界面相关信息
func (p *Account) GetTeamPvpInfo(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		Info [][]byte `codec:"topN"`
	}{}

	initReqRsp(
		"PlayerAttr/GetTeamPvpInfoRsp",
		r.RawBytes,
		req, resp, p)

	sid := p.AccountID.ShardId
	ret := team_pvp.GetModule(sid).CommandExec(team_pvp.TeamPvpCmd{
		Typ: team_pvp.TeamPvp_Cmd_GetRank,
	})

	resp.Info = make([][]byte, 0, len(ret.Enemies))
	for _, info := range ret.Enemies {
		resp.Info = append(resp.Info, encode(info))
	}

	resp.OnChangeTeamPvp()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 设置出站角色
func (p *Account) SetTeamPvpAvatar(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Avatars []int `codec:"avs"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/SetTeamPvpAvatarRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Avatar_Lock
	)

	if len(req.Avatars) != helper.TeamPvpAvatarsCount {
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}

	// 判断角色是否都解锁了
	for _, avid := range req.Avatars {
		if !p.Account.IsAvatarUnblock(avid) {
			return rpcErrorWithMsg(resp, Err_Avatar_Lock, "Err_Avatar_Lock")
		}
	}

	p.Profile.GetTeamPvp().FightAvatars = req.Avatars

	sid := p.AccountID.ShardId
	simpleInfo := p.Account.GetSimpleInfo()
	team_pvp.GetModule(sid).UpdateInfo(&simpleInfo)

	resp.OnChangeTeamPvp()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 获取敌人列表
func (p *Account) GetTeamPvpEnemy(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		Enemies [][]byte `codec:"enemies"`
	}{}

	initReqRsp(
		"PlayerAttr/GetTeamPvpEnemyRsp",
		r.RawBytes,
		req, resp, p)

	sid := p.AccountID.ShardId
	now_time := p.Profile.GetProfileNowTime()
	tpvp := p.Profile.GetTeamPvp()
	if tpvp.RankChgPassive || tpvp.NextEnemyRefTime == 0 || now_time > tpvp.NextEnemyRefTime {
		tpvp.NextEnemyRefTime = now_time +
			int64(gamedata.GetTPvpCommonCfg().GetTeamPvPRefreshTime())
		ret := team_pvp.GetModule(sid).CommandExec(team_pvp.TeamPvpCmd{
			Typ:  team_pvp.TeamPvp_Cmd_GetEnemy,
			Acid: p.AccountID.String(),
		})
		tpvp.Enemies = ret.Enemies
		tpvp.RankChgPassive = false
	}
	resp.Enemies = make([][]byte, 0, len(tpvp.Enemies))
	for _, e := range tpvp.Enemies {
		resp.Enemies = append(resp.Enemies, encode(e))
	}
	return rpcSuccess(resp)
}

// 刷新敌人列表
func (p *Account) RefreshTeamPvpEnemy(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		Enemies [][]byte `codec:"enemies"`
	}{}

	initReqRsp(
		"PlayerAttr/RefreshTeamPvpEnemyRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_GetCounts
		Err_CostData_Not_Found
		Err_Cost
		Err_Times_Not_Enough
	)

	sid := p.AccountID.ShardId
	now_time := p.Profile.GetProfileNowTime()
	lc, nt := p.Profile.GetCounts().Get(
		counter.CounterTypeTeamPvpRefresh, p.Account)
	if nt < 0 || lc < 1 {
		return rpcErrorWithMsg(resp, Err_Times_Not_Enough, "Err_GetCounts")
	}
	c := gamedata.GetGameModeControlData(counter.CounterTypeTeamPvpRefresh)
	ut := c.GetCount - lc + 1
	costCfg := gamedata.GetTPvpRefreshCost(uint32(ut))
	if costCfg == nil {
		return rpcErrorWithMsg(resp, Err_CostData_Not_Found,
			fmt.Sprintf("Err_CostData_Not_Found times %d", ut))
	}
	cost := &account.CostGroup{}
	if !cost.AddCostData(p.Account, costCfg) || !cost.CostBySync(p.Account, resp, "TeamPvpRefreshEnemy") {
		return rpcErrorWithMsg(resp, Err_Cost, "Err_Cost")
	}

	ok := p.Profile.GetCounts().Use(
		counter.CounterTypeTeamPvpRefresh, p.Account)
	if !ok {
		return rpcErrorWithMsg(resp, Err_Times_Not_Enough, "Err_Times_Not_Enough")
	}

	// 刷新敌人
	tpvp := p.Profile.GetTeamPvp()
	tpvp.NextEnemyRefTime = now_time +
		int64(gamedata.GetTPvpCommonCfg().GetTeamPvPRefreshTime())
	ret := team_pvp.GetModule(sid).CommandExec(team_pvp.TeamPvpCmd{
		Typ:  team_pvp.TeamPvp_Cmd_GetEnemy,
		Acid: p.AccountID.String(),
	})
	tpvp.Enemies = ret.Enemies
	// resp
	resp.Enemies = make([][]byte, 0, len(tpvp.Enemies))
	for _, e := range tpvp.Enemies {
		resp.Enemies = append(resp.Enemies, encode(e))
	}
	resp.OnChangeGameMode(counter.CounterTypeTeamPvpRefresh)
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 战斗
const (
	MyRankChgPassive = 1
	EnemyRankChg     = 2
)

func (p *Account) TeamPvpFight(r servers.Request) *servers.Response {
	req := &struct {
		Req
		EnemyAcid string `codec:"eid"`
		EnemyRank int    `codec:"erk"`
	}{}
	resp := &struct {
		SyncResp
		NeedRefTyp       int      `codec:"reftyp"`
		NeedRefEnemyInfo [][]byte `codec:"enmyinfo"`
		Avatar           [][]byte `codec:"avatar"`
		IsWin            bool     `codec:"iswin"`
		RankChg          int      `codec:"rankchg"`
		Enemies          [][]byte `codec:"enemies"`
		Params           []string `codec:"p"` // for test 当前存储我方战力与敌方战力
	}{}

	initReqRsp(
		"PlayerAttr/TeamPvpFightRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Times_Not_Enough
		Err_LoadPvPAccount
		Err_FromAccountByDroid
		Err_ParseAccount
		Err_FromAccount
	)

	sid := p.AccountID.ShardId
	tpvp := p.Profile.GetTeamPvp()
	now_time := p.Profile.GetProfileNowTime()

	// 自己的排名被动的变化了，先免费更新一次敌人列表
	if tpvp.RankChgPassive {
		tpvp.NextEnemyRefTime = now_time +
			int64(gamedata.GetTPvpCommonCfg().GetTeamPvPRefreshTime())
		ret := team_pvp.GetModule(sid).CommandExec(team_pvp.TeamPvpCmd{
			Typ:  team_pvp.TeamPvp_Cmd_GetEnemy,
			Acid: p.AccountID.String(),
		})
		tpvp.Enemies = ret.Enemies
		tpvp.RankChgPassive = false
		resp.NeedRefEnemyInfo = make([][]byte, 0, len(tpvp.Enemies))
		for _, e := range tpvp.Enemies {
			resp.NeedRefEnemyInfo = append(resp.NeedRefEnemyInfo, encode(e))
		}
		resp.NeedRefTyp = MyRankChgPassive
		return rpcSuccess(resp)
	}

	// 开始战斗, 锁定敌人
	ret := team_pvp.GetModule(sid).CommandExec(team_pvp.TeamPvpCmd{
		Typ:       team_pvp.TeamPvp_Cmd_LockPlayerAndBegin,
		Acid:      p.AccountID.String(),
		EnemyId:   req.EnemyAcid,
		EnemyRank: req.EnemyRank,
	})

	tpvp.FightEnemyID = req.EnemyAcid
	resp.Params = ret.Params
	if len(ret.Enemies) > 0 {
		tpvp.NextEnemyRefTime = now_time +
			int64(gamedata.GetTPvpCommonCfg().GetTeamPvPRefreshTime())
		tpvp.Enemies = ret.Enemies
		resp.NeedRefEnemyInfo = make([][]byte, 0, len(tpvp.Enemies))
		for _, e := range tpvp.Enemies {
			resp.NeedRefEnemyInfo = append(resp.NeedRefEnemyInfo, encode(e))
		}
		resp.NeedRefTyp = EnemyRankChg
		return rpcSuccess(resp)
	}

	if ret.RetState == team_pvp.LockState {
		resp.mkInfo(p)
		return rpcWarn(resp, errCode.TPVPEnemyLocked)
	}
	ok, code, warnCode, _ := p.Profile.GetGameMode().GameModeCheckAndCost(p.Account,
		gamedata.CounterTypeTeamPvp, 1, resp)
	if !ok {
		if warnCode > 0 {
			logs.Warn("TeamPvpFight GameModeCheckAndCost %d", code)
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
		// TODO by ljz unlock the enemy
		return rpcError(resp, 20+code)
	}

	robotId, err := team_pvp.ParseTPvpRobotId(req.EnemyAcid)
	if err != nil { // 玩家
		dbEnemyID, err := db.ParseAccount(req.EnemyAcid)
		if err != nil {
			return rpcErrorWithMsg(resp, Err_ParseAccount, "Err_ParseAccount")
		}
		enemy_account, err := account.LoadPvPAccount(dbEnemyID)
		if err != nil {
			return rpcErrorWithMsg(resp, Err_LoadPvPAccount, "Err_LoadPvPAccount")
		}
		for _, eav := range ret.Enemy.TeamPvpAvatar {
			a := helper.Avatar2Client{}
			err = account.FromAccount(&a, enemy_account, eav)
			if err != nil {
				return rpcErrorWithMsg(resp, Err_FromAccount, "Err_FromAccount")
			}
			//a.GS = GetCurrGS(enemy_account) // 前端不用
			resp.Avatar = append(resp.Avatar, encode(a))
		}

	} else { // 机器人
		for _, eav := range ret.Enemy.TeamPvpAvatar {
			droid := gamedata.GetDroidForTeamPvp(uint32(robotId.RobotCfgId))
			a := helper.Avatar2Client{}
			err := account.FromAccountByDroid(&a, droid, eav)
			if err != nil {
				return rpcErrorWithMsg(resp, Err_FromAccountByDroid, "Err_FromAccountByDroid")
			}
			a.Name = ret.Enemy.Name
			resp.Avatar = append(resp.Avatar, encode(a))
		}
	}
	p.Profile.GetData().Times_3V3++
	p.Profile.GetTeamPvp().PvpCountToday += 1

	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 查看战斗记录
func (p *Account) GetTeamPvpRecord(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		Records [][]byte `codec:"recs"`
	}{}

	initReqRsp(
		"PlayerAttr/GetTeamPvpRecordRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_DB
	)

	msgs, err := message.LoadPlayerMsgs(p.AccountID.String(),
		TeamPvpFightRecordKey, TeamPvpFightRecordCount)
	if err != nil {
		return rpcErrorWithMsg(resp, Err_DB, "Err_DB")
	}

	resp.Records = make([][]byte, 0, len(msgs))
	for _, msg := range msgs {
		m := tpvpRecord{}
		err := json.Unmarshal([]byte(msg.Params[0]), &m)
		if err != nil {
			continue
		}
		resp.Records = append(resp.Records, encode(m))
	}
	if len(resp.Records) > TeamPvpFightRecordCount2Client {
		resp.Records = resp.Records[0:TeamPvpFightRecordCount2Client]
	}
	return rpcSuccess(resp)
}

const TeamPvpFightRecordCount = 30
const TeamPvpFightRecordKey = "TeamPvpRecord"
const TeamPvpFightRecordCount2Client = 10

type tpvpRecord struct {
	TimStamp    int64  `json:"st" codec:"st"`
	IsWin       bool   `json:"win" codec:"win"`
	RankChg     int    `json:"rc" codec:"rc"`
	EnemyName   string `json:"en" codec:"en"`
	EnemyGs     int    `json:"egs" codec:"egs"`
	EnemyAvatar []int  `json:"eas" codec:"eas"`
	EnemyStarLv []int  `json:"st_lv" codec:"st_lv"`
}

func sendFightRecord(acid string, rec *tpvpRecord) {
	bb, err := json.Marshal(*rec)
	if err != nil {
		logs.Error("TeamPvp fightRecord json err %s", err.Error())
	}
	message.SendPlayerMsgs(acid, TeamPvpFightRecordKey, TeamPvpFightRecordCount,
		message.PlayerMsg{
			Params: []string{string(bb)},
		})
}

func calcRankChg(oldRank, newRank int) int {
	if oldRank <= 0 {
		oldRank = int(gamedata.GetTPvpRankMax()) + 1
	}
	if newRank <= 0 {
		newRank = int(gamedata.GetTPvpRankMax()) + 1
	}
	return oldRank - newRank
}
