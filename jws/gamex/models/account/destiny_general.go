package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util"
)

type destinyGeneral struct {
	Id         int    `json:"id"`
	LevelIndex int    `json:"lv"`
	Exp        uint32 `json:"exp"`
}

type PlayerDestinyGeneral struct {
	Generals            []destinyGeneral                   `json:"g"`
	SkillGenerals       [helper.DestinyGeneralSkillMax]int `json:"sg"`
	CurrGeneralIdx      int                                `json:"c"`
	VipTimes            uint32                             `json:"vip_ts"`
	VipRefreshTimeStamp int64                              `json:"vip_ref_t"`
	VipActive           bool                               `json:"vip_act"`
	isNeedSync          bool
}

func (p *PlayerDestinyGeneral) GetAllDestinyLv() (ret int) {
	for _, item := range p.Generals {
		ret += item.LevelIndex
	}
	return
}

func (p *PlayerDestinyGeneral) UpdateDGTimes(vip int, now_t int64) {
	if p.VipActive && now_t >= p.VipRefreshTimeStamp {
		p.VipTimes = gamedata.GetVIPCfg(vip).DGVipNormalTimes
		p.VipRefreshTimeStamp = util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(now_t), now_t)
	}
}

func (p *PlayerDestinyGeneral) OnVipLvUp(vip int, now_t int64) {
	vipTimes := gamedata.GetVIPCfg(vip).DGVipNormalTimes
	if vipTimes > 0 && !p.VipActive {
		p.VipActive = true
		p.UpdateDGTimes(vip, now_t)
		p.isNeedSync = true
	}
}

func (p *PlayerDestinyGeneral) AddNewGeneral(id int) {
	if p.Generals == nil {
		p.Generals = make([]destinyGeneral, 0, 32)
	}

	g := p.GetGeneral(id)
	if g != nil {
		return
	}

	p.Generals = append(p.Generals, destinyGeneral{
		Id: id,
	})

	p.isNeedSync = true

	for i := 0; i < len(p.Generals); i++ {
		lvData := gamedata.GetNewDestinyGeneralLevelDatas(p.Generals[i].Id)
		if lvData != nil {
			if p.Generals[i].LevelIndex < lvData[len(lvData)-1].LevelIndex {
				p.CurrGeneralIdx = p.Generals[i].Id
				return
			}
		}
	}
	p.CurrGeneralIdx = len(p.Generals) - 1
	return
}

func (p *PlayerDestinyGeneral) GetLastGeneralGiveGs() *gamedata.DestinyGeneralLevelData {
	// 神将数量很少
	for i := len(p.Generals) - 1; i >= 0; i-- {
		data := gamedata.GetNewDestinyGeneralLevelDatas(p.Generals[i].Id)
		dataUnLock := gamedata.GetDestinyGeneralUnlockData(p.Generals[i].Id)
		if data != nil && dataUnLock != nil && dataUnLock.IsCalcGs {
			return &data[p.Generals[i].LevelIndex]
		}
	}

	return nil
}

func (p *PlayerDestinyGeneral) GetGeneral(id int) *destinyGeneral {
	// 神将数量很少
	for i := len(p.Generals) - 1; i >= 0; i-- {
		if p.Generals[i].Id == id {
			return &p.Generals[i]
		}
	}

	return nil
}

func (p *PlayerDestinyGeneral) AddGeneralLevel(id, lv int) {
	g := p.GetGeneral(id)
	if g == nil {
		return
	}

	g.LevelIndex += lv
	p.isNeedSync = true
}

func (p *PlayerDestinyGeneral) SetSkills(skills []int) bool {
	for i := 0; i < helper.DestinyGeneralSkillMax && i < len(skills); i++ {
		if skills[i] != -1 && (p.GetGeneral(skills[i]) == nil) {
			return false
		}
		p.SkillGenerals[i] = skills[i] + 1 // +1来区分0号神将和空
	}
	p.isNeedSync = true
	return true
}

func (p *PlayerDestinyGeneral) IsNeedSync() bool {
	return p.isNeedSync
}

func (p *PlayerDestinyGeneral) HasSync() {
	p.isNeedSync = false
}

func (p *PlayerDestinyGeneral) IsJadeUnlock(id int, jadeIdx int) bool {
	dg := p.GetGeneral(id)
	if dg == nil {
		return false
	}
	data := gamedata.GetDestinyGeneralUnlockData(id)
	if data == nil {
		return false
	}
	if jadeIdx < 0 || jadeIdx >= len(data.PositionLvNeed) {
		return false
	}
	return dg.LevelIndex >= data.PositionLvNeed[jadeIdx]
}
