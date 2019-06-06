package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdCommonConfig     ProtobufGen.CONFIG
	abstractCancelCost CostData
	equipTrickSwapCost CostData
	gdRankforGWC       map[uint32]uint32
)

func loadCommonConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.CONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	// 出错就直接崩溃吧
	gdCommonConfig = *lv_data[0]

	abstractCancelCost.AddItem(VI_Hc, gdCommonConfig.GetAbstractCancel())
	equipTrickSwapCost.AddItem(VI_Sc0, gdCommonConfig.GetExchangeCost())
	//logs.Trace("gdCommonConfig %v", gdCommonConfig)

}

func loadRankForGWC(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data := &ProtobufGen.RANKFORGWC_ARRAY{}
	err = proto.Unmarshal(buffer, data)
	errcheck(err)
	gdRankforGWC = make(map[uint32]uint32, len(data.GetItems()))
	for _, x := range data.GetItems() {
		gdRankforGWC[x.GetHeroID()] = x.GetParam()
	}

}

func GetRankForGwcParam(heroId uint32) uint32 {
	return gdRankforGWC[heroId]
}

func GetCommonCfg() *ProtobufGen.CONFIG {
	return &gdCommonConfig
}

// 服务器的装备数量包括身上穿了，客户端不包括，所有服务器增加50个
func GetEquipCountUpLimit() uint32 {
	return gdCommonConfig.GetInventoryCap() + 50
}

// 宝石数量限制，服务器限制200
func GetJadeCountUpLimit() uint32 {
	return gdCommonConfig.GetJadeLimit()
}

func GetAbstractCancelCost() *CostData {
	return &abstractCancelCost
}

func GetEquipTrickSwapCost() *CostData {
	return &equipTrickSwapCost
}
