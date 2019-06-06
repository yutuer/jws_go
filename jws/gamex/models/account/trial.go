package account

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

type PlayerTrial struct {
	MostLevelId  int32 `json:"ml"` // 最远关卡id
	CurLevelId   int32 `json:"cl"` // 当前可以打的关卡id，BonusLevelId有效时此值无效；当和MostLevelId一样时无效
	BonusLevelId int32 `json:"bl"` // 当前宝箱关卡id，大于0有效

	SweepEndTime    int64        `json:"swet"`  // 扫荡结束时间
	SweepStartTime  int64        `json:"swst"`  // 扫荡开始时间
	SweepBeginLvlId int32        `json:"swbl"`  // 扫荡开始关卡id，大于0有效：表明当前有奖励可领
	SweepAwards     []TrialAward `json:"swads"` // 本次扫荡所有关的奖励，会在扫荡领奖后清理

	IsActivate bool `json:"isact"` // 是否激活
}

type TrialAward struct {
	LevelId   int32    `json:"lvl"`
	SC        int32    `json:"sc"`
	DC        int32    `json:"dc"`
	FI        int32    `json:"fi"`
	SB        int32    `json:"sb"`
	ItemId    []string `json:"itmid"` // 非空才有效，目前限制奖励物品只是宝石
	ItemCount []uint32 `json:"itmc"`
}

func (ta *TrialAward) AddAward(itemId string, count uint32) {
	if ta.ItemId == nil {
		ta.ItemId = make([]string, 0, 5)
		ta.ItemCount = make([]uint32, 0, 5)
	}
	ta.ItemId = append(ta.ItemId, itemId)
	ta.ItemCount = append(ta.ItemCount, count)
}

func (tr *PlayerTrial) Init() {
	if tr.CurLevelId <= 0 {
		tr.CurLevelId = gamedata.GetTrialFirstLvlId()
	}
}

func (tr *PlayerTrial) NextLvl() (isFirstPassLvl bool) {
	if tr.BonusLevelId > 0 {
		return isFirstPassLvl
	}
	lvlCfg := gamedata.GetTrialLvlById(tr.CurLevelId)
	// 更新最远关卡id
	if tr.MostLevelId < tr.CurLevelId {
		tr.MostLevelId = tr.CurLevelId
		isFirstPassLvl = true
		// 更新宝箱状态
		if lvlCfg.GetBonus() > 0 {
			tr.BonusLevelId = tr.CurLevelId
		}
	}
	// 更新当前可打的关卡
	nextLvlCfg := gamedata.GetTrialLvlByIndex(lvlCfg.GetTrialIndex() + 1)
	if nextLvlCfg != nil {
		tr.CurLevelId = nextLvlCfg.GetLevelID()
	} else if tr.CurLevelId == gamedata.GetTrialFinalLvlId() {
		tr.CurLevelId = tr.CurLevelId + 1
	}
	return isFirstPassLvl
}

func (tr *PlayerTrial) SetCurLvl2Most() {
	if tr.MostLevelId <= 0 {
		return
	}
	lvlCfg := gamedata.GetTrialLvlById(tr.MostLevelId)
	nextLvlCfg := gamedata.GetTrialLvlByIndex(lvlCfg.GetTrialIndex() + 1)
	if nextLvlCfg != nil {
		tr.CurLevelId = nextLvlCfg.GetLevelID()
		return
	}
	tr.CurLevelId = tr.MostLevelId + 1
}

func (tr *PlayerTrial) DebugSetCurLvl(a *Account, lvl int32) {
	if a.Profile.GetProfileNowTime() < tr.SweepEndTime {
		return
	}
	if tr.SweepBeginLvlId > 0 {
		return
	}

	if lvl > gamedata.GetTrialFinalLvlId() {
		lvl = gamedata.GetTrialFinalLvlId()
	}
	if lvl <= 0 {
		lvl = gamedata.GetTrialFirstLvlId()
	}
	tr.CurLevelId = lvl
	if tr.CurLevelId > tr.MostLevelId+1 {
		tr.MostLevelId = tr.CurLevelId - 1
	}
	lvlCfg := gamedata.GetTrialLvlById(tr.MostLevelId)
	if lvlCfg.GetBonus() > 0 {
		tr.BonusLevelId = tr.MostLevelId
	} else {
		tr.BonusLevelId = 0
	}
}

func (tr *PlayerTrial) MergeAward() (sc, dc, fi, sb int32, items map[string]uint32) {
	items = make(map[string]uint32, len(tr.SweepAwards))
	for _, aw := range tr.SweepAwards {
		sc += aw.SC
		dc += aw.DC
		fi += aw.FI
		sb += aw.SB
		if aw.ItemId != nil {
			for i := 0; i < len(aw.ItemId); i++ {
				if c, ok := items[aw.ItemId[i]]; ok {
					items[aw.ItemId[i]] = c + aw.ItemCount[i]
				} else {
					items[aw.ItemId[i]] = aw.ItemCount[i]
				}
			}
		}
	}
	return sc, dc, fi, sb, items
}
