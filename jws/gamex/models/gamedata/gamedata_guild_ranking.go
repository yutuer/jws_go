package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdGuildRankingData [][Guild_Pos_Count]string
)

func loadGuildRankingData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDRANKINGAWARD_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	gdGuildRankingData = make([][Guild_Pos_Count]string, len(data), len(data))
	for _, d := range data {
		rank := int(d.GetRanking())
		pos := int(d.GetPosition())
		if rank >= len(gdGuildRankingData) {
			panic(fmt.Errorf("gdGuildRankingData Rank Err By %d", rank))
		}
		if pos >= Guild_Pos_Count {
			panic(fmt.Errorf("gdGuildRankingData Pos Err By %d %d", rank, pos))
		}
		gdGuildRankingData[rank][pos] = d.GetRwardRank()
	}
}

func GetGuildRankingReward(rank, pos int) string {
	if rank >= len(gdGuildRankingData) {
		return ""
	}

	if pos >= Guild_Pos_Count {
		return ""
	}
	return gdGuildRankingData[rank][pos]
}
