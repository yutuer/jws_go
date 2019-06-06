package account

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/MagicPet"
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PlayerHero struct {
	helper.SyncObj
	HeroStarLevel       [AVATAR_NUM_MAX]uint32                 `json:"hl"`
	HeroStarPiece       [AVATAR_NUM_MAX]uint32                 `json:"hp"`
	HeroWholeCharHasGot [AVATAR_NUM_MAX]byte                   `json:"hwg"`
	HeroLevel           [AVATAR_NUM_MAX]uint32                 `json:"hlvl"`
	HeroExp             [AVATAR_NUM_MAX]uint32                 `json:"hexp"`
	HeroSkills          [AVATAR_NUM_MAX]HeroSkill              `json:"hesk"`
	HeroSwings          [AVATAR_NUM_MAX]HeroSwing              `json:"hsw"`
	HeroCompanionInfos  [AVATAR_NUM_MAX]HeroCompanionInfo      `json:"hci"`
	HeroExclusiveWeapon [AVATAR_NUM_MAX]HeroExclusiveWeapon    `json:"hew"`
	HeroMagicPets       [AVATAR_NUM_MAX]MagicPet.HeroMagicPets `json:"hmp"`
	IsNotShowMagicPet   bool                                   `json:"insmp"`
	handler             events.Handler
}

type HeroSwing struct {
	StarLv    int   `json:"hsw_s_lv"`
	Lv        int   `json:"hsw_lv"`
	ActSwings []int `json:"hsw_act"`
	CurSwing  int   `json:"hsw_cur"`
}

func (hs *HeroSwing) Reset() {
	hs.Lv = 0
	hs.StarLv = 0
	ret := make([]int, 0, 1)
	for _, i := range hs.ActSwings {
		if gamedata.GetHeroSwingInfo(i).GetHWUnlockType() == gamedata.Unlock_Typ_Manual {
			ret = append(ret, i)
		}
	}
	if len(ret) != 0 {
		hs.ActSwings = ret
		hs.CurSwing = ret[0]
	} else {
		hs.ActSwings = nil
		hs.CurSwing = 0
	}
	ret = nil
}

func (hs *HeroSwing) HasAct(id int) bool {
	if id == 0 {
		return true
	}
	if hs.ActSwings == nil {
		hs.InitSwing()
	}
	for _, sw := range hs.ActSwings {
		if sw == id {
			return true
		}
	}
	return false
}

func (hs *HeroSwing) UpdateAct() int {
	ret := -1
	canAct := gamedata.GetHeroSwingCanAct(uint32(hs.StarLv))
	for _, id := range canAct {
		if !hs.HasAct(id) {
			hs.ActSwing(id)
			ret = id
		}
	}
	return ret
}

func (hs *HeroSwing) ActSwing(id int) {
	if hs.ActSwings == nil {
		hs.InitSwing()
	}
	hs.ActSwings = append(hs.ActSwings, id)
}

func (hs *HeroSwing) InitSwing() {
	hs.ActSwings = make([]int, 0, 5)
}

type HeroSkill struct {
	CounterSkill []string `json:"hesk_1"`
	PassiveSkill []string `json:"hesk_2"`
	TriggerSkill []string `json:"hesk_3"`
}

func (ph *PlayerHero) IsHeroSkillActive(heroId int, skillid string) bool {
	p := ph.HeroSkills[heroId]
	for _, ok := range p.CounterSkill {
		if ok == skillid {
			return true
		}
	}
	for _, ok := range p.PassiveSkill {
		if ok == skillid {
			return true
		}
	}
	for _, ok := range p.TriggerSkill {
		if ok == skillid {
			return true
		}

	}
	return false
}

func (ph *PlayerHero) AddSkill2Hero(heroId int, skillid string) {
	p := &ph.HeroSkills[heroId]

	typ := gamedata.GetWhichSkillBySkillId(skillid)
	switch typ {
	case "Tkill":
		ss := p.TriggerSkill
		if ss == nil {
			ss = make([]string, 0, 10)
		}
		ss = append(ss, skillid)
		p.TriggerSkill = ss

	case "Pkill":
		ss := p.PassiveSkill
		if ss == nil {
			ss = make([]string, 0, 10)
		}
		ss = append(ss, skillid)
		p.PassiveSkill = ss

	case "Ckill":
		ss := p.CounterSkill
		if ss == nil {
			ss = make([]string, 0, 10)
		}
		ss = append(ss, skillid)
		p.CounterSkill = ss
	case "NOkill":
		logs.Error("Not found hero skillid %s", skillid)
	}

}

func (ph *PlayerHero) CheckHeroSkill() {
	for heroId, skillInfo := range ph.HeroSkills {
		ph.addCheckSill(skillInfo.TriggerSkill, heroId, gamedata.Tkill)
		ph.addCheckSill(skillInfo.PassiveSkill, heroId, gamedata.Pkill)
		ph.addCheckSill(skillInfo.CounterSkill, heroId, gamedata.Ckill)

	}
}

func (ph *PlayerHero) addCheckSill(heroSkill []string, heroId int, skillTyp string) {
	temp := make([]string, 0, 10)
	skills := make([]string, 0, 10) //加错类型的skillId
	for _, skill := range heroSkill {
		typ := gamedata.GetWhichSkillBySkillId(skill)
		if typ != skillTyp {
			skills = append(skills, skill)
			continue
		}
		temp = append(temp, skill)
	}
	if len(skills) > 0 {
		for _, id := range skills {
			ph.AddSkill2Hero(heroId, id)
		}
		p := &ph.HeroSkills[heroId]
		switch skillTyp {
		case gamedata.Tkill:
			p.TriggerSkill = temp
		case gamedata.Pkill:
			p.PassiveSkill = temp
		case gamedata.Ckill:
			p.CounterSkill = temp
		}
	}
}

func (p *PlayerHero) onAfterLogin(a *Account) {
	for avatarID := 0; avatarID < AVATAR_NUM_MAX; avatarID++ {
		info := gamedata.GetHeroData(avatarID)
		if info == nil {
			continue
		}
		if info.UnlockTyp == gamedata.HeroUnlockTypAuto &&
			p.HeroStarLevel[avatarID] == 0 {
			p.HeroStarLevel[avatarID] = info.UnlockInitLv
			p.HeroLevel[avatarID] = 1
			a.Profile.GetCorp().UnlockAvatar(a, avatarID)
		}
	}
}

func (p *PlayerHero) HeroStarActivity(a *Account) {
	//将星之路运营活动
	for i := 0; i < len(p.HeroStarLevel); i++ {
		starLvl := p.HeroStarLevel[i]
		a.Profile.GetMarketActivitys().OnHeroStar(a.AccountID.String(), int(i), int(starLvl),
			a.Profile.GetProfileNowTime())

	}
}

func (e *PlayerHero) SetHandler(handler events.Handler) {
	e.handler = handler
}

func (p *PlayerHero) Add(a *Account, avatarID int, piece uint32, reason string) {
	if avatarID < 0 || avatarID > len(p.HeroStarLevel) {
		return
	}
	old := p.HeroStarPiece[avatarID]
	p.HeroStarPiece[avatarID] += piece
	p.SetNeedSync()
	// log
	logiclog.LogHeroAddPiece(a.AccountID.String(), a.Profile.CurrAvatar, a.Profile.GetCorp().GetLvlInfo(),
		a.Profile.ChannelId, avatarID, piece, old, p.HeroStarPiece[avatarID], reason,
		func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
}

func (p *PlayerHero) Remove(a *Account, avatarID int, piece uint32, reason string) bool {
	if avatarID < 0 || avatarID > len(p.HeroStarLevel) {
		return false
	}
	if p.HeroStarPiece[avatarID] >= piece {
		p.HeroStarPiece[avatarID] -= piece
		p.SetNeedSync()
		return true
	} else {
		return false
	}
}

func (p *PlayerHero) AddInit(a *Account, avatarID int, piece uint32, reason string) {
	if avatarID < 0 || avatarID > len(p.HeroStarLevel) {
		return
	}
	//old := p.HeroStarPiece[avatarID]
	info := gamedata.GetHeroData(avatarID)
	p.HeroStarLevel[avatarID] = info.UnlockInitLv
	p.HeroLevel[avatarID] = 1
	a.Profile.GetCorp().UnlockAvatar(a, avatarID)
	p.SetNeedSync()

	simpleInfo := a.GetSimpleInfo()
	rank.GetModule(a.AccountID.ShardId).RankByHeroStar.Add(&simpleInfo)
	// log
	//logiclog.LogHeroAddPiece(a.AccountID.String(), a.Profile.CurrAvatar, a.Profile.GetCorp().GetLvlInfo(),
	//	a.Profile.ChannelId, avatarID, piece, old, p.HeroStarPiece[avatarID], reason,
	//	func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
}

func (p *PlayerHero) StarUp(a *Account, avatarID int, sync helper.ISyncRsp) (
	isSuccess bool, isUnlock bool, oldStar, newStar uint32) {
	info := gamedata.GetHeroData(avatarID)
	if info == nil {
		logs.Error("StarUp Err By Info Nil")
		return false, false, 0, 0
	}

	if !info.IsInCurrVersion {
		return false, false, 0, 0
	}

	lv := p.HeroStarLevel[avatarID]
	if int(lv+1) >= len(info.LvData) {
		logs.Trace("StarUp Err By Info lv %d", lv)
		return false, false, 0, 0
	}

	// UnlockInitLv
	if lv < info.UnlockInitLv-1 {
		lv = info.UnlockInitLv - 1
	}

	nextLvlCfg := info.LvData[lv+1]
	pieceNeed := nextLvlCfg.PieceNumToThisStar
	logs.Trace("HeroStar %d %d",
		pieceNeed, p.HeroStarPiece[avatarID])
	if p.HeroStarPiece[avatarID] >= pieceNeed {
		data := &gamedata.CostData{}
		data.AddItem(nextLvlCfg.CoinCost, nextLvlCfg.CoinCount)
		if uutil.IsOverseaVer() && nextLvlCfg.Cfg.GetHeroStarupItem() != "" {
			data.AddItem(nextLvlCfg.Cfg.GetHeroStarupItem(), nextLvlCfg.Cfg.GetHeroStarupCount())
		}
		cost := &CostGroup{}
		if !cost.AddCostData(a, data) || !cost.CostBySync(a, sync, "HeroStarUp") {
			logs.Error("HerpStarUp cost err %s %d %d", a.AccountID.String(), avatarID, lv+1)
			return false, false, 0, 0
		}

		old := p.HeroStarLevel[avatarID]
		isUnlock = p.HeroStarLevel[avatarID] == 0 // 是否是解锁
		p.HeroStarPiece[avatarID] -= pieceNeed
		p.HeroStarLevel[avatarID] = lv + 1
		if p.HeroLevel[avatarID] <= 0 {
			p.HeroLevel[avatarID] = 1
		}
		//p.SetNeedSync()

		simpleInfo := a.GetSimpleInfo()
		rank.GetModule(a.AccountID.ShardId).RankByHeroStar.Add(&simpleInfo)

		newStar := p.HeroStarLevel[avatarID]

		//激活将星活动
		a.Profile.GetMarketActivitys().OnHeroStar(a.AccountID.String(), avatarID, int(newStar), a.Profile.GetProfileNowTime())

		// 激活天赋
		a.Profile.GetHeroTalent().ActTalentByStar(avatarID, newStar)

		// sysnotice
		cfg := gamedata.HeroStarSysNotice(newStar)
		if cfg != nil {
			sysnotice.NewSysRollNotice(a.AccountID.ServerString(), int32(cfg.GetServerMsgID())).
				AddParam(sysnotice.ParamType_RollName, a.Profile.Name).
				AddParam(sysnotice.ParamType_Hero, fmt.Sprintf("%d", avatarID)).
				AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", newStar)).Send()
		}
		return true, isUnlock, old, newStar
	}

	return false, false, 0, 0
}

func (p *PlayerHero) GetStar(avatarID int) uint32 {
	if avatarID >= len(p.HeroStarLevel) {
		return 0
	}
	return p.HeroStarLevel[avatarID]
}

func (p *PlayerHero) GetLevel(avatarID int) uint32 {
	if avatarID >= len(p.HeroLevel) {
		return 0
	}
	return p.HeroLevel[avatarID]
}
func (p *PlayerHero) GetSwing(avatarID int) HeroSwing {
	return p.HeroSwings[avatarID]
}
func (p *PlayerHero) GetMagicPet(avatarID int) MagicPet.HeroMagicPets {
	return p.HeroMagicPets[avatarID]
}
func (p *PlayerHero) IsWholeCharHasGot(avatarId int) bool {
	return p.HeroWholeCharHasGot[avatarId] > 0
}

func (p *PlayerHero) SetWholeCharHasGot(avatarId int) {
	p.HeroWholeCharHasGot[avatarId] = 1
}

func (ph *PlayerHero) AddHeroExp(a *Account, avatarId int, xp uint32) {
	lvl := ph.HeroLevel[avatarId]
	befLvl := lvl
	if lvl <= 0 {
		logs.Warn("AddHeroExp but hero not unlock %v", avatarId)
		return
	}

	expLimit, ok := gamedata.GetHeroLevelExpLimit(int32(lvl))
	if !ok {
		logs.Error("AddHeroExp Xp Max Or Data Lose In %d", lvl)
		return
	}

	ph.HeroExp[avatarId] += xp
	clv, _ := a.Profile.GetCorp().GetXpInfo()

	for ph.HeroExp[avatarId] >= uint32(expLimit) {
		ph.HeroExp[avatarId] -= uint32(expLimit)

		if lvl >= clv {
			ph.HeroExp[avatarId] = uint32(expLimit)
			break
		}

		ph.HeroLevel[avatarId] += 1
		lvl = ph.HeroLevel[avatarId]

		expLimit, ok = gamedata.GetHeroLevelExpLimit(int32(lvl))
		if !ok {
			logs.Error("AddHeroExp2 Xp Max Or Data Lose In %d", lvl)
			break
		}
	}

	logs.Debug("Hero LevelUp %d %d->%d %d", avatarId, befLvl, lvl, ph.HeroExp[avatarId])
	if ph.handler != nil && lvl > befLvl {
		ph.handler.OnHeroLvUp(befLvl, lvl, ph.HeroExp[avatarId], "HeroLvlUp")
	}
}

func (ph *PlayerHero) GetOwnedHeroCount() (count int) {
	for _, lvInfo := range ph.HeroLevel {
		if lvInfo != 0 {
			count++
		} else {
			break
		}
	}
	return
}

func (p *PlayerHero) GetMagicPetFigure(avatarID int) uint32 {
	if p.ShowMagicPet(avatarID) {
		return gamedata.GetStar(p.HeroMagicPets[avatarID].GetPets()[0].Star).GetStarStage()
	} else {
		return 0
	}
}

func (p *PlayerHero) ShowMagicPet(avatarID int) bool {
	return p.HeroMagicPets[avatarID].GetPets()[0].Lev >= gamedata.GetMagicPetConfig().GetStarCondition()
}

func (p *PlayerHero) SetCurSwing(avatarID, swingID int) {
	p.HeroSwings[avatarID].CurSwing = swingID
}

func (p *PlayerHero) HasSwingAct(avatarID, swingID int) bool {
	return p.HeroSwings[avatarID].HasAct(swingID)
}

func (p *PlayerHero) GetCompanion(heroIdx int) *HeroCompanionInfo {
	return &p.HeroCompanionInfos[heroIdx]
}
