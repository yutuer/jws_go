package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type corpLevelInfo struct {
	CorpLv         uint32
	CorpXpNeed     uint32
	MaxEnergy      uint32
	BossFightPoint uint32
}

var (
	gdCorpLevelInfo []corpLevelInfo
)

func loadCorpLevelInfo(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	corp_lv_ar := &ProtobufGen.CORPLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, corp_lv_ar)
	errcheck(err)

	corp_lv_data := corp_lv_ar.GetItems()
	gdCorpLevelInfo = make(
		[]corpLevelInfo,
		len(corp_lv_data),
		len(corp_lv_data))

	for _, c := range corp_lv_data {
		lv := c.GetCorpLevel()
		if lv >= uint32(len(corp_lv_data)) {
			logs.Error("corp_lv_data error by lv %d", lv)
		}

		gdCorpLevelInfo[lv].CorpLv = lv
		gdCorpLevelInfo[lv].CorpXpNeed = c.GetCorpXP()
		gdCorpLevelInfo[lv].MaxEnergy = c.GetManualValue()
		gdCorpLevelInfo[lv].BossFightPoint = c.GetSpirit()
		//logs.Trace("corp_lv %d --> %v", lv, gdCorpLevelInfo[lv])
	}
}

func GetCorpLvConfig(lv uint32) *corpLevelInfo {
	if lv >= uint32(len(gdCorpLevelInfo)) {
		return nil
	}
	return &gdCorpLevelInfo[lv]
}
