package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PlayerHeroTeam struct {
	Team []int `json:"tm"`
}

type PlayerHeroTeams struct {
	HeroTeams [gamedata.LEVEL_TYPE_Count]PlayerHeroTeam `json:"hero_tms"`
}

func (hts *PlayerHeroTeams) GetHeroTeam(typ int) []int {
	if typ >= len(hts.HeroTeams) {
		logs.Error("PlayerHeroTeams GetHeroTeam typ error %d", typ)
		return nil
	}
	return hts.HeroTeams[typ].Team
}

func (hts *PlayerHeroTeams) ResetHeroTeam(typ int) {
	if typ >= len(hts.HeroTeams) {
		logs.Error("PlayerHeroTeams GetHeroTeam typ error %d", typ)
	}
	hts.HeroTeams[typ].Team = make([]int, 0)
	logs.Debug("reset hero team typ: %v, team: %v", typ, hts.HeroTeams[typ].Team)
}
