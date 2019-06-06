package astrology

import (
	"math/rand"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/mn_selector"
)

//FactoryElem 星魂产出每个级别的MN状态
type FactoryElem struct {
	AugurLevel      uint32 `json:"lv,omitempty"`
	SpecialTryCount uint32 `json:"sc,omitempty"`

	MN *mnSelector.MNSelectorState `json:"mn,omitempty"`
}

//GoString GoStringer interface
func (f *FactoryElem) GoString() string {
	str := ""

	str += fmt.Sprintf("{AugurLevel:%v,SpecialTryCount:%v,MN:%#v}", f.AugurLevel, f.SpecialTryCount, f.MN)

	return str
}

//Factory 星魂产出(占星)
type Factory struct {
	CurrLevel uint32         `json:"curr_level,omitempty"`
	Augurs    []*FactoryElem `json:"augurs,omitempty"`

	mapAugurs map[uint32]*FactoryElem
}

//GoString GoStringer interface
func (f *Factory) GoString() string {
	str := ""

	str += fmt.Sprintf("{CurrLevel:%v,Augurs:{", f.CurrLevel)
	for _, elem := range f.Augurs {
		str += fmt.Sprintf("%#v,", elem)
	}
	str += "}}"

	return str
}

func newFactory() *Factory {
	return &Factory{
		CurrLevel: gamedata.GetAstrologyAugurMinLevel(),
		Augurs:    []*FactoryElem{},
		mapAugurs: map[uint32]*FactoryElem{},
	}
}

func newFactoryElem(lv uint32, rd *rand.Rand) *FactoryElem {
	elem := &FactoryElem{}

	elem.AugurLevel = lv
	elem.SpecialTryCount = 0

	cfg := gamedata.GetAstrologyAugurCfg(lv + 1)
	if nil != cfg {
		elem.newMN(cfg.GetLevelUpM(), cfg.GetLevelUpN(), cfg.GetLevelUpSpace(), rd)
	}

	return elem
}

func (f *FactoryElem) newMN(m, n, offset uint32, rd *rand.Rand) {
	mn := &mnSelector.MNSelectorState{}
	space := int64(m + offset - (rd.Uint32() % (2*offset + 1)))
	num := int64(n)
	mn.Init(num, space)
	f.MN = mn
}

//GetLoot 获取掉落表ID
func (f *FactoryElem) GetLoot(rd *rand.Rand) string {
	cfg := gamedata.GetAstrologyAugurCfg(f.AugurLevel)
	if cfg.GetSpecialLoot() == "" {
		return cfg.GetNormalLoot()
	}

	f.SpecialTryCount++

	if cfg.GetSpecialLimitMin() >= f.SpecialTryCount {
		return cfg.GetNormalLoot()
	}

	if cfg.GetSpecialLimitMax() <= f.SpecialTryCount {
		f.SpecialTryCount = 0
		return cfg.GetSpecialLoot()
	}

	if rd.Float32() <= cfg.GetSpecialLootRate() {
		f.SpecialTryCount = 0
		return cfg.GetSpecialLoot()
	}
	return cfg.GetNormalLoot()
}

//TryUp 尝试提升占星等级
func (f *Factory) TryUp(rd *rand.Rand) bool {
	// defer logs.Debug("[Astrology] FactoryElem TryUp, Factory after:%#v", f)
	//从顶级直接跳到1级
	if gamedata.GetAstrologyAugurMaxLevel() <= f.CurrLevel {
		f.CurrLevel = gamedata.GetAstrologyAugurMinLevel()
		return false
	}

	currElem := f.GetCurrFactoryElem(rd)

	if nil == currElem.MN {
		f.CurrLevel = gamedata.GetAstrologyAugurMinLevel()
		return false
	}

	if currElem.MN.IsNowNeedNewTurn() {
		cfg := gamedata.GetAstrologyAugurCfg(f.CurrLevel + 1)
		if nil != cfg {
			currElem.newMN(cfg.GetLevelUpM(), cfg.GetLevelUpN(), cfg.GetLevelUpSpace(), rd)

			logs.Debug("[Astrology] FactoryElem TryUp, newMN:%#v", currElem.MN)
		} else {
			currElem.MN = nil
			f.CurrLevel = gamedata.GetAstrologyAugurMinLevel()
			return false
		}
	}

	doUp := currElem.MN.Selector(rd)
	if true == doUp {
		f.CurrLevel++
		return true
	}

	f.CurrLevel = gamedata.GetAstrologyAugurMinLevel()
	return false
}

//afterLogin ..
func (f *Factory) afterLogin() {
	f.mapAugurs = map[uint32]*FactoryElem{}
	for _, augur := range f.Augurs {
		f.mapAugurs[augur.AugurLevel] = augur
	}
}

//GetCurrFactoryElem 取当前的占星项
func (f *Factory) GetCurrFactoryElem(rd *rand.Rand) *FactoryElem {
	elem := f.mapAugurs[f.CurrLevel]

	if nil == elem {
		elem = newFactoryElem(f.CurrLevel, rd)
		f.Augurs = append(f.Augurs, elem)
		f.mapAugurs[f.CurrLevel] = elem

		logs.Debug("[Astrology] Factory GetCurrFactoryElem, newFactoryElem:%#v", elem)
	}

	return elem
}
