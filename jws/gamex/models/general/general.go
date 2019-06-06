package general

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type General struct {
	Id     string `json:"id"` // id
	StarLv uint32 `json:"sl"` // 当前星级
	Num    uint32 `json:"n"`  // 当前碎片数量
}

func (g *General) AddGeneralNum(v uint32) {
	g.Num += v
}

func (g *General) AddStarLevel() (bool, uint32) {
	gCfg := gamedata.GetGeneralInfo(g.Id)
	wantStar := g.StarLv + 1
	if g.StarLv < gCfg.GetGeneralBeginStar() { // 未激活
		wantStar = gCfg.GetGeneralBeginStar()
	}

	ok, need := gamedata.GeneralStarNeedNum(g.Id, wantStar)
	if !ok {
		return false, 0
	}

	if g.Num >= need {
		logs.Trace("General Lv Up %v", *g)
		g.StarLv = wantStar
		g.Num -= need
		return true, g.StarLv
	}

	return false, 0
}

// 武将是否可以用 超过一级
func (g *General) IsHas() bool {
	return g.StarLv > 0
}
