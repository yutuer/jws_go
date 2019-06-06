package logics

func (s *SyncResp) mkUpdateHeroInfo(p *Account) {
	s.mkUpdateStarLevel(p)
	s.mkUpdateStarPiece(p)
	s.mkUpdateLevel(p)
	s.mkUpdateExp(p)
	s.mkUpdateSkill(p)
	s.mkUpdateWing(p)
	s.mkUpdateCompanion(p)
	s.mkUpdateExclusive(p)
}

func (s *SyncResp) mkUpdateStarLevel(p *Account) {
	if s.UpdateHeroStarLevelNeed {
		s.UpdateHeroStarLevel = int(p.Profile.GetHero().HeroStarLevel[s.ChangedHeroAvatar])
	}
}

func (s *SyncResp) mkUpdateStarPiece(p *Account) {
	if s.UpdateHeroStarPieceNeed {
		s.UpdateHeroStarPiece = int(p.Profile.GetHero().HeroStarPiece[s.ChangedHeroAvatar])
	}
}

func (s *SyncResp) mkUpdateLevel(p *Account) {
	if s.UpdateHeroLevelNeed {
		s.UpdateHeroLevel = int(p.Profile.GetHero().HeroLevel[s.ChangedHeroAvatar])
	}
}

func (s *SyncResp) mkUpdateExp(p *Account) {
	if s.UpdateHeroExpNeed {
		s.UpdateHeroExp = int64(p.Profile.GetHero().HeroExp[s.ChangedHeroAvatar])
	}
}

type HeroSkill2Client struct {
	PassiveSkills []string `codec:"up_hs_p"`
	CounterSkills []string `codec:"up_hs_c"`
	TriggerSkills []string `codec:"up_hs_t"`
}

func (s *SyncResp) mkUpdateSkill(p *Account) {
	if s.UpdateHeroSkillsNeed {
		skill := p.Profile.GetHero().HeroSkills[s.ChangedHeroAvatar]
		skillClient := HeroSkill2Client{}
		skillClient.PassiveSkills = skill.PassiveSkill
		skillClient.CounterSkills = skill.CounterSkill
		skillClient.TriggerSkills = skill.TriggerSkill
		s.UpdateHeroSkills = encode(skillClient)
	}
}

func (s *SyncResp) mkUpdateWing(p *Account) {
	if s.UpdateHeroWingsNeed {
		wing := p.Profile.GetHero().HeroSwings[s.ChangedHeroAvatar]
		hsw := HeroSwingToClient{}
		hsw.StarLv = wing.StarLv
		hsw.Lv = wing.Lv
		hsw.ActSwings = make([]int, 0, len(wing.ActSwings))
		for _, sw := range wing.ActSwings {
			hsw.ActSwings = append(hsw.ActSwings, sw)
		}
		hsw.CurSwing = wing.CurSwing
		s.UpdateHeroWings = encode(hsw)
	}
}

func (s *SyncResp) mkUpdateCompanion(p *Account) {
	if s.UpdateHeroCompanionNeed {
		pCompanion := &p.Profile.GetHero().HeroCompanionInfos[s.ChangedHeroAvatar]
		if pCompanion.HasCompanions(s.ChangedHeroAvatar) {
			s.UpdateHeroCompanion = encode(p.ConvertClientCompanion(s.ChangedHeroAvatar, pCompanion))
		}
	}
}

func (s *SyncResp) mkUpdateExclusive(p *Account) {
	if s.UpdateHeroExclusiveNeed {
		s.UpdateHeroExclusive = encode(p.Profile.GetHero().HeroExclusiveWeapon[s.ChangedHeroAvatar])
		p.Profile.GetHero().HeroExclusiveWeapon[s.ChangedHeroAvatar].Clear()
	}
}
