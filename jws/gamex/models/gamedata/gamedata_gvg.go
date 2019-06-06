package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const GVGTopN = 2

var (
	gdGVGCity      []int
	gdGVGConfig    *ProtobufGen.GVGCONFIG
	gdGVGWinPoint  []int
	gdGVGACityGift [GVGTopN]map[uint32]*ProtobufGen.GVGACITYGIFT
	gdGVGGuildGift [GVGTopN]map[uint32]guildGBDropsData
	gdGVGDailyGift map[uint32]*ProtobufGen.GVGDAILYGIFT
	gdGVGPointGift []*ProtobufGen.GVGPOINTGIFT
	gdGVGCityPrio  []int
)

func _common_load(filepath string, v proto.Message) {
	_errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	_errcheck(err)
	err = proto.Unmarshal(buffer, v)
	_errcheck(err)

}
func loadGVGCityIDData(filepath string) {
	ar := &ProtobufGen.GVGCITY_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	gdGVGCity = make([]int, 0, len(data))
	gdGVGCityPrio = make([]int, 0, len(data))
	for _, v := range data {
		if v.GetCityOpening() == 1 {
			gdGVGCity = append(gdGVGCity, int(v.GetCityID()))
			gdGVGCityPrio = append(gdGVGCityPrio, int(v.GetCityType()))
		}
	}
}

func loadGVGConfigData(filepath string) {
	ar := &ProtobufGen.GVGCONFIG_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	gdGVGConfig = data[0]
}

func loadGVGWinsScore(filepath string) {
	ar := &ProtobufGen.GVGWINSPOINT_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	l := len(data) + 1
	gdGVGWinPoint = make([]int, l, l)
	for _, item := range data {
		gdGVGWinPoint[item.GetWinNum()] = int(item.GetGVGPoint())
	}
}

// 攻城礼包
func loadGVGActivityGift(filepath string) {
	ar := &ProtobufGen.GVGACITYGIFT_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	l := len(data)
	for i := 0; i < len(gdGVGACityGift); i++ {
		gdGVGACityGift[i] = make(map[uint32]*ProtobufGen.GVGACITYGIFT, l)
		for _, item := range data {
			gdGVGACityGift[i][item.GetCityID()] = item
		}

	}

}

// 税收礼包
func loadGVGGuildGift(filepath string) {
	ar := &ProtobufGen.GVGGUILDGIFT_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	l := len(data)
	for i := 0; i < len(gdGVGGuildGift); i++ {
		gdGVGGuildGift[i] = make(map[uint32]guildGBDropsData, l)
		for _, item := range data {
			dropData := guildGBDropsData{
				Ids:    make([]string, 0, 1),
				Counts: make([]uint32, 0, 1),
			}
			dropData.Ids = append(dropData.Ids, item.GetGuildBagID())
			if i == 0 {
				dropData.Counts = append(dropData.Counts, item.GetGuildBagNum())
			} else if i == 1 {
				dropData.Counts = append(dropData.Counts, item.GetSecondGuildBagNum())
			} else {
				logs.Error("load GVG guild gift err by no data id is: %d", i)
				// 采用默认第一个
				dropData.Counts = append(dropData.Counts, item.GetGuildBagNum())
			}
			gdGVGGuildGift[i][item.GetCityID()] = dropData
		}
	}

}

// 每日礼包
func loadGVGDailyGift(filepath string) {
	ar := &ProtobufGen.GVGDAILYGIFT_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	l := len(data)
	gdGVGDailyGift = make(map[uint32]*ProtobufGen.GVGDAILYGIFT, l)
	for _, item := range data {
		gdGVGDailyGift[item.GetCityID()] = item
	}
}

// 参与礼包
func loadGVGPointGift(filepath string) {
	ar := &ProtobufGen.GVGPOINTGIFT_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	l := len(data)
	gdGVGPointGift = make([]*ProtobufGen.GVGPOINTGIFT, 0, l)
	for _, item := range data {
		gdGVGPointGift = append(gdGVGPointGift, item)
	}
}

func GetGVGCityID() []int {
	return gdGVGCity
}

func GetGVGCityPrio() []int {
	return gdGVGCityPrio
}

func GetGVGConfig() *ProtobufGen.GVGCONFIG {
	return gdGVGConfig
}

func GetGVGWinScore(times int) int {
	if times > len(gdGVGWinPoint)-1 {
		times = len(gdGVGWinPoint) - 1
	}
	return gdGVGWinPoint[times]
}

func GetGVGActivityGiftCfg(city uint32, rank int) *ProtobufGen.GVGACITYGIFT {
	if rank > len(gdGVGACityGift) {
		return nil
	}
	return gdGVGACityGift[rank][city]
}

func GetGVGPointGiftCfg(point uint32) *ProtobufGen.GVGPOINTGIFT {
	for _, _point := range gdGVGPointGift {
		if point >= _point.GetGVGPoint1() && point <= _point.GetGVGPoint2() {
			return _point
		}
	}
	if point > gdGVGPointGift[len(gdGVGPointGift)-1].GetGVGPoint2() {
		return gdGVGPointGift[len(gdGVGPointGift)-1]
	} else {
		return nil
	}
}

func GetGVGActivityDailyGift(city uint32) *ProtobufGen.GVGDAILYGIFT {
	return gdGVGDailyGift[city]
}

func GetGVGActivityGuildGift(city uint32, rank int64) guildGBDropsData {
	if rank > int64(len(gdGVGGuildGift)) {
		return guildGBDropsData{}
	}
	return gdGVGGuildGift[rank][city]
}
