package gamedata

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	HeroDiff_TU    = helper.HeroDiff_TU
	HeroDiff_ZHAN  = helper.HeroDiff_ZHAN
	HeroDiff_HU    = helper.HeroDiff_HU
	HeroDiff_SHI   = helper.HeroDiff_SHI
	HeroDiff_Count = helper.HeroDiff_Count
)

const HeroDiffRankShowAvatarCount = 3

var (
	heroDiffStageCount   int
	heroDiffTypeCount    = []int{1, 2, 3, 4}
	heroDiffRewardData   []*ProtobufGen.HDPREWARD
	heroDiffMaxScore     map[int]*ProtobufGen.HDPREWARD
	heroDiffEnemyData    map[string]*ProtobufGen.HDPENEMY
	heroDiffLevelSection []*ProtobufGen.HDPLEVELSECTION
	heroDiffRewardList   []*ProtobufGen.HDPREWARDLIST
	heroDiffConfig       *ProtobufGen.HDPCONFIG
	heroDiffModel        map[uint32]*ProtobufGen.HDPMODEL
	heroDiffFZBTU        map[uint32]*ProtobufGen.FZBTU
	heroDiffFZBHU        map[uint32]*ProtobufGen.FZBHU
	heroDiffFZBZHAN      map[uint32]*ProtobufGen.FZBZHAN
	heroDiffFZBSHI       map[uint32]*ProtobufGen.FZBSHI
	heroDiffTUPoint      []*ProtobufGen.TUPOINT
	heroDiffZHANPoint    []*ProtobufGen.ZHANPOINT
)

func loadHeroDiffRewardData(filepath string) {
	ar := &ProtobufGen.HDPREWARD_ARRAY{}
	_common_load(filepath, ar)
	heroDiffRewardData = ar.GetItems()
	heroDiffMaxScore = make(map[int]*ProtobufGen.HDPREWARD, 0)
	for _, item := range heroDiffRewardData {
		if v, ok := heroDiffMaxScore[int(item.GetHeroDiffLevel())]; ok {
			if item.GetGVGPoint2() > v.GetGVGPoint2() {
				heroDiffMaxScore[int(item.GetHeroDiffLevel())] = item
			}
		} else {
			heroDiffMaxScore[int(item.GetHeroDiffLevel())] = item
		}
	}
}

func loadHeroDiffEnemyData(filepath string) {
	ar := &ProtobufGen.HDPENEMY_ARRAY{}
	_common_load(filepath, ar)
	heroDiffEnemyData = make(map[string]*ProtobufGen.HDPENEMY, len(ar.GetItems()))
	for _, item := range ar.GetItems() {
		heroDiffEnemyData[item.GetBossID()] = item
	}
}

func loadHeroDiffConfig(filepath string) {
	ar := &ProtobufGen.HDPCONFIG_ARRAY{}
	_common_load(filepath, ar)
	heroDiffConfig = ar.GetItems()[0]
}

func GetHeroDiffEnemyData(stageID string) *ProtobufGen.HDPENEMY {
	return heroDiffEnemyData[stageID]
}

func GetHeroDiffConfig() *ProtobufGen.HDPCONFIG {
	return heroDiffConfig
}

func loadHeroDiffLevelData(filepath string) {
	ar := &ProtobufGen.HDPLEVEL_ARRAY{}
	_common_load(filepath, ar)
	heroDiffStageCount = len(ar.GetItems())
	if heroDiffStageCount != len(heroDiffTypeCount) {
		logs.Error("fatal error by herodiff gamedata")
	}
}

func loadHeroDiffLevelSection(filepath string) {
	ar := &ProtobufGen.HDPLEVELSECTION_ARRAY{}
	_common_load(filepath, ar)
	heroDiffLevelSection = ar.GetItems()
}

func loadHeroDiffRewardList(filepath string) {
	ar := &ProtobufGen.HDPREWARDLIST_ARRAY{}
	_common_load(filepath, ar)
	heroDiffRewardList = ar.GetItems()
}

func loadHeroDiffModelData(filepath string) {
	ar := &ProtobufGen.HDPMODEL_ARRAY{}
	_common_load(filepath, ar)
	heroDiffModel = make(map[uint32]*ProtobufGen.HDPMODEL, 100)
	for _, item := range ar.GetItems() {
		heroDiffModel[item.GetPlayLevel()] = item
	}
}

func loadHeroDiffFZBTU(filepath string) {
	ar := &ProtobufGen.FZBTU_ARRAY{}
	_common_load(filepath, ar)
	heroDiffFZBTU = make(map[uint32]*ProtobufGen.FZBTU, 100)
	for _, item := range ar.GetItems() {
		heroDiffFZBTU[item.GetPlayerLevel()] = item
	}
	logs.Debug("heroDiffFZBTU: %v", heroDiffFZBHU)
}

func loadHeroDiffFZBSHI(filepath string) {
	ar := &ProtobufGen.FZBSHI_ARRAY{}
	_common_load(filepath, ar)
	heroDiffFZBSHI = make(map[uint32]*ProtobufGen.FZBSHI, 100)
	for _, item := range ar.GetItems() {
		heroDiffFZBSHI[item.GetPlayerLevel()] = item
	}
	logs.Debug("heroDiffFZBSHI: %v", heroDiffFZBSHI)
}

func loadHeroDiffFZBZHAN(filepath string) {
	ar := &ProtobufGen.FZBZHAN_ARRAY{}
	_common_load(filepath, ar)
	heroDiffFZBZHAN = make(map[uint32]*ProtobufGen.FZBZHAN, 100)
	for _, item := range ar.GetItems() {
		heroDiffFZBZHAN[item.GetPlayerLevel()] = item
	}
	logs.Debug("heroDiffFZBZHAN: %v", heroDiffFZBHU)
}

func loadHeroDiffFZBHU(filepath string) {
	ar := &ProtobufGen.FZBHU_ARRAY{}
	_common_load(filepath, ar)
	heroDiffFZBHU = make(map[uint32]*ProtobufGen.FZBHU, 100)
	for _, item := range ar.GetItems() {
		heroDiffFZBHU[item.GetPlayerLevel()] = item
	}
	logs.Debug("heroDiffFZBHU: %v", heroDiffFZBHU)
}

func loadHeroDiffTUPointData(filepath string) {
	ar := &ProtobufGen.TUPOINT_ARRAY{}
	_common_load(filepath, ar)
	heroDiffTUPoint = ar.GetItems()
}

func loadHeroDiffZHANPointData(filepath string) {
	ar := &ProtobufGen.ZHANPOINT_ARRAY{}
	_common_load(filepath, ar)
	heroDiffZHANPoint = ar.GetItems()
}

func CheckHeroDiffFZBScore(typ int, gs uint32, level uint32, score uint32) bool {
	recommandGS := heroDiffModel[level].GetShowPower()
	logs.Debug("param: %v, %v, %v, %v", gs, recommandGS, level, score, typ)
	switch typ {
	case HeroDiff_TU:
		return checkHeroDiffTU(gs, recommandGS, level, score)
	case HeroDiff_ZHAN:
		return checkHeroDiffZHAN(gs, recommandGS, level, score)
	case HeroDiff_HU:
		return checkHeroDiffHU(gs, recommandGS, level, score)
	case HeroDiff_SHI:
		return checkHeroDiffSHI(gs, recommandGS, level, score)
	default:
		logs.Error("hero diff level type err for typ: %v", typ)
	}
	return true
}

func checkHeroDiffTU(gs, rgs, level, score uint32) bool {

	data := heroDiffFZBTU[level]
	if data == nil {
		logs.Error("no data info for level: %v", level)
		return true
	}
	gsInfo := data.GetGS_Table()
	for _, item := range gsInfo {
		if gs < uint32(item.GetGSCoefficient()*float32(rgs)) && score > item.GetPointLimit() {
			return false
		}
	}
	// special check
	for i, item := range heroDiffTUPoint {
		if score > item.GetSubtrahend() && i+1 < len(heroDiffTUPoint) && score < heroDiffTUPoint[i+1].GetSubtrahend() {
			if (score-item.GetSubtrahend())%item.GetDivisor() != 0 {
				return false
			}
		}
	}
	return true
}

func checkHeroDiffZHAN(gs, rgs, level, score uint32) bool {
	data := heroDiffFZBZHAN[level]
	if data == nil {
		logs.Error("no data info for level: %v", level)
		return true
	}
	gsInfo := data.GetGS_Table()
	for _, item := range gsInfo {
		if gs < uint32(item.GetGSCoefficient()*float32(rgs)) && score > item.GetPointLimit() {
			return false
		}
	}
	// special check
	for _, item := range heroDiffZHANPoint {
		if item.GetSumPoint() == score {
			return true
		}
	}
	return false
}

func checkHeroDiffHU(gs, rgs, level, score uint32) bool {
	data := heroDiffFZBHU[level]
	if data == nil {
		logs.Error("no data info for level: %v", level)
		return true
	}
	gsInfo := data.GetGS_Table()
	for _, item := range gsInfo {
		if gs < uint32(item.GetGSCoefficient()*float32(rgs)) && score > item.GetPointLimit() {
			return false
		}
	}
	return true
}

func checkHeroDiffSHI(gs, rgs, level, score uint32) bool {
	data := heroDiffFZBSHI[level]
	if data == nil {
		logs.Error("no data info for level: %v", level)
		return true
	}
	gsInfo := data.GetGS_Table()
	for _, item := range gsInfo {
		if gs < uint32(item.GetGSCoefficient()*float32(rgs)) && score > item.GetPointLimit() {
			return false
		}
	}
	// special check
	if score%heroDiffConfig.GetShiDivisor() == 0 {
		return true
	}
	return false
}

func GetHeroDiffRewardData(score int, stageID int, level uint32) []*ProtobufGen.HDPREWARDLIST_LootRule {
	v, ok := heroDiffMaxScore[stageID]
	if !ok {
		return nil
	}
	// 定积分档位
	var rewardP *ProtobufGen.HDPREWARD
	if int(v.GetGVGPoint2()) < score {
		rewardP = v
	}
	for _, item := range heroDiffRewardData {
		if int(item.GetHeroDiffLevel()) == stageID && score <= int(item.GetGVGPoint2()) && score >= int(item.GetGVGPoint1()) {
			rewardP = item
		}
	}
	if rewardP == nil {
		return nil
	}
	scoreP := int(rewardP.GetPointRankID())
	// 定等级档位 如果表中没有玩家等级的数据，返回空奖励
	levelIndex := -1
	for _, item := range heroDiffLevelSection {
		if level <= item.GetLevelMax() && level >= item.GetLevelMin() {
			levelIndex = int(item.GetID())
		}
	}
	if levelIndex == -1 || levelIndex < 1 {
		return nil
	}
	rewardList := rewardP.GetRewardListID()
	if levelIndex > len(rewardList) {
		return nil
	}
	levelP := int(rewardList[levelIndex-1])
	for _, item := range heroDiffRewardList {
		if int(item.GetRewardListID()) == levelP && int(item.GetPointRankID()) == scoreP {
			return item.GetLoot_Table()
		}
	}
	return nil
}

func GetHeroDiffStageCount() int {
	return heroDiffStageCount
}

func GetHeroDiffTypeCount() []int {
	return heroDiffTypeCount
}
