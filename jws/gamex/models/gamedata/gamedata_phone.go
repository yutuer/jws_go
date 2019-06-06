package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	ActSpecRewardIDXPhone   = iota
	ActSpecRewardEGReward   = 1
	ActSpecRewardMailReward = 2
)

var (
	gDActivitySpecRewards []PriceDatas
)

func loadActivitySpecRewards(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.SPCIALACTIVITY_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gDActivitySpecRewards = make([]PriceDatas, len(data), len(data))

	for idx, c := range data {
		for _, d := range c.GetGoalAward_Template() {
			gDActivitySpecRewards[idx].AddItem(d.GetReward(), d.GetCount())
		}
	}

	logs.Trace("loadActivitySpecRewards %v", gDActivitySpecRewards)
}

func GetActivitySpecRewards(idx int) *PriceDatas {
	if idx < 0 || idx >= len(gDActivitySpecRewards) {
		return nil
	}
	return &(gDActivitySpecRewards[idx])
}
