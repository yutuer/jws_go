package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

// BoxItem 限时神将盒子
type BoxItem struct {
	ActivityID  int64    `codec:"acbox_id"`      // 活动ID
	BoxID       int64    `codec:"box_id"`        // 盒子ID
	NeedPointID int64    `codec:"need_point_id"` // 需要的积分Id
	BoxItem     []string `codec:"boxitem_id"`    // 盒子奖励物品ID
	BoxItemNum  []int64  `codec:"boxitem_num"`   // 盒子奖励物品数量
}

func (s *SyncResp) makeGachaBox() {
	//hotGachaPoint := gamedata.GetHotDatas().HotLimitHeroGachaData.HeroGachaRaceChestInfo
	hotGachaInfo := gamedata.GetHotDatas().HotLimitHeroGachaData.HeroGachaRaceChestReward
	s.HotLimitGachaBox = make([][]byte, 0)
	for gacha, _ := range hotGachaInfo {
		s.HotLimitGachaBox = append(s.HotLimitGachaBox, encode(convertGachaBox2Client(gacha.ActivityType, gacha.Index)))
	}
}

func convertGachaBox2Client(activity int64, index uint32) BoxItem {
	data := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHeroGachaRaceChestRewardByType(activity, index)
	point, _ := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHeroGachaRaceChestInfoByType(activity, index)
	boxItem := make([]string, 0)
	boxItemNum := make([]int64, 0)
	for _, key := range data {
		boxItem = append(boxItem, key.GetItemID())
		boxItemNum = append(boxItemNum, int64(key.GetItemNum()))
	}
	return BoxItem{
		ActivityID:  activity,
		BoxID:       int64(index),
		NeedPointID: int64(point),
		BoxItem:     boxItem,
		BoxItemNum:  boxItemNum,
	}
}

// hgrRanks 限时神将排行奖励
type hgrRanks struct {
	ActivityID    int64    `codec:"acbox_id"`        // 活动ID
	ID            int64    `codec:"id"`              // ID
	NeedRankPoint int64    `codec:"need_rank_point"` // rank需要达到的积分
	UpperLimit    int64    `codec:"upper_limit"`     // 排行榜名次上限
	LowerLimit    int64    `codec:"lower_limit"`     // 排行榜名次下限
	BoxItemID     []string `codec:"boxitem_id"`      // 盒子奖励物品id
	BoxItemNum    []int64  `codec:"boxitem_num"`     // 盒子奖励物品数量
}

func (s *SyncResp) makeGachaRank() {
	gachaRank := gamedata.GetHotDatas().HotLimitHeroGachaData.GdHGRRank
	s.HotLimitGachaRank = make([][]byte, 0)
	for _, gachakey := range gachaRank {
		for _, gacha := range gachakey {
			s.HotLimitGachaRank = append(s.HotLimitGachaRank, encode(convertGachaRank2Client(gacha)))
		}
	}
}

func convertGachaRank2Client(data *ProtobufGen.HGRRANK) hgrRanks {
	itemId := make([]string, 0)
	itemNum := make([]int64, 0)
	itemId = append(itemId, data.GetLoot_Table()[0].GetItemID())
	itemNum = append(itemNum, int64(data.GetLoot_Table()[0].GetItemNum()))
	return hgrRanks{
		ActivityID:    int64(data.GetActivityID()),
		ID:            int64(data.GetBoxID()),
		NeedRankPoint: int64(data.GetRankNeedPoint()),
		UpperLimit:    int64(data.GetRankUp()),
		LowerLimit:    int64(data.GetRankDown()),
		BoxItemID:     itemId,
		BoxItemNum:    itemNum}
}

// HgrOptions 限时神将操作
type hgrOptions struct {
	ActivityID  int64 `codec:"ac_id"`      // 运营类型ID
	LikeHeroID  int64 `codec:"lh_id"`      // 相关HeroID
	LikeGachaID int64 `codec:"lh_gachaid"` // 相关gachaID
}

func (s *SyncResp) makeGachaOption() {
	gachaOption := gamedata.GetHotDatas().HotLimitHeroGachaData.GdHGRGacha
	gachaConfig := gamedata.GetHotDatas().HotLimitHeroGachaData.GdHGRConfig
	s.HgrOptions = make([][]byte, 0)
	s.GachaRacePoint = int64(gachaConfig.GetGachaRacePoint())
	s.PublicityTime = int64(gachaConfig.GetPublicityTime())
	for _, gacha := range gachaOption {
		s.HgrOptions = append(s.HgrOptions, encode(convertGachaOption2Client(gacha)))
	}

}

func convertGachaOption2Client(data *ProtobufGen.HGRGACHAOPTION) hgrOptions {
	return hgrOptions{
		ActivityID:  int64(data.GetActivityID()),
		LikeHeroID:  int64(data.GetHeroID()),
		LikeGachaID: int64(data.GetGachaID())}
}
