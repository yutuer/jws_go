package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

type CompanionInfo2Client struct {
	Companions  [][]byte `codec:"hcp_list"`
	EvolveLevel int      `codec:"hcp_evo_lv"`
}

type Companion2Client struct {
	CompanionId int `codec:"hcp_hero_idx"` // heroIdx  !!!关平的武将ID是0
	Level       int `codec:"hcp_c_lvl"`
}

func (s *SyncResp) mkHeroCompanionInfo(p *Account) {
	if p.Profile.GetData().IsNeedCheckCompanion() {
		s.SyncHeroCompanionNeed = true
		p.Profile.GetData().SetNeedCheckCompanion(false)
	}
	if s.SyncHeroCompanionNeed {
		hcp_len := len(p.Profile.GetHero().HeroCompanionInfos)
		s.SyncHeroCompanion = make([][]byte, hcp_len)

		for i := range p.Profile.GetHero().HeroCompanionInfos {
			pCompanion := &p.Profile.GetHero().HeroCompanionInfos[i]
			if pCompanion.HasCompanions(i) {
				s.SyncHeroCompanion[i] = encode(p.ConvertClientCompanion(i, pCompanion))
			}
		}
	}
}

func (p *Account) ConvertClientCompanion(heroIdx int, hcp *account.HeroCompanionInfo) CompanionInfo2Client {
	ret := CompanionInfo2Client{}
	ret.EvolveLevel = hcp.EvolveLevel
	companions := hcp.GetAllCompanions(heroIdx)
	ret.Companions = make([][]byte, len(companions))
	for i, cp := range companions {
		ret.Companions[i] = encode(Companion2Client{
			CompanionId: cp.GetCompanionId(),
			Level:       cp.GetLevel(),
		})
	}
	return ret
}

func (p *Account) IsHeroCompanionOpen(heroIdx int) bool {
	heroLevel := gamedata.GetHeroCommonConfig().GetCompanionUnlockLv()
	heroStar := gamedata.GetHeroCommonConfig().GetCompanionUnlockStar()
	return p.Profile.Hero.HeroLevel[heroIdx] >= heroLevel &&
		p.Profile.Hero.HeroStarLevel[heroIdx] >= heroStar
}
