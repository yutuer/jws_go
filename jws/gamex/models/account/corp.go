package account

import (
	"fmt"

	"math"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MAX_UINT32 = math.MaxUint32
)

type Corp struct {
	Level uint32 `json:"lv"`
	Xp    uint32 `json:"xp"`

	UnlockAvatars []byte `json:"unl"`
	isNeedSync    bool

	unlockAvatar2Client []int

	handler events.Handler
	player  *Profile
}

func (c *Corp) OnAccountInit() {
	c.UnlockAvatars = make([]byte, AVATAR_NUM_CURR, AVATAR_NUM_MAX)
}

func (c *Corp) OnAfterLogin() {
	c.rebuild()
}

func (c *Corp) rebuild() {
	c.unlockAvatar2Client = make([]int, 0, AVATAR_NUM_CURR)
	for idx, b := range c.UnlockAvatars {
		if b > 0 {
			c.unlockAvatar2Client =
				append(c.unlockAvatar2Client, idx)
		}
	}
}

func (c *Corp) IsAvatarHasUnlock(avatarID int) bool {
	if avatarID < 0 || avatarID >= len(c.UnlockAvatars) {
		return false
	}
	return c.UnlockAvatars[avatarID] > 0
}

func (c *Corp) HasAvatarHasUnlok() int {
	var avatarNum int = 0
	for _, num := range c.UnlockAvatars {
		if num > 0 {
			avatarNum = avatarNum + 1
		}
	}
	return avatarNum
}

//TODO by ljz 修改函数形参(p)
func (c *Corp) UnlockAvatar(p *Account, avatarID int) {
	if c.IsAvatarHasUnlock(avatarID) {
		return
	} else {
		acid := p.AccountID.String()
		for avatarID >= len(c.UnlockAvatars) {
			c.UnlockAvatars =
				append(c.UnlockAvatars, 0)
		}
		c.UnlockAvatars[avatarID] = 1
		c.rebuild()
		c.isNeedSync = true

		newStar := c.player.GetHero().HeroStarLevel[avatarID]
		// 激活天赋
		c.player.GetHeroTalent().ActTalentByStar(avatarID, newStar)
		// 跑马灯
		info := gamedata.GetHeroData(avatarID)
		cfg := gamedata.HeroUnlockSysNotice()
		if info.RareLv >= cfg.GetLampValueIP1() {
			id, _ := db.ParseAccount(acid)
			sysnotice.NewSysRollNotice(id.ServerString(), int32(cfg.GetServerMsgID())).
				AddParam(sysnotice.ParamType_RollName, c.player.Name).
				AddParam(sysnotice.ParamType_Hero, fmt.Sprintf("%d", avatarID)).Send()
		}
		// 解锁穿装备
		af := gamedata.GetAvatarInitFashionData(avatarID)
		if af != nil { // 在创建角色的时候其实已经加了所有英雄时装，这里再加一次主要为了老账号加新英雄
			AvatarGiveAndThenEquip(p, avatarID, af.GetInitFWeapon(), gamedata.FashionPart_Weapon)
			AvatarGiveAndThenEquip(p, avatarID, af.GetInitFAmor(), gamedata.FashionPart_Armor)
		}
		p.Profile.GetData().SetNeedCheckMaxGS()

		// log
		logiclog.LogHeroUnlock(
			acid,
			c.player.CurrAvatar,
			c.player.GetCorp().GetLvlInfo(),
			c.player.ChannelId, avatarID, p.Profile.Hero.HeroStarPiece[avatarID],
			func(last string) string {
				return c.player.GetLastSetCurLogicLog(last)
			},
			"")
	}
}

func (c *Corp) GetUnlockedAvatar() []int {
	if c.unlockAvatar2Client == nil || len(c.unlockAvatar2Client) == 0 {
		c.rebuild()
	}
	return c.unlockAvatar2Client
}

func (c *Corp) IsNeedSyncUnlocked() bool {
	return c.isNeedSync
}

func (c *Corp) SetNoNeedSync() {
	c.isNeedSync = false
}

func (e *Corp) SetHandler(handler events.Handler) {
	e.handler = handler
}

func (e *Corp) AddExp(account string, xp uint32, reason string) {
	if e.handler != nil {
		e.handler.OnCorpExpAdd(e.Xp, xp, reason)
	}

	//防止经验增加溢出
	if xp > MAX_UINT32-e.Xp {
		xp = MAX_UINT32 - e.Xp
	}

	e.Xp += xp
	e.update(account, reason)
}

func (e *Corp) GetXpInfo() (uint32, uint32) {
	e.update(e.player.dbkey.Account.String(), "refreshByGet")
	return e.Level, e.Xp
}

// 此接口给离线啦玩家存档的玩法，比如pvp之类的用
func (e *Corp) GetXpInfoNoUpdate() (uint32, uint32) {
	return e.Level, e.Xp
}

func (e *Corp) GetLvlInfo() uint32 {
	e.update(e.player.dbkey.Account.String(), "refreshByGet")
	return e.Level
}

func (e *Corp) OnLevelUp(account string, l uint32, reason string) {
	logs.Trace("[%s]Corp Level Up %d", account, e.Level)

	if e.handler != nil {
		e.handler.OnCorpLvUp(l, e.Xp, reason)
	}

}

func (e *Corp) update(account, reason string) {
	info := gamedata.GetCorpLvConfig(e.Level)
	if info == nil {
		logs.Error("Corp Xp Max Or Data Lose In %d", e.Level)
		return
	}
	befExp := e.Xp
	befLevel := e.Level
	for e.Xp >= uint32(info.CorpXpNeed) {
		// 达到等级上限了
		limit := gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
		if e.Level >= limit {
			logs.Trace("Level up to limit %d %d", e.Level, limit)

			//记录溢出的经验值, 不再截断经验值, by 20170317 qiaozhu
			//e.Xp = uint32(info.CorpXpNeed)
			break
		}

		e.Xp -= uint32(info.CorpXpNeed)
		e.Level += 1

		info = gamedata.GetCorpLvConfig(e.Level)

		if info == nil {
			logs.Error("Corp Xp Max Or Data Lose In %d", e.Level)
			break
		}
	}

	toLv := e.Level
	toExp := e.Xp
	if toLv > befLevel {
		for l := befLevel + 1; l <= toLv; l++ {
			e.OnLevelUp(account, l, reason)
		}

		logiclog.LogCorpLevelChg(account, e.player.GetCurrAvatar(),
			e.player.GetCorp().GetLvlInfo(), e.player.ChannelId, reason,
			befLevel, befExp, toLv, toExp,
			func(last string) string { return e.player.GetLastSetCurLogicLog(last) }, "")
	}
}
