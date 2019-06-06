package gamedata

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdGuildSignCost         []PriceData
	gdGuildSignGive         []PriceData
	gdGuildSignGuildXPAddon []uint32

	gdGuildSignNewGive    []PriceDatas
	gdGuildSignNewXPAddon []uint32
)

func loadGuildSignData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDSIGN_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	if len(data) < 2 {
		panic(errors.New("GuildSignDataLenErr"))
	}

	gdGuildSignCost = make([]PriceData, 0, len(data))
	gdGuildSignGive = make([]PriceData, 0, len(data))
	gdGuildSignGuildXPAddon = make([]uint32, 0, len(data))

	gdGuildSignNewGive = make([]PriceDatas, 0, len(data))

	for i := 0; i < len(data); i++ {
		d := data[i]
		cost := PriceData{}
		give := PriceData{}
		cost.AddItem(d.GetGuildCoin(), d.GetPrice())
		gdGuildSignCost = append(gdGuildSignCost, cost)
		give.AddItem(VI_GC, d.GetReward())
		gdGuildSignGive = append(gdGuildSignGive, give)
		gdGuildSignGuildXPAddon = append(gdGuildSignGuildXPAddon, d.GetContribution())

		newGive := PriceDatas{}
		for _, reward := range d.GetLoot_Table() {
			newGive.AddItem(
				reward.GetLootID(),
				reward.GetLootNum())
			if reward.GetLootID() == VI_GuildXP {
				gdGuildSignNewXPAddon = append(gdGuildSignNewXPAddon, reward.GetLootNum())
			}
		}
		gdGuildSignNewGive = append(gdGuildSignNewGive, newGive)

	}

	logs.Trace("gdGuildPosNumMax %v", gdGuildPosNumMax)
}

func GetGuildSignNewInfo(id int) (ok bool, cost PriceData, give PriceDatas, xp uint32) {
	ok, cost, give = false, PriceData{}, PriceDatas{}
	if id < 0 || id >= len(gdGuildSignCost) {
		return
	}
	ok = true
	cost = gdGuildSignCost[id]
	give = gdGuildSignNewGive[id]
	xp = gdGuildSignNewXPAddon[id]
	return
}

func GetGuildSignInfo(id int) (ok bool, cost, give PriceData, xp uint32) {
	ok, cost, give, xp = false, PriceData{}, PriceData{}, 0
	if id < 0 || id >= len(gdGuildSignCost) {
		return
	}
	ok = true
	cost = gdGuildSignCost[id]
	give = gdGuildSignGive[id]
	xp = gdGuildSignGuildXPAddon[id]
	return
}
