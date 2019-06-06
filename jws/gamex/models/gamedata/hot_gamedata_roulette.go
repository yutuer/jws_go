package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type LuckyWheelSetings struct {
}

func (gs *LuckyWheelSetings) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.WHEELSETTINGS_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWheelSetings = make(map[uint32]*ProtobufGen.WHEELSETTINGS, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWheelSetings[r.GetActivityID()] = r
	}

	logs.Debug("Load Hot Data LuckyRouletteSetings Success")
	return nil
}

type LuckyWheelGacha struct {
}

func (nl *LuckyWheelGacha) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.WHEELGACHA_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWheelGacha = make(map[uint32]WheelGachaConfig, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWheelGacha[r.GetGachaID()] = append(datas.Activity.gdWheelGacha[r.GetGachaID()], r)
	}
	logs.Debug("Load Hot Data LuckyRouletteGacha Success")
	return nil
}

type LuckyWheelCost struct {
}

func (nl *LuckyWheelCost) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.WHEELCOST_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWheelCost = make(map[uint32][]*ProtobufGen.WHEELCOST, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWheelCost[r.GetActivityID()] = append(datas.Activity.gdWheelCost[r.GetActivityID()], r)
	}

	logs.Debug("Load Hot Data LuckyRouletteCost Success")
	return nil
}

type LuckyWheelShow struct {
}

func (nl *LuckyWheelShow) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.WHEELSHOW_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdWheelShow = make(map[uint32][]*ProtobufGen.WHEELSHOW_ItemCondition, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdWheelShow[r.GetActivityID()] = r.GetItem_Table()
	}
	logs.Debug("Load Hot Data WhiteShowGacha Success")
	return nil
}
func (d hotActivityData) GetWheelSeting(activityId uint32) *ProtobufGen.WHEELSETTINGS {
	return d.gdWheelSetings[activityId]
}

func (d hotActivityData) GetWheelGachaNormal(gachaId uint32) WheelGachaConfig {
	return d.gdWheelGacha[gachaId]
}

func (d hotActivityData) GetWheelCost(activityId uint32) []*ProtobufGen.WHEELCOST {
	return d.gdWheelCost[activityId]
}

//暗控组直接使用白盒宝箱

func (d hotActivityData) GetWheelGachaShow(activityId uint32) []*ProtobufGen.WHEELSHOW_ItemCondition {
	return d.gdWheelShow[activityId]
}
