package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/festivalboss"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FestivalBossEnd : 节日Boss挑战结束后协议
// 节日Boss挑战结束后返回boss等级和输赢结果

// reqMsgFestivalBossEnd 节日Boss挑战结束后协议请求消息定义
type reqMsgFestivalBossEnd struct {
	ReqWithAnticheat
	BossLvl    int64 `codec:"bosslvl"` // 挑战倍率0无倍率,1二倍,2五倍
	FestivalId int64 `codec:"fes_id"`  //节日类型ID
	StageId    int64 `codec:"stid"`    // 节日Boss关卡
	IsWin      bool  `codec:"iswin"`   // 是否胜利
}

// rspMsgFestivalBossEnd 节日Boss挑战结束后协议回复消息定义
type rspMsgFestivalBossEnd struct {
	SyncRespWithRewardsAnticheat
}

// FestivalBossEnd 节日Boss挑战结束后协议: 节日Boss挑战结束后返回boss等级和输赢结果
func (p *Account) FestivalBossEnd(r servers.Request) *servers.Response {
	req := new(reqMsgFestivalBossEnd)
	rsp := new(rspMsgFestivalBossEnd)

	initReqRsp(
		"Attr/FestivalBossEndRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		CODE_Cost_Err
	)

	// 反作弊检查
	if cheatRsp := p.AntiCheatCheck(&rsp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0,
		account.Anticheat_Typ_FestivalBoss); cheatRsp != nil {
		return cheatRsp
	}
	costTyp, costnum := gamedata.GetFestivalBossCostChallengeCost(req.FestivalId)
	data := &gamedata.CostData{}
	data.AddItem(costTyp, costnum)

	temples, tcs := gamedata.GetFestivalGameRewardCfg(uint32(req.FestivalId))

	givesAll := gamedata.NewPriceDatas(32)
	if req.IsWin {
		if req.BossLvl == 0 {
			if !account.CostBySync(p.Account, data, rsp, "FestivalBoss Cost") {
				return rpcErrorWithMsg(rsp, CODE_Cost_Err, "CODE_Cost_Baozi_Er")
			}
			for idx, tid := range temples {
				tc := tcs[idx]
				for i := 0; i < int(tc); i++ {
					gives, err := p.GetGivesByTemplate(tid)
					if err != nil {
						logs.Error("OnPlayerMsgFestival Loot err by %v",
							gives)
						continue
					}
					givesAll.AddOther(&gives)
				}
			}
		} else {
			multiple, costtype1, costnum2 := gamedata.GetFestivalBossReward(uint32(req.FestivalId), req.BossLvl)
			data.AddItem(costtype1, costnum2)
			if !account.CostBySync(p.Account, data, rsp, "FestivalBoss Cost") {
				return rpcErrorWithMsg(rsp, CODE_Cost_Err, "CODE_Cost_Er")
			}
			for idx, tid := range temples {
				tc := tcs[idx] * multiple
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
		}
		// 次数检查
		if !p.Profile.GetCounts().Use(counter.CounterTypeFestivalBoss, p.Account) {
			return rpcWarn(rsp, errCode.ClickTooQuickly)
		}
		// 更新击杀FestivalBoss次数
		p.Profile.GetFestivalBossInfo().UpdataFbFestivalBossKillTime()
		//称号
		p.Profile.GetTitle().OnFestivalBoss(p.Account)
		// 击杀存库
		festivalboss.GetModule(p.AccountID.ShardId).TryAddFestivalBossInfo(p.AccountID.ShardId, p.Profile.Name, p.Profile.GetProfileNowTime())
	}
	fbGives := gamedata.PriceDatas{}
	for idx, rid := range givesAll.Item2Client {
		c := givesAll.Count2Client[idx]
		fbGives.AddItem(rid, c)
	}
	account.GiveBySync(p.Account, fbGives.Gives(), rsp, "FESTIVAL REWARDS")
	var result int
	if req.IsWin {
		result = 1
	} else {
		result = 0
	}
	logiclog.LogFestivalBossFight(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		p.Profile.GetData().CorpCurrGS,
		p.Profile.GetFestivalBossInfo().GetBossKillTime(),
		result,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	rsp.OnChangeGameMode(counter.CounterTypeFestivalBoss)
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) getBIBaseInfo() logiclog.BIBaseInfo {
	return logiclog.BIBaseInfo{
		AccountID: p.AccountID.String(),
		Avatar:    p.Profile.GetCurrAvatar(),
		CorpLvl:   p.Profile.GetCorp().GetLvlInfo(),
		Channel:   p.Profile.ChannelId,
		Fgs:       func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
	}
}
