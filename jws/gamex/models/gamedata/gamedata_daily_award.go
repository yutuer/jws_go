package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdDailyAwardCount int
	gdDailyAwardInfos map[uint32]map[uint32]*ProtobufGen.DAILYAWARD // id -> subId
)

func loadDailyAward(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.DAILYAWARD_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()
	gdDailyAwardInfos = make(map[uint32]map[uint32]*ProtobufGen.DAILYAWARD, 5)
	for _, da := range ar {
		info, ok := gdDailyAwardInfos[da.GetDailyAwardID()]
		if !ok {
			info = make(map[uint32]*ProtobufGen.DAILYAWARD, 5)
		}
		info[da.GetDailyAwardSubID()] = da
		gdDailyAwardInfos[da.GetDailyAwardID()] = info
	}
	gdDailyAwardCount = len(gdDailyAwardInfos)
}

func GetDailyAwardCount() int {
	return gdDailyAwardCount
}

func GetDailyAwardbyId(id uint32) map[uint32]*ProtobufGen.DAILYAWARD {
	return gdDailyAwardInfos[id]
}
