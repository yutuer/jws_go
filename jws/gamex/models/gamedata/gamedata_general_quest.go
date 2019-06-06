package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	GeneralQuestListMax = 3
)

var (
	gdGeneralQuest        map[string]*ProtobufGen.NGQDETAIL
	gdGeneralQuestCondMap map[Condition][]string
	gdGeneralQuestRefTime []string
	gdGeneralQuestSetting *ProtobufGen.NGQSETTINGS
)

func loadGeneralQuestConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.NGQDETAIL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGeneralQuest = make(map[string]*ProtobufGen.NGQDETAIL, len(dataList.Items))
	gdGeneralQuestCondMap = make(map[Condition][]string, 16)
	for _, q := range dataList.Items {
		gdGeneralQuest[q.GetNGQID()] = q
		cond := Condition{
			Ctyp:   q.GetFCType(),
			Param1: int64(q.GetFCValueIP1()),
			Param2: int64(q.GetFCValueIP2()),
		}
		qs, ok := gdGeneralQuestCondMap[cond]
		if !ok {
			qs = make([]string, 0, 56)
		}
		qs = append(qs, q.GetNGQID())
		gdGeneralQuestCondMap[cond] = qs

		if q.GetBonusType() != 0 && q.GetBonusType() != 1 {
			panic(
				fmt.Errorf("loadGeneralQuestConfig quest %s  q.GetBonusType() != 0 && q.GetBonusType() != 1",
					q.GetNGQID()))
		}
	}
}

func loadGeneralQuestRefTimeConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.NGQREFRESHTIME_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGeneralQuestRefTime = make([]string, 0, len(dataList.Items))
	for _, t := range dataList.Items {
		gdGeneralQuestRefTime = append(gdGeneralQuestRefTime, t.GetRefreshTime())
	}
}

func loadGeneralQuestSettingConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.NGQSETTINGS_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGeneralQuestSetting = dataList.Items[0]
}

func GeneralQuestCfg(questId string) *ProtobufGen.NGQDETAIL {
	return gdGeneralQuest[questId]
}

func GeneralQuestCondCfg() map[Condition][]string {
	return gdGeneralQuestCondMap
}

func GeneralQuestRefreshTime() []string {
	return gdGeneralQuestRefTime
}

func GeneralQuestSetting() *ProtobufGen.NGQSETTINGS {
	return gdGeneralQuestSetting
}
