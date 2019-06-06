package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdActivityGiftByTime []ActivityGiftByCondition
)

func loadActivityGiftByTime(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.CDTGIFTTIMING_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)

	corp_lv_data := data_ar.GetItems()
	gdActivityGiftByTime = make(
		[]ActivityGiftByCondition,
		len(corp_lv_data),
		len(corp_lv_data))

	for _, c := range corp_lv_data {
		idx := int(c.GetIndex()) - 1
		if idx < 0 || idx >= len(gdActivityGiftByTime) {
			panic(fmt.Errorf("gdActivityGiftByTime len Err %d", idx))
		}

		gdActivityGiftByTime[idx].ID = c.GetActivityID()
		gdActivityGiftByTime[idx].Desc = c.GetDesc()
		gdActivityGiftByTime[idx].Cond.Ctyp = c.GetFCType()
		gdActivityGiftByTime[idx].Cond.Param1 = int64(c.GetFCValue())
		for _, r := range c.GetGoalAward_Template() {
			gdActivityGiftByTime[idx].Reward.AddItem(r.GetReward(), r.GetCount())
		}

		//logs.Trace("gdActivityGiftByTime %d --> %v", idx, gdActivityGiftByTime[idx])
	}
}
