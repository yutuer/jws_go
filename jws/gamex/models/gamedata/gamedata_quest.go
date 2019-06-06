package gamedata

import (
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdQuestDetailed map[uint32]*ProtobufGen.Quest
	gdQuestGiveData map[uint32]*givesData
	gdQuestInit     []uint32
	/*
		对于任务，需要检测是否达到可接取条件
		很多任务不是通过可接取条件触发接取的，所以这个列表中剔除了不需检测的任务
	*/
	gdQuestPostQuest     map[uint32][]uint32
	gdQuestNeedCheck     []*ProtobufGen.Quest
	gdQuestNeedCheck_M   map[uint32]*ProtobufGen.Quest
	gdDailyTask          [helper.Quest_Typ_count][]uint32
	gdQuestNeedAutoAccep []uint32
	gdQuestNeedAutoIdx   []uint32
)

func loadQuestDetailed(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.Quest_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdQuestDetailed = make(
		map[uint32]*ProtobufGen.Quest,
		len(data))

	gdQuestNeedCheck = make(
		[]*ProtobufGen.Quest,
		0,
		len(data))

	gdQuestNeedCheck_M = make(
		map[uint32]*ProtobufGen.Quest,
		len(data))

	gdQuestGiveData = make(
		map[uint32]*givesData,
		len(data))

	gdQuestPostQuest = make(
		map[uint32][]uint32,
		len(data))
	gdQuestNeedAutoAccep = make(
		[]uint32,
		0,
		len(data))
	for i := 0; i < helper.Quest_Typ_count; i++ {
		gdDailyTask[i] = []uint32{}
	}

	for _, c := range data {
		gdQuestDetailed[c.GetQuestID()] = c
		if c.GetAutoAccept() == 1 {
			gdQuestNeedAutoAccep = append(gdQuestNeedAutoAccep, c.GetFCType())
			gdQuestNeedAutoIdx = append(gdQuestNeedAutoIdx, c.GetQuestID())
		}

		ac_condition := c.GetAccCon_Table()
		if len(ac_condition) > 0 {
			cond := ac_condition[0]
			if cond.GetACType() != 999 {
				gdQuestNeedCheck = append(gdQuestNeedCheck, c)
				gdQuestNeedCheck_M[c.GetQuestID()] = c
				//logs.Warn("quest need check %d --> %v", c.GetQuestID(), c)
			}
		}

		give_data := givesData{}
		give_data.AddItem(VI_XP, c.GetCharaXP())
		give_data.AddItem(VI_CorpXP, c.GetCorpXP())
		give_data.AddItem(VI_Sc0, c.GetSC())
		if c.GetItem1() != "" && c.GetCount1() != 0 {
			give_data.AddItem(c.GetItem1(), c.GetCount1())
		}
		if c.GetItem2() != "" && c.GetCount2() != 0 {
			give_data.AddItem(c.GetItem2(), c.GetCount2())
		}
		if c.GetItem3() != "" && c.GetCount3() != 0 {
			give_data.AddItem(c.GetItem3(), c.GetCount3())
		}

		if c.GetItem4() != "" && c.GetCount4() != 0 {
			give_data.AddItem(c.GetItem4(), c.GetCount4())
		}
		gdQuestGiveData[c.GetQuestID()] = &give_data

		gdDailyTask[int(c.GetQuestType())] =
			append(gdDailyTask[int(c.GetQuestType())], c.GetQuestID())

		post_ar := strings.Split(c.GetPostQuest(), ",")
		post_id_ar := make([]uint32, 0, 8)
		for _, post := range post_ar {
			post_id, err := strconv.Atoi(post)
			if err != nil {
				//logs.Error("GetPostQuest Err by %d -> %v", c.GetQuestID(), post_ar)
				continue
			} else {
				post_id_ar = append(post_id_ar, uint32(post_id))
			}
		}
		gdQuestPostQuest[c.GetQuestID()] = post_id_ar
	}
}

func loadQuestInit(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.BORNQUEST_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdQuestInit = make(
		[]uint32,
		0,
		len(data))

	for _, c := range data {
		gdQuestInit = append(gdQuestInit, c.GetBornQuestID())
	}
	//logs.Info("gdQuestInit %v", gdQuestInit)
}

func GetQuestData(id uint32) *ProtobufGen.Quest {
	res, ok := gdQuestDetailed[id]
	if !ok {
		return nil
	} else {
		return res
	}
}

func GetPostQuest(id uint32) []uint32 {
	res, ok := gdQuestPostQuest[id]
	if !ok {
		return nil
	} else {
		return res[:]
	}
}

func GetQuestNeedCheck() []*ProtobufGen.Quest {
	return gdQuestNeedCheck[:]
}

func GetQuestNeedCheckById(questId uint32) *ProtobufGen.Quest {
	return gdQuestNeedCheck_M[questId]
}

func GetQuestInit() []uint32 {
	return gdQuestInit[:]
}

func GetQuestGiveData(id uint32) *givesData {
	res, ok := gdQuestGiveData[id]
	if !ok {
		return nil
	} else {
		return res
	}
}

func GetDailyTaskQuestId(typ int) []uint32 {
	return gdDailyTask[typ][:]
}
func GetAutoAccep() []uint32 {
	return gdQuestNeedAutoAccep[:]
}
func GetAutoAccepIdx() []uint32 {
	return gdQuestNeedAutoIdx[:]
}
