package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gbWorshipBaseRewardCfg    []*ProtobufGen.GUILDWORSHIPREWARD
	gbWorshipBaseBoxRewardCfg []*ProtobufGen.GUILDWORSHIPBOX
	gbWorshipCost             map[uint32]uint32
	gbWorshipCrit             map[uint32]uint32
)

func loadGuildWorshipData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDWORSHIPREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	gbWorshipBaseRewardCfg = make([]*ProtobufGen.GUILDWORSHIPREWARD, 0, len(data))

	for _, e := range ar.GetItems() {
		gbWorshipBaseRewardCfg = append(gbWorshipBaseRewardCfg, e)
	}

}

func loadGuildWorshipBoxRewardData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDWORSHIPBOX_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	gbWorshipBaseBoxRewardCfg = make([]*ProtobufGen.GUILDWORSHIPBOX, 0, len(data))

	for _, e := range ar.GetItems() {
		gbWorshipBaseBoxRewardCfg = append(gbWorshipBaseBoxRewardCfg, e)
	}
}

func loadGuildWorshipCritData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDWORSHIPCRIT_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)
	data := ar.GetItems()
	gbWorshipCrit = make(map[uint32]uint32, len(data))

	for _, e := range data {
		gbWorshipCrit[e.GetWorshipDrawID()] = e.GetWorshipDrawCrit()
	}

}

func loadGuildWorshipCostData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDWORSHIPCOST_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)
	data := ar.GetItems()
	gbWorshipCost = make(map[uint32]uint32, len(data))

	for _, e := range data {
		gbWorshipCost[e.GetWorshipTimes()] = e.GetWorshipCost()
	}

}

func GetGuildWorshipRewardCost(id int64) uint32 {
	return gbWorshipCost[uint32(id)]

}

func GetguildWorshipReward(id int) []*ProtobufGen.GUILDWORSHIPREWARD_LootRule {
	return gbWorshipBaseRewardCfg[id].GetLoot_Table()
}

func GetguildWorshipBoxReward(id int) []*ProtobufGen.GUILDWORSHIPBOX_LootRule {
	return gbWorshipBaseBoxRewardCfg[id].GetLoot_Table()
}

func IsGuildWorshipCrit(num int64) bool {
	worshipDrawCrit := gbWorshipCrit[uint32(num)]
	x := rander.Int63n(100)
	if x <= int64(worshipDrawCrit) {
		return true
	} else {
		return false
	}

}
