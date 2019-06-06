package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

type hotMoneyCatActivity struct {
}

func (act *hotMoneyCatActivity) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.MONEYGOD_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}
	item_data := dataList.GetItems()
	datas.Activity.gdMoneyCatData = item_data
	for _, value := range item_data {
		var maxWeight int32 = 0
		for i := 0; i < len(value.GetFixed_Num()); i++ {
			maxWeight += value.GetFixed_Num()[i].GetWeight()
			if i == len(value.GetFixed_Num())-1 {
				datas.Activity.gdMoneyWeight = append(datas.Activity.gdMoneyWeight, maxWeight)
				maxWeight = 0
			}
		}
		datas.Activity.gdMoneyCatSubData = append(datas.Activity.gdMoneyCatSubData, value.GetFixed_Num())
	}
	return nil
}

func (d hotActivityData) GetMoneyCatNum(step int64, section int) (minnum, maxnum int64) {
	data := d.GetMoneyCatSubData(step)
	minnum = int64(data[section].GetSMinNum())
	maxnum = int64(data[section].GetSMaxNum())
	return

}

func (d hotActivityData) GetMoneyCatCost(step int64) uint32 {
	return uint32(d.gdMoneyCatData[step].GetCostHC())

}

func (d hotActivityData) GetMoneyCatMarquee(step int64) uint32 {
	return d.gdMoneyCatData[step].GetOpenAdv()

}

func (d hotActivityData) GetMoneyCatWeight(step int64) int32 {
	return d.gdMoneyWeight[step]
}

func (d hotActivityData) GetMoneyCatSubData(step int64) []*ProtobufGen.MONEYGOD_Num1 {
	return d.gdMoneyCatSubData[step]
}
