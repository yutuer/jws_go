package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	contentIDS  map[int32]string
	isAvailable map[int32]bool
	ChannelType map[int32]int
)

func loadRollInfo(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			logs.Error("err：%v", err)
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	//logs.Trace(string(buffer))
	rollinfo := &ProtobufGen.ROLLINFO_ARRAY{}
	err = proto.Unmarshal(buffer, rollinfo)
	errcheck(err)

	item_data := rollinfo.GetItems()

	contentIDS = make(map[int32]string)
	isAvailable = make(map[int32]bool)
	ChannelType = make(map[int32]int)

	for _, item := range item_data {
		contentIDS[item.GetEnum()] = item.GetContentIDS()
		isAvailable[item.GetEnum()] = item.GetIsAvailable() == 1
		ChannelType[item.GetEnum()] = int(item.GetChannelType())
	}
}

func GetContentIDS(enum int32) string {
	if _, ok := contentIDS[enum]; ok {
		return contentIDS[enum]
	}
	return ""
}

func GetIsAvailable(enum int32) bool {
	return isAvailable[enum]
}

func GetChannelType(enum int32) int {
	return ChannelType[enum]
}
