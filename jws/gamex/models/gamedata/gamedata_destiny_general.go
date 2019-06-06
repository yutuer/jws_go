package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type DestinyGeneralLevelData struct {
	ID          int
	LevelIndex  int
	Atk         float32
	Def         float32
	Hp          float32
	Exp         uint32
	Cfg         *ProtobufGen.NEWDESTINYGENERALLEVEL
	LevelUpCost PriceData // 老版本神兽需要
}

type DestinyGeneralPosition struct {
	UnlockDestinyLevelNeed int
	JadeType               string
	JadeIdx                int
}

type DestinyGeneralUnlockData struct {
	ID                 int
	UnlockLevel        uint32
	UnlockGeneralID    int
	UnlockGeneralLevel int
	UnlockCost         CostData
	Position           []DestinyGeneralPosition
	PositionLvNeed     []int
	IsCalcGs           bool
	Cfg                *ProtobufGen.DESTINYGENERALUNLOCK
}

var (
	gdDestinyGeneralLevelData    [][]DestinyGeneralLevelData
	gdNewDestinyGeneralLevelData [][]DestinyGeneralLevelData
	gdDestinyGeneralUnlockData   []DestinyGeneralUnlockData
	DestinyGeneralUnlockFirstLv  uint32
	gdDestConfig                 *ProtobufGen.DESTINYCONFIG
)

func GetNewDestinyGeneralLevelData(id, lv int) *DestinyGeneralLevelData {
	if id >= len(gdNewDestinyGeneralLevelData) || id < 0 {
		return nil
	}

	if lv >= len(gdNewDestinyGeneralLevelData[id]) || lv < 0 {
		return nil
	}

	return &gdNewDestinyGeneralLevelData[id][lv]
}

func GetNewDestinyGeneralLevelDatas(id int) []DestinyGeneralLevelData {
	if id >= len(gdNewDestinyGeneralLevelData) || id < 0 {
		return nil
	}

	return gdNewDestinyGeneralLevelData[id][:]
}

func GetDestinyGeneralLevelData(id, lv int) *DestinyGeneralLevelData {
	if id >= len(gdDestinyGeneralLevelData) || id < 0 {
		return nil
	}

	if lv >= len(gdDestinyGeneralLevelData[id]) || lv < 0 {
		return nil
	}

	return &gdDestinyGeneralLevelData[id][lv]
}

func GetDestinyConfig() *ProtobufGen.DESTINYCONFIG {
	return gdDestConfig
}

func GetDestinyGeneralLevelDatas(id int) []DestinyGeneralLevelData {
	if id >= len(gdDestinyGeneralLevelData) || id < 0 {
		return nil
	}

	return gdDestinyGeneralLevelData[id][:]
}

func GetDestinyGeneralUnlockData(id int) *DestinyGeneralUnlockData {
	if id >= len(gdDestinyGeneralUnlockData) || id < 0 {
		return nil
	}

	return &gdDestinyGeneralUnlockData[id]
}

func GetDestingGeneralIdCount() int {
	return len(gdDestinyGeneralUnlockData)
}

func loadNewDestinyGeneralLevelData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.NEWDESTINYGENERALLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()

	gdNewDestinyGeneralLevelData = make([][]DestinyGeneralLevelData, 0, len(ar))
	for _, a := range ar {
		nd := DestinyGeneralLevelData{}
		nd.ID = int(a.GetDestinyGeneralID())
		nd.LevelIndex = int(a.GetDestinyGeneralLevelID())
		nd.Atk = float32(a.GetAttackIncrease())
		nd.Def = float32(a.GetDefenseIncrease())
		nd.Hp = float32(a.GetHPIncrease())
		nd.Exp = a.GetDestinyGeneralExp()
		nd.Cfg = a
		for nd.ID >= len(gdNewDestinyGeneralLevelData) {
			gdNewDestinyGeneralLevelData = append(gdNewDestinyGeneralLevelData,
				make([]DestinyGeneralLevelData, 0, 128))
		}

		for nd.LevelIndex >= len(gdNewDestinyGeneralLevelData[nd.ID]) {
			gdNewDestinyGeneralLevelData[nd.ID] = append(gdNewDestinyGeneralLevelData[nd.ID],
				DestinyGeneralLevelData{})
		}

		gdNewDestinyGeneralLevelData[nd.ID][nd.LevelIndex] = nd
	}
}

func loadDestinyGeneralLevelData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.DESTINYGENERALLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()

	gdDestinyGeneralLevelData = make([][]DestinyGeneralLevelData, 0, len(ar))
	for _, a := range ar {
		nd := DestinyGeneralLevelData{}
		nd.ID = int(a.GetDestinyGeneralID())
		nd.LevelIndex = int(a.GetDestinyGeneralLevelID())
		nd.Atk = float32(a.GetAttackIncrease())
		nd.Def = float32(a.GetDefenseIncrease())
		nd.Hp = float32(a.GetHPIncrease())
		for _, c := range a.GetMaterial_Table() {
			nd.LevelUpCost.AddItem(c.GetCostMaterial(), c.GetCostMaterialNumber())
		}
		for nd.ID >= len(gdDestinyGeneralLevelData) {
			gdDestinyGeneralLevelData = append(gdDestinyGeneralLevelData,
				make([]DestinyGeneralLevelData, 0, 128))
		}

		for nd.LevelIndex >= len(gdDestinyGeneralLevelData[nd.ID]) {
			gdDestinyGeneralLevelData[nd.ID] = append(gdDestinyGeneralLevelData[nd.ID],
				DestinyGeneralLevelData{})
		}

		gdDestinyGeneralLevelData[nd.ID][nd.LevelIndex] = nd
	}
}

func loadDestinyGeneralUnlockData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.DESTINYGENERALUNLOCK_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()

	gdDestinyGeneralUnlockData = make([]DestinyGeneralUnlockData, 0, len(ar))
	for _, a := range ar {
		nd := DestinyGeneralUnlockData{}
		nd.ID = int(a.GetDestinyGeneralID())
		nd.UnlockGeneralID = int(a.GetUnlockGeneralID())
		nd.UnlockGeneralLevel = int(a.GetUnlockGeneralLevel())
		nd.UnlockLevel = (a.GetUnlockLevel())
		if a.GetUnlockGeneralMaterial() != "" {
			nd.UnlockCost.AddItem(a.GetUnlockGeneralMaterial(), 1)
		} else {
			nd.IsCalcGs = true
		}
		nd.Position = make([]DestinyGeneralPosition, 0, len(a.GetPosition_Table()))
		nd.PositionLvNeed = make([]int, 0, len(a.GetPosition_Table())+1)
		nd.Cfg = a
		for _, c := range a.GetPosition_Table() {
			JadeIdx := GetJadeSlot(c.GetJadeType())
			nd.Position = append(nd.Position, DestinyGeneralPosition{
				UnlockDestinyLevelNeed: int(c.GetPositonLevel()),
				JadeType:               c.GetJadeType(),
				JadeIdx:                JadeIdx,
			})
			for JadeIdx >= len(nd.PositionLvNeed) {
				nd.PositionLvNeed = append(nd.PositionLvNeed, 0)
			}
			nd.PositionLvNeed[JadeIdx] = int(c.GetPositonLevel())
		}

		for nd.ID >= len(gdDestinyGeneralUnlockData) {
			gdDestinyGeneralUnlockData = append(gdDestinyGeneralUnlockData,
				DestinyGeneralUnlockData{})
		}

		gdDestinyGeneralUnlockData[nd.ID] = nd

	}

	DestinyGeneralUnlockFirstLv = gdDestinyGeneralUnlockData[0].UnlockLevel

	logs.Trace("loadDestinyGeneralUnlockData %v", gdDestinyGeneralUnlockData)
}

func loadDestinyGeneralConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.DESTINYCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()
	gdDestConfig = ar[0]
}
