package gamedata

import "vcs.taiyouxi.net/jws/gamex/protogen"

var (
	gdOPPOSignData       map[uint32][]*ProtobufGen.OPPOSEVENDAYS_LootRule
	gdOPPODailyQuestData []*ProtobufGen.OPPOEVERYDAY_LootRule
	oppoMaxSignDay       int
)

func loadOPPOSignData(filepath string) {
	ar := &ProtobufGen.OPPOSEVENDAYS_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	gdOPPOSignData = make(map[uint32][]*ProtobufGen.OPPOSEVENDAYS_LootRule, len(data))
	oppoMaxSignDay = len(data)
	for _, v := range data {
		gdOPPOSignData[v.GetLastDay()] = v.GetLoot_Table()
	}
}

func loadOPPODailyQuestData(filepath string) {
	ar := &ProtobufGen.OPPOEVERYDAY_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	gdOPPODailyQuestData = data[0].GetLoot_Table()
}

func GetOPPOSignData(day int) []*ProtobufGen.OPPOSEVENDAYS_LootRule {
	if day > oppoMaxSignDay {
		day = oppoMaxSignDay
	}
	return gdOPPOSignData[uint32(day)]
}

func GetOPPODailyQuestData() []*ProtobufGen.OPPOEVERYDAY_LootRule {
	return gdOPPODailyQuestData
}
