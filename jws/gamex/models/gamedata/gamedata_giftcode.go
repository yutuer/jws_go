package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
)

type giftCodeData struct {
	ID        uint32
	StartTime int64
	EndTime   int64
	GiftName  string
	Rewards   []givesData
}

var (
	gdGiftCodeData []giftCodeData
)

//GetGiftCodeData 获取礼品码信息
func GetGiftCodeData(batchID int64) *giftCodeData {
	if batchID < 0 || int(batchID) >= len(gdGiftCodeData) {
		return nil
	}
	return &gdGiftCodeData[int(batchID)]
}

// 必须在loadGiftCodeGroupData之前调用
func loadGiftCodeBatchData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.CODEGIFTPATCH_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGiftCodeData = make([]giftCodeData, len(data)+1, len(data)+1)

	for _, c := range data {
		id := int(c.GetPatchID())
		if id < 0 || id >= len(gdGiftCodeData) {
			panic(fmt.Errorf("gdGiftCodeData id Err By %d", id))
		}
		gdGiftCodeData[id] = giftCodeData{
			ID:        c.GetPatchID(),
			StartTime: util.TimeFromString(c.GetStartTime()),
			EndTime:   util.TimeFromString(c.GetEndTime()),
			GiftName:  c.GetGiftName(),
			Rewards:   make([]givesData, 8, 8),
		}
	}
	//logs.Trace("gdGiftCodeData %v", gdGiftCodeData)
}

func loadGiftCodeGroupData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.CODEGIFTGROUP_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	for _, c := range data {
		data := &gdGiftCodeData[c.GetPatchID()]
		groupID := c.GetGroupID()
		reward := &data.Rewards[int(groupID)]
		for _, r := range c.GetGiftReward() {
			reward.AddItem(r.GetAWardID(), r.GetCount())
		}
		//logs.Trace("GiftCodeGroupData %v", gdGiftCodeData[c.GetPatchID()])
	}
	//logs.Trace("GiftCodeGroupData all %v", gdGiftCodeData)
}

//
//
//
//
//
//
