package logics

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/gve_notify"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) OnPlayerMsgGVEStart(r servers.Request) *servers.Response {
	req := player_msg.PlayerMsgGVEGameStart{}

	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerMsgGVEStart %s %v", p.AccountID.String(), req)

	accData := &helper.Avatar2ClientByJson{}
	// 用客户端设置的阵容里的角色
	curAvatar := p.Profile.CurrAvatar
	heroTm := p.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_TEAMBOSS)
	if heroTm != nil && len(heroTm) > 0 {
		curAvatar = heroTm[0]
	}
	account.FromAccount2Json(accData, p.Account, curAvatar)
	p.Tmp.GetGVEData()
	temples, tcs := gamedata.GetGVEGameRewardCfg(p.Tmp.GameIsHard)

	givesAll := gamedata.NewPriceDatas(32)
	var rewardCountRatio uint32 = 1

	if p.Tmp.GameIsDouble {
		rewardCountRatio = rewardCountRatio * 2
	}

	if p.Tmp.GameIsUseHc {
		rewardCountRatio = rewardCountRatio * 2
	}

	for idx, tid := range temples {
		tc := tcs[idx] * rewardCountRatio
		for i := 0; i < int(tc); i++ {
			gives, err := p.GetGivesByTemplate(tid)
			if err != nil {
				logs.Error("OnPlayerMsgGVEStart Loot err by %v",
					gives)
				continue
			}
			givesAll.AddOther(&gives)
		}
	}

	gve_notify.SendAccountData(accData.AcID,
		req.GameID,
		req.GameServerUrl,
		accData, givesAll.Item2Client, givesAll.Count2Client,
		p.Tmp.GameIsDouble, p.Tmp.GameIsUseHc, req.IsBot)
	p.Tmp.SetGVEData(
		req.GameID,
		req.GameSecret,
		req.GameServerUrl,
		givesAll.Item2Client,
		givesAll.Count2Client,
		req.IsBot)
	p.Tmp.CurrWaitting = false

	// log
	gs := GetCurrGS(p.Account)
	logiclog.LogGveGameStart(p.AccountID.String(), p.Profile.CurrAvatar,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, p.Tmp.GameID, []string{},
		p.Tmp.GameIsHard, p.Tmp.GameIsDouble, p.Tmp.GameIsUseHc,
		time.Now().Unix()-p.Tmp.MatchBeginTime, gs,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return nil
}

func (p *Account) OnPlayerMsgGVEStop(r servers.Request) *servers.Response {
	req := player_msg.PlayerMsgGVEGameStop{}

	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerMsgGVEStop %s %v", p.AccountID.String(), req)

	hcCost, hardHcCost, scCosts := gamedata.GetGVEGameCostCfg()

	// log
	gs := GetCurrGS(p.Account)
	logiclog.LogGveGameStop(p.AccountID.String(), p.Profile.CurrAvatar,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, p.Tmp.GameID, []string{},
		req.IsSuccess, p.Tmp.GameIsHard, p.Tmp.GameIsDouble, p.Tmp.GameIsUseHc, gs,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	if req.IsSuccess && req.IsHasReward {
		gveCost := gamedata.CostData{}
		currCount := p.Tmp.CurrCount
		if currCount < 0 || currCount >= len(scCosts) {
			currCount = len(scCosts) - 1
		}
		gveCost.AddItem(gamedata.VI_Sc0, scCosts[currCount])
		if p.Tmp.GameIsUseHc {
			if p.Tmp.GameIsHard {
				gveCost.AddItem(gamedata.VI_Hc, uint32(hardHcCost))
			} else {
				gveCost.AddItem(gamedata.VI_Hc, uint32(hcCost))
			}
		}
		if !account.CostBySync(p.Account, &gveCost, nil, "GVE") {
			logs.SentryLogicCritical(p.AccountID.String(), "GVE Use HC Err")
			p.Tmp.CleanGVEData()
			return nil
		}

		if !p.Profile.GetCounts().UseJustDayEnd(counter.CounterTypeGVE, p.Account) {
			logs.SentryLogicCritical(p.AccountID.String(), "GVE Use Count Err")
			p.Tmp.CleanGVEData()
			return nil
		}

		gveGives := gamedata.PriceDatas{}
		for idx, rid := range p.Tmp.GameRewards {
			c := p.Tmp.GameCounts[idx]
			gveGives.AddItem(rid, c)
		}

		account.GiveBySync(p.Account, gveGives.Gives(), nil, "GVE")
		// condition
		p.updateCondition(account.COND_TYP_Gve_Times, 1, 0, "", "", nil)
		// market activity
		p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
			gamedata.CounterTypeGVE,
			1,
			p.Profile.GetProfileNowTime())
	}
	p.Tmp.CleanGVEData()
	return nil
}
