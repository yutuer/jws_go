package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdHitEggCost   map[uint32]*ProtobufGen.REBATECOST
	gdHitEggReward map[uint32]*ProtobufGen.REBATEREWARD
	gdEggNum       int
)

func loadHitEggCost(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.REBATECOST_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	data := ar.GetItems()
	gdHitEggCost = make(map[uint32]*ProtobufGen.REBATECOST, len(data))
	for _, v := range data {
		gdHitEggCost[v.GetCostTime()] = v
	}
}

func loadHitEggReward(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.REBATEREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	data := ar.GetItems()
	gdHitEggReward = make(map[uint32]*ProtobufGen.REBATEREWARD, len(data))
	for _, v := range data {
		gdHitEggReward[v.GetRewardTier()] = v
		gdEggNum = len(v.GetLoots()) + 1
	}
}

func GetHitEggCost(idx uint32) *ProtobufGen.REBATECOST {
	return gdHitEggCost[idx]
}

func GetHitEggReward(idx uint32) *ProtobufGen.REBATEREWARD {
	return gdHitEggReward[idx]
}

func HitEggDailyEggCount() int {
	return len(gdHitEggCost)
}

func EggCountInAGame() int {
	return gdEggNum
}

func HitEggInitWeight(idx uint32) []uint32 {
	var sw uint32
	ws := make([]uint32, 0, gdEggNum)
	cfg := gdHitEggReward[idx]
	ws = append(ws, cfg.GetSpecialLootWeight())
	sw += cfg.GetSpecialLootWeight()
	for _, v := range cfg.GetLoots() {
		sw += v.GetWeight()
		ws = append(ws, sw)
	}
	return ws
}
