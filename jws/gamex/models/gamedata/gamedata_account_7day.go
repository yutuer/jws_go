package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
)

var (
	gd7DayQuestOpenDay       map[uint32]uint32
	gd7DayShop               map[uint32]*ProtobufGen.ACTIVITYSHOP
	gd7DayGoodHaveSeverCount map[uint32]uint32
	gd7DayGoodLimitCount     map[uint32]*ProtobufGen.ACTIVITYSHOP
	gd7DayDayCount           uint32
)

func loadAccount7DayQuest(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.ACTIVITYQUEST_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)

	gd7DayQuestOpenDay = make(map[uint32]uint32, len(data_ar.GetItems()))
	for _, data := range data_ar.GetItems() {
		gd7DayQuestOpenDay[data.GetQuestID()] = data.GetOpeningParameters()
		if data.GetOpeningParameters() > gd7DayDayCount {
			gd7DayDayCount = data.GetOpeningParameters()
		}
	}
}

func loadAccount7DayShop(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.ACTIVITYSHOP_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)

	gd7DayShop = make(map[uint32]*ProtobufGen.ACTIVITYSHOP, len(data_ar.GetItems()))
	gd7DayGoodHaveSeverCount = make(map[uint32]uint32, len(data_ar.GetItems()))
	gd7DayGoodLimitCount = make(map[uint32]*ProtobufGen.ACTIVITYSHOP, len(data_ar.GetItems()))
	for _, v := range data_ar.GetItems() {
		if v.GetServerCountLimit() > 0 && v.GetCountLimit() <= 0 {
			panic(fmt.Errorf("Account7DayShop good id %v ServerCountLimit() > 0 but CountLimit() <= 0", v.GetPromotionID()))
		}

		gd7DayShop[v.GetPromotionID()] = v
		if v.GetServerCountLimit() > 0 {
			gd7DayGoodHaveSeverCount[v.GetPromotionID()] = v.GetServerCountLimit()
		}
		if v.GetCountLimit() > 0 {
			gd7DayGoodLimitCount[v.GetPromotionID()] = v
		}
		if v.GetOpeningParameters() > gd7DayDayCount {
			gd7DayDayCount = v.GetOpeningParameters()
		}
	}
}

func GetAccount7DayGood(promotionID uint32) *ProtobufGen.ACTIVITYSHOP {
	return gd7DayShop[promotionID]
}

func GetAccount7DaySerGood() map[uint32]uint32 {
	return gd7DayGoodHaveSeverCount
}

func GetAccount7DayLimitCount() map[uint32]*ProtobufGen.ACTIVITYSHOP {
	return gd7DayGoodLimitCount
}

func GetAccount7DayOverTime(profileCreateTime int64) int64 {
	ct := GetCommonDayBeginSec(profileCreateTime)
	return ct + int64(GetAccount7DaySumDays()*util.DaySec)
}

func GetAccount7DayQuestOpenDay(qid uint32) uint32 {
	return gd7DayQuestOpenDay[qid]
}

func GetAccount7DaySumDays() int {
	return int(gd7DayDayCount)
}
