package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type HeroDestinyConfig struct {
	Cfg       *ProtobufGen.FATETABLE
	AvatarIds []int
}

var heroDestinyData []*HeroDestinyConfig
var heroDestinyLevelData []*ProtobufGen.FATELEVEL
var heroDestinyMaxLevelMap map[int]int // 记录每种情缘的最大等级

func loadHeroDestinyData(filePath string) {
	buffer, err := loadBin(filePath)
	panicIfErr(err)

	ar := new(ProtobufGen.FATETABLE_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	heroDestinyData = make([]*HeroDestinyConfig, 0)
	for _, item := range ar.GetItems() {
		cfg := &HeroDestinyConfig{
			Cfg:       item,
			AvatarIds: make([]int, len(item.GetHeroList_Table())),
		}
		for i, strId := range item.GetHeroList_Table() {
			cfg.AvatarIds[i] = GetHeroByHeroID(strId.GetHero())
		}
		heroDestinyData = append(heroDestinyData, cfg)
	}

	for index, item := range heroDestinyData {
		if index != int(item.Cfg.GetFateID()-1) {
			panic("hero destiny data err, index not match id")
		}
	}
}

func GetHeroDestinyById(id int) *HeroDestinyConfig {
	if id < 0 || id > len(heroDestinyData) {
		logs.Error("try to get hero destiny cfg that not exsit, %d", id)
		return nil
	}
	return heroDestinyData[id-1]
}

func (h *HeroDestinyConfig) ContainsAvatarId(avatarId int) bool {
	for _, id := range h.AvatarIds {
		if id == avatarId {
			return true
		}
	}
	return false
}

func loadFateLevelData(filePath string) {
	buffer, err := loadBin(filePath)
	panicIfErr(err)

	ar := new(ProtobufGen.FATELEVEL_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	heroDestinyLevelData = ar.GetItems()

	heroDestinyMaxLevelMap = make(map[int]int)
	for _, item := range heroDestinyLevelData {
		if v, ok := heroDestinyMaxLevelMap[int(item.GetFateID())]; ok {
			if int(item.GetFateLevel()) > v {
				heroDestinyMaxLevelMap[int(item.GetFateID())] = int(item.GetFateLevel())
			}
		} else {
			heroDestinyMaxLevelMap[int(item.GetFateID())] = int(item.GetFateLevel())
		}
	}
}

func GetFateMaxLevel(fateId int) int {
	return heroDestinyMaxLevelMap[fateId]
}

func GetFateLevelConfig(fateId int, level int) *ProtobufGen.FATELEVEL {
	for _, item := range heroDestinyLevelData {
		if int(item.GetFateID()) == fateId && int(item.GetFateLevel()) == level {
			return item
		}
	}
	return nil
}
