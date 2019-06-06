package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	TitleTyp_Common = iota
	TitleTyp_SimplePvp
	TitleTyp_TeamPvp
	TitleTyp_7DayGsRank
	TitleTyp_GVG
	TitleTyp_WuShuang
)

var (
	gdTitleSimplePvpRank map[int]*ProtobufGen.TITLELIST
	gdTitleTeamPvpRank   map[int]*ProtobufGen.TITLELIST
	gdTitle7DayGsRank    map[int]*ProtobufGen.TITLELIST
	gdTitles             map[string]*ProtobufGen.TITLELIST
	gdTitleCond          map[uint32][]*ProtobufGen.TITLELIST
	gdTitleGVG           *ProtobufGen.TITLELIST
	gdTitleWuShuang      map[int]*ProtobufGen.TITLELIST
)

func loadTitle(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.TITLELIST_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdTitleSimplePvpRank = make(map[int]*ProtobufGen.TITLELIST, 10)
	gdTitleTeamPvpRank = make(map[int]*ProtobufGen.TITLELIST, 10)
	gdTitle7DayGsRank = make(map[int]*ProtobufGen.TITLELIST, 10)
	gdTitles = make(map[string]*ProtobufGen.TITLELIST, len(ar.GetItems()))
	gdTitleCond = make(map[uint32][]*ProtobufGen.TITLELIST, len(ar.GetItems()))
	gdTitleWuShuang = make(map[int]*ProtobufGen.TITLELIST, len(ar.GetItems()))
	for _, item := range ar.GetItems() {
		gdTitles[item.GetTitleID()] = item
		if item.GetTitleConditionType() == TitleTyp_SimplePvp {
			for i := item.GetFCValueIP1(); i <= item.GetFCValueIP2(); i++ {
				gdTitleSimplePvpRank[int(i)] = item
			}
		} else if item.GetTitleConditionType() == TitleTyp_TeamPvp {
			for i := item.GetFCValueIP1(); i <= item.GetFCValueIP2(); i++ {
				gdTitleTeamPvpRank[int(i)] = item
			}
		} else if item.GetTitleConditionType() == TitleTyp_7DayGsRank {
			for i := item.GetFCValueIP1(); i <= item.GetFCValueIP2(); i++ {
				gdTitle7DayGsRank[int(i)] = item
			}
		} else if item.GetTitleConditionType() == TitleTyp_GVG {
			gdTitleGVG = item
		} else if item.GetTitleConditionType() == TitleTyp_WuShuang {
			for i := item.GetFCValueIP1(); i <= item.GetFCValueIP2(); i++ {
				gdTitleWuShuang[int(i)] = item
			}
		} else {
			ts, ok := gdTitleCond[item.GetFCType()]
			if !ok {
				ts = make([]*ProtobufGen.TITLELIST, 0, 10)
				gdTitleCond[item.GetFCType()] = ts
			}
			ts = append(ts, item)
			gdTitleCond[item.GetFCType()] = ts
		}
	}
}

func TitleSimpePvpSum() int {
	return len(gdTitleSimplePvpRank)
}

func TitleSimplePvpRank(rank int) *ProtobufGen.TITLELIST {
	return gdTitleSimplePvpRank[rank]
}

func TitleTeamPvpRankSum() int {
	return len(gdTitleTeamPvpRank)
}

func TitleTeamPvpRank(rank int) *ProtobufGen.TITLELIST {
	return gdTitleTeamPvpRank[rank]
}

func Title7DayGsRankSum() int {
	return len(gdTitle7DayGsRank)
}

func Title7DayGsRank(rank int) *ProtobufGen.TITLELIST {
	return gdTitle7DayGsRank[rank]
}

func GetTitleCfg(titleId string) *ProtobufGen.TITLELIST {
	return gdTitles[titleId]
}

func GetTitleCond(cond int) []*ProtobufGen.TITLELIST {
	return gdTitleCond[uint32(cond)]
}

func TitleGVG() *ProtobufGen.TITLELIST {
	return gdTitleGVG
}

func TitleWushuangRankSum() int {
	return len(gdTitleWuShuang)
}

func TitleWushuangRank(rank int) *ProtobufGen.TITLELIST {
	return gdTitleWuShuang[rank]
}
