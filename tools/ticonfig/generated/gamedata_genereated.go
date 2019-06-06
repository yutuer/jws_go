package gamedata

import (
	"path/filepath"

	"sort"

	"github.com/gogo/protobuf/proto"
	"vcs.taiyouxi.net/comic/gamedata/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/**
自动生成的文件，不要修改
*/
func loadAllDatas(rootPath, dataAbsPath string) {
	load := func(dfilepath string, loadfunc func(string)) {
		loadfunc(filepath.Join(rootPath, dataAbsPath, dfilepath))
		logs.Info("LoadGameData %s success", dfilepath)
	}

	load("activityconfig.data", loadACTIVITYCONFIGData)
	load("level_info.data", loadLEVEL_INFOData)

}

var activityconfigs []*ProtobufGen.ACTIVITYCONFIG

func loadACTIVITYCONFIGData(filePath string) {
	buffer, err := loadBin(filePath)
	panicIfError(err)

	ar := new(ProtobufGen.ACTIVITYCONFIG_ARRAY)
	panicIfError(proto.Unmarshal(buffer, ar))

	activityconfigs = ar.GetItems()
		if len(activityconfigs) > 100 {
		sort.Slice(activityconfigs, func(i, j int) bool {
			return activityconfigs[i].GetActivityTimes() < activityconfigs[j].GetActivityTimes()
		})
	}
}

func GetAllACTIVITYCONFIGS() []*ProtobufGen.ACTIVITYCONFIG {
	return activityconfigs
}

func GetACTIVITYCONFIGById(ActivityTimes uint32) *ProtobufGen.ACTIVITYCONFIG {
	if len(activityconfigs) > 100 {
		index := sort.Search(len(activityconfigs), func(i int) bool {
			return activityconfigs[i].GetActivityTimes() >= ActivityTimes
		})
		if activityconfigs[index].GetActivityTimes() != ActivityTimes {
			return nil
		} else {
			return activityconfigs[index]
		}
	} else {
		for _, item := range activityconfigs {
			if item.GetActivityTimes() == ActivityTimes {
				return item
			}
		}
	}
	return nil
}


var level_infos []*ProtobufGen.LEVEL_INFO
var level_infomap map[string]*ProtobufGen.LEVEL_INFO

func loadLEVEL_INFOData(filePath string) {
	buffer, err := loadBin(filePath)
	panicIfError(err)

	ar := new(ProtobufGen.LEVEL_INFO_ARRAY)
	panicIfError(proto.Unmarshal(buffer, ar))

	level_infos = ar.GetItems()
		if len(level_infos) > 100 {
		level_infomap = make(map[string]*ProtobufGen.LEVEL_INFO, len(ar.GetItems()))
		for _, item := range level_infos {
			level_infomap[item.GetLevelID()] = item
		}
	}
}

func GetAllLEVEL_INFOS() []*ProtobufGen.LEVEL_INFO {
	return level_infos
}

func GetLEVEL_INFOById(LevelID string) *ProtobufGen.LEVEL_INFO {
	if len(level_infos) > 100 {
		return level_infomap[LevelID]
	} else {
		for _, item := range level_infos {
			if item.GetLevelID() == LevelID {
				return item
			}
		}
	}
	return nil
}


