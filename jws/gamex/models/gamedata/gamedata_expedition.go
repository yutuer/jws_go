package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdExpeditionData          []*ProtobufGen.EXPEDITIONLEVEL
	gdExpeditionPassAwardData []*ProtobufGen.PASSAWARD
	gdExpeditionRewardData    []*ProtobufGen.EXPEDITIONREWARD
	gdExpeditionPassAward     map[uint32]*ProtobufGen.PASSAWARD
	gdExpeditionCfg           []*ProtobufGen.EXPEDITIONCONFIG
	gdExpeditionSweep         map[uint32]float32
)

func loadExpeditionConfig(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.EXPEDITIONCONFIG_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	gdExpeditionCfg = ar.GetItems()
}

func loadExpeditionSweep(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.EXPEDITIONSWEEP_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	gdExpeditionSweep = make(map[uint32]float32, len(ar.Items))
	for _, e := range ar.Items {
		logs.Debug("GSD %d Cost %d", e.GetGSDisparity(), e.GetHPCost())
		gdExpeditionSweep[e.GetGSDisparity()] = e.GetHPCost()
	}

}

func loadExpeditionData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.EXPEDITIONLEVEL_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	gdExpeditionData = ar.GetItems()
}

func loadExpeditionRewardData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.EXPEDITIONREWARD_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	gdExpeditionRewardData = ar.GetItems()

}

func loadExpeditionPassAwardData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.PASSAWARD_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	gdExpeditionPassAward = make(map[uint32]*ProtobufGen.PASSAWARD, len(ar.Items))
	for _, e := range ar.Items {
		gdExpeditionPassAward[e.GetPassLevel()] = e
	}
	gdExpeditionPassAwardData = ar.GetItems()
}

func GetAwardByStep(step int64) (string, string, string) {
	return gdExpeditionData[step].Loot_Table[0].GetItemID(),
		gdExpeditionData[step].Loot_Table[1].GetItemID(),
		gdExpeditionData[step].Loot_Table[2].GetItemID()
}

func GetAwardNumByStep(step int64) (uint32, uint32, uint32) {
	return gdExpeditionData[step].Loot_Table[0].GetItemNum(),
		gdExpeditionData[step].Loot_Table[1].GetItemNum(),
		gdExpeditionData[step].Loot_Table[2].GetItemNum()
}

func GetExpeditionLvlCfgs() []*ProtobufGen.EXPEDITIONLEVEL {
	return gdExpeditionData
}

func GetExpeditionAwardGive(awardlevel int) float32 {
	return gdExpeditionRewardData[awardlevel].GetGetUp()

}

func GetExpeditionAwardCost(awardlevel int) uint32 {
	return gdExpeditionRewardData[awardlevel].GetCostNum()

}

func GetExpeditionPassAwardCfgs(passlvl uint32) *ProtobufGen.PASSAWARD {
	return gdExpeditionPassAward[passlvl]
}

func GetExpeditionPassAward() []*ProtobufGen.PASSAWARD {
	return gdExpeditionPassAwardData
}

func GetExpeditionCfg() *ProtobufGen.EXPEDITIONCONFIG {
	return gdExpeditionCfg[0]
}

func GetExpeditionSweep(id uint32) float32 {
	return gdExpeditionSweep[id]
}
