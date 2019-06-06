package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
)

var (
	gdTrickRuleRandSets    util.RandSetMap
	gdTrickDetailAttrAddon map[string]*avatarAttrAddon
)

// 注意：表中的池长度最大值这里获取不到，所以用常数，如果调整需要修改这里
const MAX_POOL_SIZE = 128

// 装备生成表的权总值
const TrickRule_Power_Max = 10000

func GetTrickDetailAttrAddon(id string) *avatarAttrAddon {
	tg, ok := gdTrickDetailAttrAddon[id]
	if ok {
		return tg
	} else {
		return nil
	}
}

func GetTrickRuleRand() *util.RandSetMap {
	return &gdTrickRuleRandSets
}

const MAX_Trick_RandSet_Typ = 2

var (
	gdTrickRandPool map[string][]util.RandSet
)

func loadTrickRandPool(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.EUIPTRICKRULE_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	trick_rule_data := ar.GetItems()

	gdTrickRandPool = make(map[string][]util.RandSet,
		len(trick_rule_data))

	for _, t := range trick_rule_data {
		id := t.GetTrickRuleID()
		typ := int(t.GetTrickType())

		res, ok := gdTrickRandPool[id]
		if !ok || res == nil {
			new_randsets := make([]util.RandSet, 0, MAX_Trick_RandSet_Typ)
			for i := 0; i < MAX_Trick_RandSet_Typ; i++ {
				n := util.RandSet{}
				n.Init(64)
				new_randsets = append(new_randsets, n)
			}
			gdTrickRandPool[id] = new_randsets
		}

		gdTrickRandPool[id][typ].Add(t.GetTrickID(), t.GetWeight())

		//logs.Trace("Add New TRICKDETAIL %s : %v", id, gdTrickRandPool[id][typ])
	}

	for _, rsets := range gdTrickRandPool {
		for i := 0; i < MAX_Trick_RandSet_Typ; i++ {
			rsets[i].Make()
			//logs.Trace("rset %s %d --> %v", rk, i, rsets[i])
		}
	}

	//logs.Trace("gdTrickRandPool %v", gdTrickRandPool)
}

func GetTrickRandPool(trick_id string) []util.RandSet {
	res, ok := gdTrickRandPool[trick_id]
	if ok {
		return res[:]
	} else {
		return nil
	}
}
