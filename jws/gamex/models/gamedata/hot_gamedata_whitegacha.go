package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type WhiteGachaSetings struct {
}

func (gs *WhiteGachaSetings) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.HOTGACHASETTINGS_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWhiteGachaSetings = make(map[uint32]*ProtobufGen.HOTGACHASETTINGS, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWhiteGachaSetings[r.GetActivityID()] = r
	}

	logs.Debug("Load Hot Data WhiteGachaSetings Success")
	return nil
}

type WhiteNormalGacha struct {
}

func (nl *WhiteNormalGacha) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.HOTNORMALGACHA_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWhiteNormalGacha = make(map[uint32]NormalGachaConfig, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWhiteNormalGacha[r.GetGachaID()] = append(datas.Activity.gdWhiteNormalGacha[r.GetGachaID()], r)
	}

	logs.Debug("Load Hot Data WhiteNormalGacha Success")
	return nil
}

type WhiteGachaSpecil struct {
}

func (nl *WhiteGachaSpecil) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.HOTGACHASPECIAL_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWhiteGachaSpecil = make(map[uint32][]*ProtobufGen.HOTGACHASPECIAL, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWhiteGachaSpecil[r.GetActivityID()] = append(datas.Activity.gdWhiteGachaSpecil[r.GetActivityID()], r)
	}

	logs.Debug("Load Hot Data WhiteSpecilGacha Success")
	return nil
}

type WhiteGachaLowest struct {
}

func (nl *WhiteGachaLowest) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.HOTGACHALOWEST_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWhiteGachaLowest = make(map[uint32][]*ProtobufGen.HOTGACHALOWEST, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWhiteGachaLowest[r.GetActivityID()] = append(datas.Activity.gdWhiteGachaLowest[r.GetActivityID()], r)
	}

	logs.Debug("Load Hot Data WhiteLostGacha Success")
	return nil
}

type WhiteGachaShow struct {
}

func (nl *WhiteGachaShow) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.HOTGACHASHOW_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWhiteGachaShow = make(map[uint32][]*ProtobufGen.HOTGACHASHOW_ItemCondition, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWhiteGachaShow[r.GetActivityID()] = r.GetItem_Table()
	}

	logs.Debug("Load Hot Data WhiteShowGacha Success")
	return nil
}

func (d hotActivityData) GetActivityGachaSeting(activityId uint32) *ProtobufGen.HOTGACHASETTINGS {
	return d.gdWhiteGachaSetings[activityId]
}

func (d hotActivityData) GetActivityGachaNormal(gachaId uint32) NormalGachaConfig {
	return d.gdWhiteNormalGacha[gachaId]
}

func (d hotActivityData) IsWhiteGachaLowest(gachaId, gachaNum uint32) (string, uint32, bool) {
	for _, m := range d.gdWhiteGachaLowest[gachaId] {
		if m.GetLowestTimes() == gachaNum+1 {
			return m.GetItemID(), m.GetItemCount(), true
		}
	}
	return "", 0, false
}

func (d hotActivityData) IsWhiteGachaSpecial(activityId, gachaNum uint32) (uint32, bool) {
	for _, m := range d.gdWhiteGachaSpecil[activityId] {
		if m.GetSpecialTimes() == gachaNum+1 {
			return m.GetGachaID(), true
		}

	}
	return 0, false
}

func (d hotActivityData) GetActivityWhiteGachaShow(gachaId uint32) []*ProtobufGen.HOTGACHASHOW_ItemCondition {
	return d.gdWhiteGachaShow[gachaId]
}

func (d hotActivityData) GetWhiteGachaLowest(gachaId uint32) []*ProtobufGen.HOTGACHALOWEST {
	return d.gdWhiteGachaLowest[gachaId]
}
