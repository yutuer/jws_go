package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"

	"vcs.taiyouxi.net/jws/gamex/models/account"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

// ExpeditionOverFight : 远征战斗结束结算
// 远战战斗结算由客户端发来胜负消息,和主将状态

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑
const Has_fight = 1
const No_fight = 0

// reqMsgExpeditionOverFight 远征战斗结束结算请求消息定义
type reqMsgExpeditionOverFight struct {
	ReqWithAnticheat
	Iswin     int64   `codec:"iswin"`   // 战斗结果
	HeroId    []int64 `codec:"heroid"`  // 主将id
	HeroState []int64 `codec:"hestate"` // 主将状态

	HeroBlood   []float32 `codec:"heblood"` // 主将血量
	HeroWuSkill []float32 `codec:"hewu"`    // 主将无双技能槽
	HeroExSkill []float32 `codec:"heex"`    // 主将普通技能槽

	EnmyNum     int64     `codec:"enm"` //第几个敌人
	EnmyHp      []float32 `codec:"ehp"` //敌人的血量
	EnmyWuSkill []float32 `codec:"ews"` //敌人无双技能
	EnmyExSkill []float32 `codec:"exs"` //敌人普通技能

}

// rspMsgExpeditionOverFight 远征战斗结束结算回复消息定义
type rspMsgExpeditionOverFight struct {
	SyncRespWithRewardsAnticheat
}

// ExpeditionOverFight 远征战斗结束结算: 远战战斗结算由客户端发来胜负消息,和主将状态
func (p *Account) ExpeditionOverFight(r servers.Request) *servers.Response {
	req := new(reqMsgExpeditionOverFight)
	rsp := new(rspMsgExpeditionOverFight)

	initReqRsp(
		"Attr/ExpeditionOverFightRsp",
		r.RawBytes,
		req, rsp, p)

	// 反作弊检查
	if cheatRsp := p.AntiCheatCheck(&rsp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0,
		account.Anticheat_Typ_Expedition); cheatRsp != nil {
		return cheatRsp
	}

	pg := p.Profile.GetExpeditionInfo()
	if req.EnmyNum == int64(pg.ExpeditionAward) {

		for i, r := range req.HeroId {
			pg.ExpeditionMyHero[r].HeroIslive = int(req.HeroState[i])
			pg.ExpeditionMyHero[r].HeroHp = req.HeroBlood[i]
			pg.ExpeditionMyHero[r].HeroWuSkill = req.HeroWuSkill[i]
			pg.ExpeditionMyHero[r].HeroExSkill = req.HeroExSkill[i]

		}
		if req.Iswin == 0 {
			pg.ExpeditionEnmySkillInfo[req.EnmyNum-1].Hp = req.EnmyHp[:]
			pg.ExpeditionEnmySkillInfo[req.EnmyNum-1].WuSkill = req.EnmyWuSkill[:]
			pg.ExpeditionEnmySkillInfo[req.EnmyNum-1].ExSkill = req.EnmyExSkill[:]
			pg.ExpeditionEnmySkillInfo[req.EnmyNum-1].State = Has_fight
		}
		if req.Iswin == 1 {
			pg.ExpeditionAward += 1
			//条件更新
			p.updateCondition(account.COND_TYP_Expedition, 1, 0, "", "", rsp)
			p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(), gamedata.CounterTypeExpedition, 1, p.Profile.GetProfileNowTime())

		}
		logiclog.LogExpedition(
			p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId,
			int(req.Iswin),
			int(req.EnmyNum),
			int(p.Profile.GetData().CorpCurrGS),
			req.HeroId,
			0,
			func(last string) string {
				return p.Profile.GetLastSetCurLogicLog(last)
			},
			"")
	}
	rsp.OnChangerExpeditionInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
