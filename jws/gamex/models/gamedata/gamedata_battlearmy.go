package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	battleArmy            map[uint32]*ProtobufGen.BATTLEARMY
	battleArmyAnotherSave map[battleArmyData]*ProtobufGen.BATTLEARMY
	battleArmyLevel       map[battleArmyLevData]*ProtobufGen.BATTLEARMYLEVEL
)

func loadBattleArmy(filepath string) {
	buffer, err := loadBin(filepath)
	errCheck(err)
	dataList := &ProtobufGen.BATTLEARMY_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	battleArmy = make(map[uint32]*ProtobufGen.BATTLEARMY)
	battleArmyAnotherSave = make(map[battleArmyData]*ProtobufGen.BATTLEARMY)
	for _, v := range dataList.GetItems() {
		battleArmy[v.GetID()] = v
		battleArmyAnotherSave[battleArmyData{battleArmyNationality: v.GetBattleArmyNationality(), battleArmyLoc: v.GetBattleArmyLoc()}] = v
	}
}

type battleArmyLevData struct {
	battleArmyLoc   uint32
	battleArmyLevel uint32
}

type battleArmyData struct {
	battleArmyNationality uint32
	battleArmyLoc         uint32
}

func loadBattleArmyLevel(filepath string) {
	buffer, err := loadBin(filepath)
	errCheck(err)
	dataList := &ProtobufGen.BATTLEARMYLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	battleArmyLevel = make(map[battleArmyLevData]*ProtobufGen.BATTLEARMYLEVEL)
	for _, v := range dataList.GetItems() {
		battleArmyLevel[battleArmyLevData{battleArmyLoc: v.GetBattleArmyLoc(),
			battleArmyLevel: v.GetBattleArmyLevel()}] = v
	}
}

func GetBattleArmyNum() int {
	return len(battleArmy)
}
func GetBattleArmyByStruct(battleArmyNationality, battleArmyLoc uint32) *ProtobufGen.BATTLEARMY {
	return battleArmyAnotherSave[battleArmyData{battleArmyNationality: battleArmyNationality, battleArmyLoc: battleArmyLoc}]
}

func GetBattleArmy(ID int) *ProtobufGen.BATTLEARMY {
	return battleArmy[uint32(ID)]
}

func GetBattleArmyLevel(battleArmyLoc uint32, battleArmyLev uint32) *ProtobufGen.BATTLEARMYLEVEL {
	return battleArmyLevel[battleArmyLevData{battleArmyLoc: battleArmyLoc, battleArmyLevel: battleArmyLev}]
}
