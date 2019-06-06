package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
)

func loadActivityGiftByCondMain(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.CDTGIFTMAIN_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)

	datas := data_ar.GetItems()
	gdActivityGiftByConds = make(
		[]ActivityGiftByConds,
		0,
		ActTypInitLen)
	gdActivityGiftByCondsMap = make(map[uint32]*ActivityGiftByConds, ActTypInitLen)
	for _, c := range datas {
		newMainData := ActivityGiftByConds{}
		newMainData.ID = c.GetActivityID()
		newMainData.Gifts = make([]ActivityGiftByCondition, 0, ActGiftInitLen)
		newMainData.TimeBegin = util.TimeFromString(c.GetStartTime())
		newMainData.TimeEnd = util.TimeFromString(c.GetEndTime())
		newMainData.Title = c.GetActivityTitle()
		newMainData.Desc = c.GetActivityDesc()
		newMainData.Index = int(c.GetSortWeight())
		gdActivityGiftByConds = append(gdActivityGiftByConds, newMainData)
		gdActivityGiftByCondsMap[c.GetActivityID()] =
			&gdActivityGiftByConds[len(gdActivityGiftByConds)-1]
	}
}
