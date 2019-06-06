package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Exp struct {
	ArousalLv uint32 `json:"alv"` // 觉醒等级
	Level     uint32 `json:"lv"`  // 等级
	Xp        uint32 `json:"xp"`  // 当前经验
	handler   events.Handler
}

func (e *Exp) Add(xp uint32, t int, p *Account) {
	e.Xp += xp
	e.update(t, p)
}

func (e *Exp) OnLevelUp(t int) {
	// 这里不再使用了
	//
	// 注意
	// 现在要求去掉角色等级而使用战队等级
	// 这使得之前经验升级变成了经验升级和解锁升级,又因为解锁是玩家手动选择的所以需要在解锁时触发AvatarLv的handle
	// 同时AvatarLv的handle的行为和CorpLvHandle的行为不同, 需要保留
	//
	//if e.handler != nil {
	//
	//	e.handler.OnAvatarLvUp(t, e.Level)
	//}
}

func (e *Exp) update(t int, p *Account) {
	lv_data := gamedata.GetPlayerLevelAttr(e.Level)
	if lv_data == nil {
		logs.Error("Xp Max Or Data Lose In %d", e.Level)
		return
	}

	clv, _ := p.Profile.GetCorp().GetXpInfo()

	for e.Xp >= uint32(lv_data.GetXP()) {
		e.Xp -= uint32(lv_data.GetXP())

		if e.Level >= clv {
			e.Xp = uint32(lv_data.GetXP())
			return
		}

		e.Level += 1
		e.OnLevelUp(t)

		lv_data = gamedata.GetPlayerLevelAttr(e.Level)

		if lv_data == nil {
			logs.Error("Xp Max Or Data Lose In %d", e.Level)
			return
		}

	}
}

type AvatarExp struct {
	Avatars [AVATAR_NUM_MAX]Exp
	p       *Profile
}

func (e *AvatarExp) SetAccount(p *Profile) {
	e.p = p
}

func (e *AvatarExp) SetHandler(handler events.Handler) {
	for i := 0; i < len(e.Avatars); i++ {
		e.Avatars[i].handler = handler
	}
}

func (p *AvatarExp) AddExp(pA *Account, t int, m uint32, reason string) {
	if t >= len(p.Avatars) {
		logs.Error("No Acatar Type %d", t)
		return
	}

	p.Avatars[t].Add(m, t, pA)
}

func (p *AvatarExp) AddExp2All(pA *Account, m uint32, reason string) {
	for i := 0; i < AVATAR_NUM_CURR; i++ {
		p.Avatars[i].Add(m, i, pA)
	}
}

func (p *AvatarExp) Get(t int) (uint32, uint32) {
	if p.p != nil {
		return p.p.GetCorp().GetXpInfo()
	}
	return 1, 0

}

func (p *AvatarExp) GetArousalLv(t int) uint32 {
	p.Init()
	if t >= len(p.Avatars) {
		return 0
	} else {
		return p.Avatars[t].ArousalLv
	}
}

func (p *AvatarExp) AddArousalLv(t int) bool {
	p.Init()
	if t >= len(p.Avatars) {
		return false
	} else {
		p.Avatars[t].ArousalLv += 1
		return true
	}
}

func (p *AvatarExp) Init() {
	for i := 0; i < AVATAR_NUM_CURR; i++ {
		if p.Avatars[i].Level == 0 {
			p.Avatars[i].Level = 1
			p.Avatars[i].OnLevelUp(i)
		} else {
			return
		}
	}
}

func (p *AvatarExp) GetAll() []uint32 {
	p.Init()
	re := make([]uint32, AVATAR_NUM_MAX*2, AVATAR_NUM_MAX*2)
	for i, _ := range p.Avatars {
		re[i*2], re[i*2+1] = p.Get(i)
	}
	return re[:]
}

func (p AvatarExp) GetAvatarArousalLv() []uint32 {
	p.Init()
	re := make([]uint32, AVATAR_NUM_MAX, AVATAR_NUM_MAX)
	for i, exp := range p.Avatars {
		re[i] = exp.ArousalLv
	}
	return re[:]
}
