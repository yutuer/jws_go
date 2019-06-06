package logics

type HeroSwingToClient struct {
	StarLv    int   `codec:"hsw_s_lv"`
	Lv        int   `codec:"hsw_lv"`
	ActSwings []int `codec:"hsw_act"`
	CurSwing  int   `codec:"hsw_cur"`
}

func (s *SyncResp) mkHeroSwingAllInfo(p *Account) {
	if s.SyncHeroSwingNeed {
		sw_len := len(p.Profile.GetHero().HeroSwings)
		s.SyncHeroSwing = make([][]byte, sw_len, sw_len)
		for i, _ := range p.Profile.GetHero().HeroSwings {
			p.Profile.GetHero().HeroSwings[i].UpdateAct()
			v := p.Profile.GetHero().HeroSwings[i]
			hsw := HeroSwingToClient{}
			hsw.StarLv = v.StarLv
			hsw.Lv = v.Lv
			hsw.ActSwings = make([]int, 0, len(v.ActSwings))
			for _, sw := range v.ActSwings {
				hsw.ActSwings = append(hsw.ActSwings, sw)
			}
			hsw.CurSwing = v.CurSwing
			s.SyncHeroSwing[i] = encode(hsw)
		}

		s.SyncShowHeroSwing = !p.Profile.IsHideSwing
	}
}
