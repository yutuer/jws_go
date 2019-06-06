package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	WantGeneralDiceCount = 6
)

var (
	gdWantGeneralCommon     *ProtobufGen.ACTIVITYCONFIG
	gdWantGeneralGift       map[uint32]*ProtobufGen.ACTIVITYGIFT
	gdWantGeneralResetCost  map[uint32]*ProtobufGen.RESETCOST
	gdWantGeneralProbablity map[uint32]*ProtobufGen.ACTIVITYPROBABILITY
	gdHeroPiece             string
)

func mkWantGeneralDatas(load loadDataFunc) {
	load("activityconfig.data", loadWantGeneralConfig)
	load("activitygift.data", loadWantGeneralGiftConfig)
	load("resetcost.data", loadWantGeneralResetCostConfig)
	load("activityprobability.data", loadWantProbabilityConfig)
}

func GetWantGeneralCommonConfig() *ProtobufGen.ACTIVITYCONFIG {
	return gdWantGeneralCommon
}

func GetWantGeneralAwardConfig(n uint32) *ProtobufGen.ACTIVITYGIFT {
	return gdWantGeneralGift[n]
}

func GetWantGeneralAwardResetCostConfig(n uint32) *ProtobufGen.RESETCOST {
	return gdWantGeneralResetCost[n]
}

func GetWantGeneralProbablityConfig(n uint32) *ProtobufGen.ACTIVITYPROBABILITY {
	return gdWantGeneralProbablity[n]
}

func GetWantGeneralHeroPiece() string {
	return gdHeroPiece
}

func loadWantGeneralConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.ACTIVITYCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()
	gdWantGeneralCommon = ar[0]
}

func loadWantGeneralGiftConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.ACTIVITYGIFT_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()
	gdWantGeneralGift = make(map[uint32]*ProtobufGen.ACTIVITYGIFT, len(ar))
	for _, a := range ar {
		gdWantGeneralGift[a.GetHeroAmount()] = a
		for _, v := range a.GetFixed_Loot() {
			if IsHeroPiece(v.GetGiftID()) {
				if gdHeroPiece == "" {
					gdHeroPiece = v.GetGiftID()
				}
				if gdHeroPiece != "" && gdHeroPiece != v.GetGiftID() {
					panic(fmt.Errorf("WantGeneral hero piece diff %s %s", gdHeroPiece, v.GetGiftID()))
				}
			}
		}
	}
}

func loadWantGeneralResetCostConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.RESETCOST_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()
	gdWantGeneralResetCost = make(map[uint32]*ProtobufGen.RESETCOST, len(ar))
	for _, a := range ar {
		gdWantGeneralResetCost[a.GetResetTimes()] = a
	}
}

func loadWantProbabilityConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.ACTIVITYPROBABILITY_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	ar := lv_ar.GetItems()
	gdWantGeneralProbablity = make(map[uint32]*ProtobufGen.ACTIVITYPROBABILITY, len(ar))
	for _, a := range ar {
		gdWantGeneralProbablity[a.GetThrowTimes()] = a
	}
}
