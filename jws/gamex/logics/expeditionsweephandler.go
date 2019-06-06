package logics

import (
	"math"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

// ExpeditionSweep : 扫荡远征
// 扫荡远征
func (p *Account) ExpeditionSweepHandler(req *reqMsgExpeditionSweep, resp *rspMsgExpeditionSweep) uint32 {
	const (
		_ = iota
		CODE_Cost_Err
		COED_Time_Not_Enough
	)
	data1 := &gamedata.CostData{}
	data1.AddItem(gamedata.VI_XZ_SD, gamedata.GetExpeditionCfg().GetSweepCost())

	if !account.CostBySync(p.Account, data1, resp, "ExpeditionSweep Cost") {
		return CODE_Cost_Err
	}
	gsHero := p.Profile.GetData().HeroGs

	pe := p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.EnemyId-1]
	enemyHpInfo := &p.Profile.GetExpeditionInfo().ExpeditionEnmySkillInfo[req.EnemyId-1]

	pes := p.Profile.GetExpeditionInfo()

	FightDetail := make([]ExpeditionFightDetail, 0, 5)
	var enemyID = 0
	var fightId = 0
	for _, hId := range req.FightHeroId {

		if enemyID > 2 {
			break
		}
		if fightId > 5 {
			break
		}
		for {
			if enemyID > 2 {
				break
			}
			if fightId > 5 {
				break
			}
			if pes.ExpeditionMyHero[hId].HeroHp <= 0 {
				break
			}
			if enemyHpInfo.Hp[enemyID] <= 0 {
				enemyID += 1
				continue
			}
			currenGs := float32(gsHero[hId]) * gamedata.GetExpeditionCfg().GetGSBuff()
			if currenGs >= float32(pe.HeroGs[enemyID]) {
				myHpId := float32(currenGs-float32(pe.HeroGs[enemyID])) / float32(currenGs) * 100
				if myHpId <= 1 {
					myHpId = 1
				}

				M := gamedata.GetExpeditionSweep(uint32(math.Ceil(float64(myHpId))))

				hpKey := pes.ExpeditionMyHero[hId].HeroHp / enemyHpInfo.Hp[enemyID]
				if hpKey > M {
					//记录战斗结束状态 我方赢了
					FightDetail = append(FightDetail, ExpeditionFightDetail{
						AvatarId:      int64(hId),
						BeforeHp:      int64(pes.ExpeditionMyHero[hId].HeroHp * 100),
						AfterHp:       int64((pes.ExpeditionMyHero[hId].HeroHp - M*enemyHpInfo.Hp[enemyID]) * 100),
						EnemyId:       int64(enemyID),
						EnemyBeforeHp: int64(enemyHpInfo.Hp[enemyID] * 100),
						EnemyAfterHp:  0,
					})
					//自己状态更改
					pes.ExpeditionMyHero[hId].HeroIslive = 0

					pes.ExpeditionMyHero[hId].HeroHp = pes.ExpeditionMyHero[hId].HeroHp - M*enemyHpInfo.Hp[enemyID]
					//敌人状态更改
					enemyHpInfo.State = 1
					enemyHpInfo.Hp[enemyID] = 0
					enemyID += 1
				} else if hpKey < M {
					//记录战斗结束状态 我方输了
					FightDetail = append(FightDetail, ExpeditionFightDetail{
						AvatarId:      int64(hId),
						BeforeHp:      int64(pes.ExpeditionMyHero[hId].HeroHp * 100),
						AfterHp:       0,
						EnemyId:       int64(enemyID),
						EnemyBeforeHp: int64(enemyHpInfo.Hp[enemyID] * 100),
						EnemyAfterHp:  int64((enemyHpInfo.Hp[enemyID] - (pes.ExpeditionMyHero[hId].HeroHp / M)) * 100),
					})
					//敌人状态更改
					enemyHpInfo.State = 1

					if enemyHpInfo.Hp[enemyID]-(pes.ExpeditionMyHero[hId].HeroHp/M) <= 0 {
						enemyHpInfo.Hp[enemyID] = 0
					} else {
						enemyHpInfo.Hp[enemyID] = enemyHpInfo.Hp[enemyID] - (pes.ExpeditionMyHero[hId].HeroHp / M)
					}

					//自己状态更改
					pes.ExpeditionMyHero[hId].HeroIslive = 1

					pes.ExpeditionMyHero[hId].HeroHp = 0
					break
				} else {
					//记录战斗结束状态 同归于今
					FightDetail = append(FightDetail, ExpeditionFightDetail{
						AvatarId:      int64(hId),
						BeforeHp:      int64(pes.ExpeditionMyHero[hId].HeroHp * 100),
						AfterHp:       0,
						EnemyId:       int64(enemyID),
						EnemyBeforeHp: int64(enemyHpInfo.Hp[enemyID] * 100),
						EnemyAfterHp:  0,
					})

					//敌人状态更改
					enemyHpInfo.State = 1
					enemyHpInfo.Hp[enemyID] = 0

					//自己状态更改
					pes.ExpeditionMyHero[hId].HeroIslive = 1

					pes.ExpeditionMyHero[hId].HeroHp = 0
					enemyID += 1
					break
				}

				fightId += 1

			} else {
				// 对方战力高的时候
				enemyHpId := float32(float32(pe.HeroGs[enemyID])-currenGs) / float32(pe.HeroGs[enemyID]) * 100
				if enemyHpId == 0 {
					enemyHpId = 1
				}
				M := gamedata.GetExpeditionSweep(uint32(math.Ceil(float64(enemyHpId))))

				hpKey := enemyHpInfo.Hp[enemyID] / pes.ExpeditionMyHero[hId].HeroHp
				if hpKey < M {
					//记录战斗结束状态 我方赢了
					FightDetail = append(FightDetail, ExpeditionFightDetail{
						AvatarId:      int64(hId),
						BeforeHp:      int64(pes.ExpeditionMyHero[hId].HeroHp * 100),
						AfterHp:       int64((pes.ExpeditionMyHero[hId].HeroHp - enemyHpInfo.Hp[enemyID]/M) * 100),
						EnemyId:       int64(enemyID),
						EnemyBeforeHp: int64(enemyHpInfo.Hp[enemyID] * 100),
						EnemyAfterHp:  0,
					})

					//自己状态更改
					pes.ExpeditionMyHero[hId].HeroIslive = 0
					pes.ExpeditionMyHero[hId].HeroHp = pes.ExpeditionMyHero[hId].HeroHp - enemyHpInfo.Hp[enemyID]/M
					//敌人状态更改
					enemyHpInfo.State = 1
					enemyHpInfo.Hp[enemyID] = 0
					enemyID += 1

				} else if hpKey > M {
					//记录战斗结束状态 我方输了
					FightDetail = append(FightDetail, ExpeditionFightDetail{
						AvatarId:      int64(hId),
						BeforeHp:      int64(pes.ExpeditionMyHero[hId].HeroHp * 100),
						AfterHp:       0,
						EnemyId:       int64(enemyID),
						EnemyBeforeHp: int64(enemyHpInfo.Hp[enemyID] * 100),
						EnemyAfterHp:  int64((enemyHpInfo.Hp[enemyID] - (pes.ExpeditionMyHero[hId].HeroHp * M)) * 100),
					})
					//敌人状态更改
					enemyHpInfo.State = 1
					if enemyHpInfo.Hp[enemyID]-(pes.ExpeditionMyHero[hId].HeroHp*M) <= 0 {
						enemyHpInfo.Hp[enemyID] = 0
					} else {
						enemyHpInfo.Hp[enemyID] -= pes.ExpeditionMyHero[hId].HeroHp * M
					}
					//自己状态更改
					pes.ExpeditionMyHero[hId].HeroIslive = 1
					pes.ExpeditionMyHero[hId].HeroHp = 0
					break
				} else {
					//记录战斗结束状态 同归于今
					FightDetail = append(FightDetail, ExpeditionFightDetail{
						AvatarId:      int64(hId),
						BeforeHp:      int64(pes.ExpeditionMyHero[hId].HeroHp * 100),
						AfterHp:       0,
						EnemyId:       int64(enemyID),
						EnemyBeforeHp: int64(enemyHpInfo.Hp[enemyID] * 100),
						EnemyAfterHp:  0,
					})

					//敌人状态更改
					enemyHpInfo.State = 1
					enemyHpInfo.Hp[enemyID] = 0

					//自己状态更改
					pes.ExpeditionMyHero[hId].HeroIslive = 1

					pes.ExpeditionMyHero[hId].HeroHp = 0
					break
				}

				fightId += 1

			}

		}

	}
	var isWin = 1
	//检查是否胜利
	for _, state := range enemyHpInfo.Hp {
		if state > 0 {
			isWin = 0
		}
	}
	if isWin == 1 {
		pes.ExpeditionAward += 1
		resp.IsWin = 1
		//条件更新
		p.updateCondition(account.COND_TYP_Expedition, 1, 0, "", "", resp)
		p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(), gamedata.CounterTypeExpedition,
			1, p.Profile.GetProfileNowTime())
		logiclog.LogExpedition(
			p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId,
			int(resp.IsWin),
			int(req.EnemyId),
			int(p.Profile.GetData().CorpCurrGS),
			req.FightHeroId,
			1,
			func(last string) string {
				return p.Profile.GetLastSetCurLogicLog(last)
			},
			"")
	} else {
		resp.IsWin = 0
		logiclog.LogExpedition(
			p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId,
			int(resp.IsWin),
			int(req.EnemyId),
			int(p.Profile.GetData().CorpCurrGS),
			req.FightHeroId,
			1,
			func(last string) string {
				return p.Profile.GetLastSetCurLogicLog(last)
			},
			"")
	}
	//同步客户端状态
	for _, x := range FightDetail {
		resp.FightResult = append(resp.FightResult, encode(x))
	}
	return 0
}
