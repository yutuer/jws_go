package astrology

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Hero ..
type Hero struct {
	HeroID   uint32  `json:"id,omitempty"`
	Holes    []*Hole `json:"hs,omitempty"`
	mapHoles map[uint32]*Hole
}

//GoString GoStringer interface
func (h *Hero) GoString() string {
	str := ""

	str += fmt.Sprintf("{HeroID:%v,Holes:{", h.HeroID)
	for _, hole := range h.Holes {
		str += fmt.Sprintf("%#v,", hole)
	}
	str += "}}"

	return str
}

//Hole ..
type Hole struct {
	HoleID  uint32 `json:"id,omitempty"`
	Rare    uint32 `json:"r,omitempty"`
	Upgrade uint32 `json:"u,omitempty"`
}

//GoString GoStringer interface
func (h *Hole) GoString() string {
	str := ""

	str += fmt.Sprintf("{HoleID:%v,Rare:%v,Upgrade:%v}", h.HoleID, h.Rare, h.Upgrade)

	return str
}

func newHero(id uint32) *Hero {
	return &Hero{
		HeroID:   id,
		Holes:    []*Hole{},
		mapHoles: map[uint32]*Hole{},
	}
}

func (h *Hero) afterLogin() {
	h.mapHoles = map[uint32]*Hole{}
	for _, hole := range h.Holes {
		h.mapHoles[hole.HoleID] = hole
	}
}

//addHole ..
func (h *Hero) addHole(id, rare, upgrade uint32) *Hole {
	hole, exist := h.mapHoles[id]
	if false == exist {
		hole := &Hole{
			HoleID:  id,
			Rare:    rare,
			Upgrade: upgrade,
		}
		h.Holes = append(h.Holes, hole)
		h.mapHoles[id] = hole

		return nil
	}

	old := &Hole{
		HoleID:  hole.HoleID,
		Rare:    hole.Rare,
		Upgrade: hole.Upgrade,
	}

	hole.HoleID = id
	hole.Rare = rare
	hole.Upgrade = upgrade

	return old
}

//GetHoles ..
func (h *Hero) GetHoles() []*Hole {
	return h.Holes
}

//GetHole ..
func (h *Hero) GetHole(id uint32) *Hole {
	return h.mapHoles[id]
}

//UnsetHole ..
func (h *Hero) UnsetHole(id uint32) *Hole {
	for i := 0; i < len(h.Holes); i++ {
		if id == h.Holes[i].HoleID {
			old := h.Holes[i]
			if i+1 < len(h.Holes) {
				h.Holes = append(h.Holes[:i], h.Holes[i+1:]...)
			} else {
				h.Holes = h.Holes[:i]
			}
			delete(h.mapHoles, id)
			return old
		}
	}
	return nil
}

//IntoHole 镶嵌一个孔
func (h *Hero) IntoHole(holeID uint32, soulID string) *Hole {
	cfg := gamedata.GetAstrologySoulCfg(soulID)
	if nil == cfg {
		logs.Error(fmt.Sprintf("[Astrology] Astrology Hero IntoHole, get soul config nil"))
		return nil
	}

	oldHole := h.addHole(holeID, uint32(cfg.GetRareLevel()), 0)

	logs.Debug("[Astrology] Hero IntoHole, oldHole:%#v", oldHole)
	return oldHole
}
