package astrology

import (
	"math/rand"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

//DebugAugurUpStatistic ..
func (a *Astrology) DebugAugurUpStatistic(count int) (map[uint32]uint32, map[string]uint32) {
	factory := newFactory()

	rd := rand.New(rand.NewSource(time.Now().Unix()))

	giveDatas := gamedata.NewPriceDatas(1)
	statis := map[uint32]uint32{}
	for i := 0; i < count; i++ {
		statis[factory.CurrLevel] = statis[factory.CurrLevel] + 1

		elem := factory.GetCurrFactoryElem(rd)
		loot := elem.GetLoot(rd)
		giveData, err := gamedata.LootTemplateRand(rd, loot)
		if nil != err {
			continue
		}
		giveDatas.AddOther(&giveData)
		factory.TryUp(rd)
	}

	goods := map[string]uint32{}

	for i, id := range giveDatas.Cost.Items {
		goods[id] = giveDatas.Cost.Count[i]
	}

	return statis, goods
}
