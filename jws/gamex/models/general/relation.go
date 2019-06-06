package general

import (
	"math"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Relation struct {
	Id    string `json:"id"`  // id
	Level uint32 `json:"lvl"` // lvl
}

func (rel *Relation) RelationLevelup(pGenerals *PlayerGenerals, heroLv []uint32) (bool, uint32) {
	cfg := gamedata.GetGeneralRelationInfo(rel.Id)
	if cfg == nil {
		logs.Error("not found generalrelation %s", rel.Id)
		return false, 0
	}
	oldLvl := rel.Level
	var minStar uint32
	minStar = math.MaxUint32
	for idx, genId := range cfg.Generals {
		switch cfg.GeneralTypes[idx] {
		case gamedata.GeneralTypeInRelHero:
			heroIdx := cfg.HeroIdxIfTypeHero[idx]
			if heroIdx < 0 || heroIdx >= len(heroLv) {
				return false, 0
			}
			if heroLv[heroIdx] < minStar {
				minStar = heroLv[heroIdx]
			}
		case gamedata.GeneralTypeInRelGeneral:
			gen, ok := pGenerals.generals[genId]
			if !ok {
				return false, 0
			}
			if gen.StarLv < minStar {
				minStar = gen.StarLv
			}
		case gamedata.GeneralTypeInRelNull:
			return false, 0
		default:
			return false, 0
		}

	}
	if minStar > rel.Level {
		rel.Level += 1
		if oldLvl <= 0 {
			pGenerals.updateGen2ActRel(rel.Id, cfg)
		}
		return true, rel.Level
	}
	return false, 0
}
