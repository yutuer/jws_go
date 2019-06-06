package gamedata

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type universalMaterialData struct {
	ID      int
	Lv      int
	ItemID  string
	AllNeed int64
}

var (
	gdUniversalMaterialDatas      []universalMaterialData
	gdUniversalMaterialItem2Datas map[string]*universalMaterialData
	gdUniversalMaterialByLv       [][]*universalMaterialData
	gdMaxLvUniverMaterialDatas    []*universalMaterialData
	gdMaxLvUniverMaterialIdxs     []int
)

func loadUniversalMaterialDatas(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.UNIVERSALMATERIAL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	items := ar.GetItems()
	gdUniversalMaterialDatas = make([]universalMaterialData, len(items), len(items))
	gdUniversalMaterialItem2Datas = make(map[string]*universalMaterialData, len(items))
	gdUniversalMaterialByLv = make([][]*universalMaterialData, len(items), len(items))

	for _, item := range items {
		id := int(item.GetUniversalMaterialID())
		lv := int(item.GetUniversalMaterialLevel())
		u := universalMaterialData{
			ID:     id,
			Lv:     lv,
			ItemID: item.GetMaterialID(),
		}
		gdUniversalMaterialDatas[id] = u
		gdUniversalMaterialItem2Datas[item.GetMaterialID()] =
			&gdUniversalMaterialDatas[id]
		gdUniversalMaterialByLv[lv] = append(gdUniversalMaterialByLv[lv],
			&gdUniversalMaterialDatas[id])
	}
	for _, lvs := range gdUniversalMaterialByLv {
		if lvs == nil || len(lvs) <= 0 {
			continue
		}
		if gdMaxLvUniverMaterialDatas == nil ||
			len(gdMaxLvUniverMaterialDatas) <= 0 ||
			gdMaxLvUniverMaterialDatas[0].Lv < lvs[0].Lv {
			gdMaxLvUniverMaterialDatas = lvs[:]
		}
	}
	gdMaxLvUniverMaterialIdxs = make([]int,
		0,
		len(gdMaxLvUniverMaterialDatas))
	for _, dataInMaxLv := range gdMaxLvUniverMaterialDatas {
		gdMaxLvUniverMaterialIdxs = append(
			gdMaxLvUniverMaterialIdxs,
			dataInMaxLv.ID)
	}

	logs.Trace("gdUniversalMaterialDatas, %v", gdUniversalMaterialDatas)
	logs.Trace("gdUniversalMaterialItem2Datas, %v", gdUniversalMaterialItem2Datas)
	logs.Trace("gdUniversalMaterialByLv, %v", gdUniversalMaterialByLv)
	logs.Trace("gdUniversalMaterialMax, %v %d", gdMaxLvUniverMaterialDatas,
		gdMaxLvUniverMaterialDatas[0].Lv)

}

func GetUniversalMaterialData(ID int) *universalMaterialData {
	if ID < 0 || ID >= len(gdUniversalMaterialDatas) {
		return nil
	}
	return &gdUniversalMaterialDatas[ID]
}

func GetUniversalMaterialDatas() []universalMaterialData {
	return gdUniversalMaterialDatas[:]
}

func GetUniversalMaterialDataByLv(lv int) []*universalMaterialData {
	if lv < 0 || lv >= len(gdUniversalMaterialByLv) {
		return []*universalMaterialData{}
	}
	return gdUniversalMaterialByLv[lv][:]
}

func GetUniversalMaterialMaxLvData() ([]int, int) {
	return gdMaxLvUniverMaterialIdxs[:],
		gdMaxLvUniverMaterialDatas[0].Lv
}

func GetUniversalMaterialDataByItem(itemID string) *universalMaterialData {
	r, ok := gdUniversalMaterialItem2Datas[itemID]
	if ok {
		return r
	} else {
		return nil
	}
}

func addNeed(itemID string, needCount uint32) {
	if gdUniversalMaterialDatas == nil || len(gdUniversalMaterialDatas) == 0 {
		panic(errors.New("addNeed should before loadUniversalMaterialDatas"))
	}
	data := GetUniversalMaterialDataByItem(itemID)
	if data != nil {
		data.AllNeed += int64(needCount)
	}
}
