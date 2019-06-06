package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/servers"
)

func (p *Account) unlockAvatar(r servers.Request) *servers.Response {
	req := &struct {
		Req
		AvatarID int `json:"aid"`
	}{}

	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/UnlockAvatarRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ uint32 = iota
		Err_Param
	)

	p.Profile.GetCorp().UnlockAvatar(p.Account, req.AvatarID)

	// 解锁穿装备
	af := gamedata.GetAvatarInitFashionData(req.AvatarID)
	if af != nil {
		account.AvatarGiveAndThenEquip(p.Account, req.AvatarID, af.GetInitFWeapon(), gamedata.FashionPart_Weapon)
		account.AvatarGiveAndThenEquip(p.Account, req.AvatarID, af.GetInitFAmor(), gamedata.FashionPart_Armor)
	}

	p.Profile.CurrAvatar = req.AvatarID
	// 更新排行榜
	simpleInfo := p.GetSimpleInfo()
	lv, _ := p.Profile.GetCorp().GetXpInfo()
	if lv >= FirstIntoCorpLevel {
		rank.GetModule(p.AccountID.ShardId).RankCorpGs.Add(&simpleInfo,
			int64(simpleInfo.CurrCorpGs), int64(simpleInfo.CurrCorpGs))
	}

	resp.OnChangeUnlockAvatar()
	resp.OnChangeHeroTalent()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
