package gamedata

import (
	"math"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdTrialFirstLvlId int32
	gdTrialFinalLvlId int32
	gdTrialLvlInfo    map[int32]*ProtobufGen.LEVEL_TRIAL
	gdTrialLvlOrder   map[int32]*ProtobufGen.LEVEL_TRIAL
)

func loadTrial(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.LEVEL_TRIAL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdTrialLvlInfo = make(map[int32]*ProtobufGen.LEVEL_TRIAL, len(ar.Items))
	gdTrialLvlOrder = make(map[int32]*ProtobufGen.LEVEL_TRIAL, len(ar.Items))
	var minIndex int32
	minIndex = math.MaxInt32
	var maxIndex int32
	for _, rec := range ar.Items {
		gdTrialLvlInfo[rec.GetLevelID()] = rec
		gdTrialLvlOrder[rec.GetTrialIndex()] = rec
		if rec.GetTrialIndex() < minIndex {
			minIndex = rec.GetTrialIndex()
			gdTrialFirstLvlId = rec.GetLevelID()
		}
		if rec.GetTrialIndex() > maxIndex {
			maxIndex = rec.GetTrialIndex()
			gdTrialFinalLvlId = rec.GetLevelID()
		}
	}
}

func GetTrialLvlById(lvlId int32) *ProtobufGen.LEVEL_TRIAL {
	return gdTrialLvlInfo[lvlId]
}

func GetTrialLvlByIndex(lvlIdx int32) *ProtobufGen.LEVEL_TRIAL {
	return gdTrialLvlOrder[lvlIdx]
}

func GetTrialFirstLvlId() int32 {
	return gdTrialFirstLvlId
}

func GetTrialFinalLvlId() int32 {
	return gdTrialFinalLvlId
}
