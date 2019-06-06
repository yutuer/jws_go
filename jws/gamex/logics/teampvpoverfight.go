package logics

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/team_pvp"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// TeamPvpOverFight : 3v3战斗结束结算，由客户端发来胜负结果
// 、IsWin: 1为胜利 0为平局 -1为失败

// reqMsgTeamPvpOverFight 3v3战斗结束结算，由客户端发来胜负结果请求消息定义
type reqMsgTeamPvpOverFight struct {
	ReqWithAnticheat
	IsWin int64 `codec:"is_win"` // 战斗结果
}

// rspMsgTeamPvpOverFight 3v3战斗结束结算，由客户端发来胜负结果回复消息定义
type rspMsgTeamPvpOverFight struct {
	SyncRespWithRewardsAnticheat
	RankChg int64    `codec:"rankchg"` // 排名变化
	Enemies [][]byte `codec:"enemies"` // 新的对手信息
}

// TeamPvpOverFight 3v3战斗结束结算，由客户端发来胜负结果: 、IsWin: 1为胜利 0为平局 -1为失败
// 暂时没有做平局的处理
func (p *Account) TeamPvpOverFight(r servers.Request) *servers.Response {
	req := new(reqMsgTeamPvpOverFight)
	rsp := new(rspMsgTeamPvpOverFight)

	initReqRsp(
		"Attr/TeamPvpOverFightRsp",
		r.RawBytes,
		req, rsp, p)
	// 反作弊检查
	if cheatRsp := p.AntiCheatCheck(&rsp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0,
		account.Anticheat_Typ_TeamPVP); cheatRsp != nil {
		return cheatRsp
	}

	const fetal_Err = 1
	// logic imp begin
	sid := p.AccountID.ShardId
	tpvp := p.Profile.GetTeamPvp()
	now_time := p.Profile.GetProfileNowTime()
	simpleInfo := p.Account.GetSimpleInfo()
	ret := team_pvp.GetModule(sid).CommandExec(team_pvp.TeamPvpCmd{
		Typ:      team_pvp.TeamPvp_Cmd_UnlockPlayerAndEnd,
		Acid:     p.AccountID.String(),
		IsWin:    req.IsWin > 0,
		EnemyId:  tpvp.FightEnemyID,
		AcidInfo: &simpleInfo,
	})
	if !ret.Success {
		if ret.RetState == team_pvp.InvalidState {
			return rpcWarn(rsp, errCode.TPVPInvalidFight)
		} else {
			logs.Error("Fetal Error")
			return rpcError(rsp, fetal_Err)
		}
	}

	oldRank := tpvp.Rank
	tpvp.SetRank(ret.MyNewRank, false)
	p.Profile.GetFirstPassRank().OnRank(gamedata.FirstPassRankTypTeamPvp, ret.MyNewRank)
	logs.Debug("logic, new rank is: %d, and enemy info is %v", ret.MyNewRank, ret.Enemies)
	rsp.RankChg = int64(calcRankChg(oldRank, ret.MyNewRank))
	eas := make([]int, 0, len(ret.Enemy.TeamPvpAvatar))
	eas = append(eas, ret.Enemy.TeamPvpAvatar[:]...)
	eslv := make([]int, 0, len(ret.Enemy.TeamPvpAvatarLv))
	for _, item := range ret.Enemy.TeamPvpAvatarLv {
		eslv = append(eslv, int(item))
	}
	aslv := make([]int, 0, len(simpleInfo.TeamPvpAvatarLv))
	for _, item := range simpleInfo.TeamPvpAvatarLv {
		aslv = append(aslv, int(item))
	}
	_, err := team_pvp.ParseTPvpRobotId(tpvp.FightEnemyID)

	if err != nil {
		sendFightRecord(tpvp.FightEnemyID, &tpvpRecord{
			TimStamp:    now_time,
			IsWin:       req.IsWin < 0,
			RankChg:     int(-1 * rsp.RankChg),
			EnemyName:   p.Profile.Name,
			EnemyGs:     simpleInfo.TeamPvpGs,
			EnemyAvatar: tpvp.FightAvatars,
			EnemyStarLv: aslv,
		})
	}
	myRecord := &tpvpRecord{
		TimStamp:    now_time,
		IsWin:       req.IsWin > 0,
		RankChg:     int(rsp.RankChg),
		EnemyName:   ret.Enemy.Name,
		EnemyGs:     ret.Enemy.TeamPvpGs,
		EnemyAvatar: eas,
		EnemyStarLv: eslv,
	}
	sendFightRecord(p.AccountID.String(), myRecord)

	// log
	lvl, _ := p.Profile.GetCorp().GetXpInfo()
	logicInfo := logiclog.LogicInfo_TeamPvp{
		AttackerAcid:      p.AccountID.String(),
		AttackerCorpLvl:   lvl,
		AttackerGs:        simpleInfo.TeamPvpGs,
		AttackerRankChg:   int(rsp.RankChg),
		AttackerAvatars:   p.Profile.GetTeamPvp().FightAvatars,
		AttackerWinRate:   fmt.Sprintf("%.2f", ret.WinRate),
		AttackerRnd:       fmt.Sprintf("%.5f", ret.MyRnd),
		BeAttackerAcid:    ret.Enemy.AccountID,
		BeAttackerCorpLvl: ret.Enemy.CorpLv,
		BeAttackerGs:      ret.Enemy.TeamPvpGs,
		BeAttackerRankChg: int(-1 * rsp.RankChg),
		BeAttackerAvatars: eas,
	}
	logiclog.LogTeamPvp(p.AccountID.String(), p.Profile.CurrAvatar,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, logicInfo, req.IsWin > 0,
		p.Profile.GetHero().HeroStarLevel, ret.Enemy.AvatarStarLvl,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	// 刷新敌人
	tpvp.NextEnemyRefTime = now_time +
		int64(gamedata.GetTPvpCommonCfg().GetTeamPvPRefreshTime())
	ret = team_pvp.GetModule(sid).CommandExec(team_pvp.TeamPvpCmd{
		Typ:  team_pvp.TeamPvp_Cmd_GetEnemy,
		Acid: p.AccountID.String(),
	})
	tpvp.Enemies = ret.Enemies
	// resp
	rsp.Enemies = make([][]byte, 0, len(tpvp.Enemies))
	for _, e := range tpvp.Enemies {
		rsp.Enemies = append(rsp.Enemies, encode(e))
	}

	// condition
	p.updateCondition(account.COND_TYP_TeamPvp_Times, 1, 0, "", "", rsp)

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeTeamPvp,
		1,
		p.Profile.GetProfileNowTime())

	rsp.OnChangeGameMode(counter.CounterTypeTeamPvp)
	rsp.OnChangeTeamPvp()

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
