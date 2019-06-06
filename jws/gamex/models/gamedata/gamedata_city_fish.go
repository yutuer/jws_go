package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdFishCost   *ProtobufGen.FISHINGCOST
	gdFishReward []*ProtobufGen.FISHINGREWARD
)

func loadFishCostConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.FISHINGCOST_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdFishCost = dataList.Items[0]
}

func loadFishReward(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.FISHINGREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdFishReward = make([]*ProtobufGen.FISHINGREWARD, 0, len(dataList.Items))
	for _, r := range dataList.Items {
		gdFishReward = append(gdFishReward, r)
	}
}

func FishCost() *ProtobufGen.FISHINGCOST {
	return gdFishCost
}

func FishRewardCount() ([]uint32, uint32) {
	var sum uint32
	res := make([]uint32, 0, len(gdFishReward))
	for _, r := range gdFishReward {
		res = append(res, r.GetLimit())
		sum += r.GetLimit()
	}
	return res, sum
}

func GetFishReward(idx int) *ProtobufGen.FISHINGREWARD {
	if idx >= len(gdFishReward) {
		return nil
	}
	return gdFishReward[idx]
}
