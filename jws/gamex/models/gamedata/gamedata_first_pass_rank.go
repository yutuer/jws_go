package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	_ = iota
	FirstPassRankTypSimplePvp
	FirstPassRankTypTeamPvp
	FirstPassRankTypCount
)

type FirstPassRewardData struct {
	Index  int
	Start  int
	End    int
	Reward PriceDatas
}

var (
	gdFirstPassRewardData [FirstPassRankTypCount][]FirstPassRewardData
)

func GetFirstPassRewardData(t int) []FirstPassRewardData {
	if t < 0 || t >= len(gdFirstPassRewardData) {
		return nil
	}

	return gdFirstPassRewardData[t][:]
}

func GetFirstPassRankIs1MineTheMax(t int) bool {
	switch t {
	case FirstPassRankTypSimplePvp:
		return false
	case FirstPassRankTypTeamPvp:
		return true
	default:
		logs.Error("Unknown FirstPassRankTyp %v", t)
		return true
	}
}

func loadSimplePvpFirstPassReward(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.BSCPVPSECTOR_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	gdFirstPassRewardData[FirstPassRankTypSimplePvp] =
		make([]FirstPassRewardData,
			len(ar.GetItems())+1,
			len(ar.GetItems())+1)
	for _, data := range ar.GetItems() {
		r := FirstPassRewardData{
			Index: int(data.GetIndex()),
			Start: int(data.GetStart()),
			End:   int(data.GetEnd()),
		}
		for _, reward := range data.GetFixed_Loot() {
			r.Reward.AddItem(
				reward.GetFixedLootID(),
				reward.GetFixedLootNumber())
		}
		gdFirstPassRewardData[FirstPassRankTypSimplePvp][r.Index] = r
	}
	logs.Trace("gdFirstPassRewardData FirstPassRankTypSimplePvp, %v",
		gdFirstPassRewardData[FirstPassRankTypSimplePvp])
}

func loadTeamPvpFirstPassReward(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.TPVPFPASS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	gdFirstPassRewardData[FirstPassRankTypTeamPvp] =
		make([]FirstPassRewardData,
			len(ar.GetItems())+1,
			len(ar.GetItems())+1)
	for _, data := range ar.GetItems() {
		r := FirstPassRewardData{
			Index: int(data.GetIndex()),
			Start: int(data.GetStart()),
			End:   int(data.GetEnd()),
		}
		for _, reward := range data.GetFixed_Loot() {
			r.Reward.AddItem(
				reward.GetFixedLootID(),
				reward.GetFixedLootNumber())
		}
		gdFirstPassRewardData[FirstPassRankTypTeamPvp][r.Index] = r
	}
	logs.Trace("gdFirstPassRewardData FirstPassRankTypTeamPvp, %v",
		gdFirstPassRewardData[FirstPassRankTypTeamPvp])
}
