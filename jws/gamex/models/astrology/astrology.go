package astrology

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Astrology ..
type Astrology struct {
	Bag     *SoulBag `json:"bag,omitempty"`
	Factory *Factory `json:"factory,omitempty"`
	Heros   []*Hero  `json:"heros,omitempty"`

	mapHeros map[uint32]*Hero
}

//GoString GoStringer interface
func (a *Astrology) GoString() string {
	str := ""

	str += fmt.Sprintf("{Factory:%#v,Bag:%#v,Heros:{", a.Factory, a.Bag)
	for _, hero := range a.Heros {
		str += fmt.Sprintf("%#v,", hero)
	}
	str += "}}"

	return str
}

//NewAstrology ..
func NewAstrology() *Astrology {
	return &Astrology{
		Heros:    []*Hero{},
		mapHeros: map[uint32]*Hero{},
	}
}

//AfterLogin ..
func (a *Astrology) AfterLogin() {
	if nil != a.Bag {
		a.Bag.afterLogin()
	}

	if nil != a.Factory {
		a.Factory.afterLogin()
	}

	a.mapHeros = map[uint32]*Hero{}
	if nil != a.Heros {
		for _, hero := range a.Heros {
			a.mapHeros[hero.HeroID] = hero
			hero.afterLogin()
		}
	}

	logs.Debug("[Astrology] AfterLogin, Astrology:%#v", a)
}

//GetBag ..
func (a *Astrology) GetBag() *SoulBag {
	if nil == a.Bag {
		a.Bag = newSoulBag()
	}
	return a.Bag
}

//GetFactory ..
func (a *Astrology) GetFactory() *Factory {
	if nil == a.Factory {
		a.Factory = newFactory()
	}
	return a.Factory
}

//GetHeros ..
func (a *Astrology) GetHeros() []*Hero {
	return a.Heros
}

//GetHero ..
func (a *Astrology) GetHero(id uint32) *Hero {
	hero, exist := a.mapHeros[id]
	if false == exist {
		hero = newHero(id)
		a.Heros = append(a.Heros, hero)
		a.mapHeros[id] = hero

		logs.Debug("[Astrology] Astrology GetHero, newHero:%#v", hero)
	}
	return hero
}

//CheckHero ..
func (a *Astrology) CheckHero(id uint32) *Hero {
	return a.mapHeros[id]
}

//ClearData ..
func (a *Astrology) ClearData() {
	a.Bag = nil
	a.Factory = nil
	a.Heros = []*Hero{}
	a.mapHeros = map[uint32]*Hero{}
}

//ClearBag ..
func (a *Astrology) ClearBag() {
	a.Bag = nil
}

//CalculateHerosSoulExp ..
func (a *Astrology) CalculateHerosSoulExp() uint32 {
	sum := uint32(0)
	for _, hero := range a.Heros {
		for _, hole := range hero.Holes {
			ms := gamedata.AstrologyTranslateSoulToMaterial(hole.HoleID, hole.Rare, hole.Upgrade)
			sum += ms[gamedata.VI_SSXP]
		}
	}
	return sum
}
