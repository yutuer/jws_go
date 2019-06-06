package logics

func (s *SyncResp) makeExclusiveWeaponInfo(p *Account) {
	if s.SyncExclusiveWeaponNeed {
		hew_len := len(p.Profile.GetHero().HeroExclusiveWeapon)
		s.SyncExclusiveWeaponInfo = make([][]byte, hew_len)

		for i, weapon := range p.Profile.GetHero().HeroExclusiveWeapon {
			s.SyncExclusiveWeaponInfo[i] = encode(weapon)
			p.Profile.GetHero().HeroExclusiveWeapon[i].Clear()
		}
	}
}
