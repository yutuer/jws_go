package config

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

type Gift struct {
	Type     uint32
	CondVal  uint32
	GiftData []*ProtobufGen.YYBGIFT_ItemCondition
	PKey     string
}

var (
	giftDatas map[uint32]Gift
)

const (
	Sign_Typ  = 0
	Login_Typ = 1
	Level_Typ = 2
	GS_Type   = 3
)

func LoadGiftData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.YYBGIFT_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)
	giftDatas = make(map[uint32]Gift, 20)
	items := dataList.GetItems()
	for _, item := range items {
		giftDatas[item.GetGiftID()] = Gift{
			Type:     item.GetGiftType(),
			CondVal:  item.GetFCValue1(),
			GiftData: item.GetItem_Table(),
			PKey:     item.GetPkey(),
		}
	}
}

func GetGiftByID(taskID uint32) (Gift, bool) {
	gift, exist := giftDatas[taskID]
	return gift, exist
}
