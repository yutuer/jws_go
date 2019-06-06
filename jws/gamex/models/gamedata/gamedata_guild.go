package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 公会职位
const (
	Guild_Pos_Mem = iota
	Guild_Pos_Chief
	Guild_Pos_ViceChief
	Guild_Pos_Elite
	Guild_Pos_Count
)

var (
	gdGuildLevel2GuildSignAward  []float32
	gdGuildLevelNextLevelExpNeed []uint32
	gdGuildPosNumMax             []int
	gdGuildPosData               []*ProtobufGen.GUILDPOSITION
)

func loadGuildMemNumberData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDNUMBERS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	gdGuildLevelNextLevelExpNeed = make([]uint32, len(data)+1, len(data)+1)
	gdGuildLevel2GuildSignAward = make([]float32, len(data)+1, len(data)+1)
	for _, d := range data {
		gdGuildLevelNextLevelExpNeed[int(d.GetGuildLevel())] = d.GetGuildEX()
		gdGuildLevel2GuildSignAward[int(d.GetGuildLevel())] = d.GetGuildSignAward()
	}
}

func loadGuildMemPosData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDPOSITION_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	gdGuildPosNumMax = make([]int, len(data)+1, len(data)+1)
	gdGuildPosData = make([]*ProtobufGen.GUILDPOSITION, len(data)+1, len(data)+1)
	for _, d := range data {
		gdGuildPosNumMax[int(d.GetPosition())] = int(d.GetPositionNumber())
		gdGuildPosData[int(d.GetPosition())] = d
		logs.Debug("rename guild config %d, %d", d.GetPosition(), d.GetReNamePower())
	}
	logs.Trace("gdGuildPosNumMax %v", gdGuildPosNumMax)
}

func GetGuildPosMaxNum(pos int) int {
	if pos <= 0 || pos >= len(gdGuildPosNumMax) {
		return 0
	}

	return gdGuildPosNumMax[pos]
}

func GetGuildGuildSignAward(lv int) float32 {
	if lv <= 0 || lv >= len(gdGuildLevel2GuildSignAward) {
		return 0.0
	}

	return gdGuildLevel2GuildSignAward[lv]
}

func GetGuildPosData(pos int) *ProtobufGen.GUILDPOSITION {
	if pos < 0 || pos >= len(gdGuildPosData) {
		return nil
	}

	return gdGuildPosData[pos]
}

func GetGuildXpNeedNext(lvl uint32) int64 {
	i := int(lvl)
	if i <= 0 || i >= len(gdGuildLevelNextLevelExpNeed) {
		return 0
	}

	return int64(gdGuildLevelNextLevelExpNeed[i])
}

// 是否有同意申请的权限
func CheckApprovePosition(position int) bool {
	posCfg := GetGuildPosData(position)
	return posCfg.GetAddMember() == 1
}

func GuildPositionString(position int) string {
	switch position {
	case Guild_Pos_Mem:
		return "Mem"
	case Guild_Pos_Chief:
		return "Chief"
	case Guild_Pos_ViceChief:
		return "ViceChief"
	case Guild_Pos_Elite:
		return "Elite"
	default:
		return fmt.Sprintf("%d", position)
	}
}
