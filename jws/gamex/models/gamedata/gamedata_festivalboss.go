package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	lvlOne = 1
	lvlTwo = 2
)

var (
	gdFestivallBossCfgData  map[uint32]*ProtobufGen.FBCONFIG
	gdFbReward              map[uint32]*ProtobufGen.FBREWARD
	gdFbRwardTemple         []string
	gdFbRwardTempleCount    []uint32
	gdShopRewardCfg         []*ProtobufGen.FBSHOP
	gdShopRewardTemple      []string
	gdShopRewardTempleCount []uint32
	gdNeedItemId            []string
	gdNeedItemCount         []uint32
)

func loadFestivallBossCfgData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.FBCONFIG_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))
	gdFestivallBossCfgData = make(map[uint32]*ProtobufGen.FBCONFIG, len(ar.Items))
	for _, e := range ar.GetItems() {
		gdFestivallBossCfgData[e.GetActivityID()] = e
	}
}

func loadFestivallBossLootData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.FBREWARD_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))
	gdFbReward = make(map[uint32]*ProtobufGen.FBREWARD, len(ar.Items))

	for _, e := range ar.GetItems() {
		gdFbReward[e.GetActivityID()] = e
	}
}

func loadFestivallShopData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.FBSHOP_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))
	gdShopRewardCfg = make([]*ProtobufGen.FBSHOP, len(ar.Items))

	for _, e := range ar.GetItems() {
		gdShopRewardCfg = append(gdShopRewardCfg, e)
	}
}

func GetFestivallBossCfg(festivalid uint32) *ProtobufGen.FBCONFIG {
	return gdFestivallBossCfgData[festivalid]
}

func GetFestivallBossLootCfg(festivalid uint32) *ProtobufGen.FBREWARD {
	return gdFbReward[festivalid]
}

func GetFestivalBossReward(festivalid uint32, bosslvl int64) (multiple uint32, costtype string, costnum uint32) {
	switch bosslvl {
	case lvlOne:
		return gdFestivallBossCfgData[festivalid].GetMoreRewardOne(), gdFestivallBossCfgData[festivalid].GetMoreRewardOneCost(), gdFestivallBossCfgData[festivalid].GetMoreRewardOneCostNum()
	case lvlTwo:
		return gdFestivallBossCfgData[festivalid].GetMoreRewardTwo(), gdFestivallBossCfgData[festivalid].GetMoreRewardTwoCost(), gdFestivallBossCfgData[festivalid].GetMoreRewardTwoCostNum()
	}
	return
}

func GetFestivalBossCostChallengeCost(festivalid int64) (string, uint32) {
	return gdFestivallBossCfgData[uint32(festivalid)].GetChallengeCost(), gdFestivallBossCfgData[uint32(festivalid)].GetChallengeCostNum()
}

func GetFestivalGameRewardCfg(festivalid uint32) ([]string, []uint32) {
	e := GetFestivallBossLootCfg(festivalid)
	gdFbRwardTemple = make([]string, 0, len(e.GetLoot_Table()))
	gdFbRwardTempleCount = make([]uint32, 0, len(e.GetLoot_Table()))
	for _, loot := range e.GetLoot_Table() {
		gdFbRwardTemple = append(gdFbRwardTemple, loot.GetLootTemplateID())
		gdFbRwardTempleCount = append(gdFbRwardTempleCount, loot.GetLootTime())
	}

	return gdFbRwardTemple[:], gdFbRwardTempleCount[:]
}

func GetFestivalShopRewardCfg() []*ProtobufGen.FBSHOP {
	return gdShopRewardCfg
}

func GetFestivalShopReward(festivalid uint32, goods uint32) ([]string, []uint32, []string, []uint32) {
	e := GetFestivalShopRewardCfg()
	for _, v := range e {
		if v.GetActivityID() == festivalid && v.GetGoodsID() == goods {
			gdShopRewardTemple = make([]string, 0, len(v.GetLoot_Table()))
			gdShopRewardTempleCount = make([]uint32, 0, len(v.GetLoot_Table()))
			gdNeedItemId = make([]string, 0, len(v.GetNeed_Table()))
			gdNeedItemCount = make([]uint32, 0, len(v.GetNeed_Table()))
			for _, r := range v.GetLoot_Table() {
				gdShopRewardTemple = append(gdShopRewardTemple, r.GetLootGroupID())
				gdShopRewardTempleCount = append(gdShopRewardTempleCount, r.GetLootTime())
			}
			for _, r := range v.GetNeed_Table() {
				gdNeedItemId = append(gdNeedItemId, r.GetItemID())
				gdNeedItemCount = append(gdNeedItemCount, r.GetItemNum())
			}

		}
	}
	return gdShopRewardTemple[:], gdShopRewardTempleCount[:], gdNeedItemId[:], gdNeedItemCount[:]
}

func GetFestivalShopGoodsCount(festivalid uint32) int64 {
	e := GetFestivalShopRewardCfg()
	var goodsCount int64 = 0
	for _, v := range e {
		if v.GetActivityID() == festivalid {
			goodsCount += 1

		}
	}
	return goodsCount

}

func GetFestivalShopMaxRewardTime(festivalid uint32, goods uint32) uint32 {
	e := GetFestivalShopRewardCfg()
	for _, v := range e {
		if v.GetActivityID() == festivalid && v.GetGoodsID() == goods {
			return v.GetLimitTime()
		}
	}
	return 0
}
