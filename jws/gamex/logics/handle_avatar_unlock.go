package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func onAvatarUnlock(a *Account, avatarID int) {
	a.Profile.GetCorp().UnlockAvatar(a.Account, avatarID)
	h := a.Profile.GetHero()
	info := gamedata.GetHeroData(avatarID)
	if h.HeroStarLevel[avatarID] == 0 { // 防止是升星解锁
		h.HeroStarLevel[avatarID] = info.UnlockInitLv
		h.HeroStarPiece[avatarID] = 0
		if h.HeroLevel[avatarID] <= 0 {
			h.HeroLevel[avatarID] = 1
		}
		h.SetNeedSync()
	}
	acID := a.AccountID.String()
	logs.Trace("[%s]onAvatarUnlock avatarID %d", acID, avatarID)
}
