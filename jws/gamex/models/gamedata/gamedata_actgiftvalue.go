package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

type ActivityGiftByConds struct {
	Gifts     []ActivityGiftByCondition
	TimeBegin int64
	TimeEnd   int64
	Title     string
	Desc      string
	ID        uint32
	Index     int
}

var (
	gdActivityGiftByConds    []ActivityGiftByConds
	gdActivityGiftByCondsMap map[uint32]*ActivityGiftByConds
)

func GetActivityGiftByCondData(ID, typ uint32, now_t int64) *ActivityGiftByCondition {
	conds, ok := gdActivityGiftByCondsMap[typ]
	if !ok || conds == nil {
		return nil
	}

	if now_t < conds.TimeBegin || now_t > conds.TimeEnd {
		return nil
	}

	id := int(ID)
	if id < 0 || id >= len(conds.Gifts) {
		return nil
	}

	return &conds.Gifts[id]
}

func loadActivityGiftByCond(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.CDTGIFTVALUE_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)

	corp_lv_data := data_ar.GetItems()
	for _, c := range corp_lv_data {
		act, ok := gdActivityGiftByCondsMap[c.GetActivityID()]

		if !ok {
			panic(fmt.Errorf("gdActivityGiftByCondsMap No Info By %d", c.GetActivityID()))
		}

		nAct := ActivityGiftByCondition{}
		nAct.loadFromData(c)
		act.Gifts = append(act.Gifts, nAct)
	}
}
