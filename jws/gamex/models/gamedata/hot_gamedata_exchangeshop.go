package gamedata

import (
	"math/rand"

	"fmt"

	"sort"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	exchangeShopExchangeType     = 1
	exchangeShopAutoExchangeType = 2
)

type hotExchangeShopData struct {
	exchangePropData *ProtobufGen.HOTSHOP_ARRAY
}

func (hes *hotExchangeShopData) loadData(buffer []byte, datas *HotDatas) error {
	hes.exchangePropData = &ProtobufGen.HOTSHOP_ARRAY{}
	if err := proto.Unmarshal(buffer, hes.exchangePropData); err != nil {
		return err
	}

	for id, _ := range hes.getExistActivityID() {
		showData := hes.GetExchangePropShowData(id)
		index := make([]int, 0)
		for _, data := range showData {
			index = append(index, int(data.GetIndex()))
		}
		sort.Ints(index)
		var curIndex int = 1
		for _, item := range index {
			if item != curIndex {
				return fmt.Errorf("hotexchangeshopdata index format error")
			}
			curIndex++
		}
	}
	datas.HotExchangeShopData = *hes

	return nil
}

func (hes *hotExchangeShopData) getExistActivityID() (ret map[uint32]struct{}) {
	ret = make(map[uint32]struct{}, 0)
	for _, item := range hes.exchangePropData.GetItems() {
		ret[item.GetActivityID()] = struct{}{}
	}
	return ret
}

func (hes *hotExchangeShopData) GetHotExchangeShopData() *ProtobufGen.HOTSHOP_ARRAY {
	return hes.exchangePropData
}

func (hes *hotExchangeShopData) GetCanAutoExchangePropData(activityID uint32) (ret []*ProtobufGen.HOTSHOP) {
	ret = make([]*ProtobufGen.HOTSHOP, 0)
	for _, item := range hes.exchangePropData.GetItems() {
		if item.GetChangeType() == exchangeShopAutoExchangeType && item.GetActivityID() == activityID {
			ret = append(ret, item)
		}
	}
	return ret
}

func (hes *hotExchangeShopData) GetExchangePropShowData(activityID uint32) (ret []*ProtobufGen.HOTSHOP) {
	ret = make([]*ProtobufGen.HOTSHOP, 0)
	for _, item := range hes.exchangePropData.GetItems() {
		if item.GetChangeType() == exchangeShopExchangeType && item.GetActivityID() == activityID {
			ret = append(ret, item)
		}
	}
	return ret
}

func (hes *hotExchangeShopData) GetExchangePropData(index, activityID uint32) *ProtobufGen.HOTSHOP {
	for _, item := range hes.exchangePropData.GetItems() {
		if item.GetIndex() == index && item.GetActivityID() == activityID {
			return item
		}
	}
	return nil
}

type hotStageLootExchangeData struct {
	stageLootExchangeData *ProtobufGen.FALL_ARRAY
}

func (hsle *hotStageLootExchangeData) loadData(buffer []byte, datas *HotDatas) error {
	hsle.stageLootExchangeData = &ProtobufGen.FALL_ARRAY{}
	if err := proto.Unmarshal(buffer, hsle.stageLootExchangeData); err != nil {
		return err
	}
	datas.HotStageLootExchangeData = *hsle
	return nil
}

func (hsle *hotStageLootExchangeData) GetStageLootExchangeData() *ProtobufGen.FALL_ARRAY {
	return hsle.stageLootExchangeData
}

func (hsle *hotStageLootExchangeData) convertStageType(chapterType uint32) uint32 {
	switch chapterType {
	case LEVEL_TYPE_MAIN:
		return 1
	case LEVEL_TYPE_ELITE:
		return 2
	case LEVEL_TYPE_HELL:
		return 3
	default:
		return 1
	}
}

func (hsle *hotStageLootExchangeData) GetStageLootExchangeProp(activityID uint32,
	chapterID string, chapterType uint32, level uint32) *ProtobufGen.HOTNORMALGACHA {
	var ret *ProtobufGen.HOTNORMALGACHA
	var data *ProtobufGen.FALL
	typ := hsle.convertStageType(chapterType)
	for _, item := range hsle.stageLootExchangeData.GetItems() {
		if (item.GetChapterID() == chapterID ||
			(item.GetChapterID() == "" && item.GetChapterType() == typ)) &&
			item.GetActivityID() == activityID &&
			level <= item.GetLevelMax() && level >= item.GetLevelMin() {
			data = item
			break
		}
	}
	logs.Debug("get stage loot group data: %v", data)
	if data != nil {
		groupID := data.GetExtraItemGroupID()
		ret = GetRandWeightProp(groupID)
	}
	logs.Debug("get stage loot exchange prop, activityID: %d, chapterID: %s, rewards: %v",
		activityID, chapterID, ret)
	return ret
}

func GetRandWeightProp(groupID uint32) *ProtobufGen.HOTNORMALGACHA {
	propDatas := GetHotDatas().Activity.GetActivityGachaNormal(groupID)
	var weight int32
	for _, item := range propDatas {
		weight += int32(item.GetWeight())
	}
	randWeight := rand.Int31n(weight)
	var curWeight int32
	logs.Debug("randWeight %d", randWeight)
	for _, item := range propDatas {
		curWeight += int32(item.GetWeight())
		if curWeight > randWeight {
			return item
		}
	}
	return nil
}
