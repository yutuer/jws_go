package gamedata

import (
	"github.com/golang/protobuf/proto"
	"time"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const HGR_Chest_Count = 10

type HeroGachaData struct {
	HeroGachaRaceChestReward map[HeroGachaChestKey][]*ProtobufGen.HGRBOX_LootRule // 限时神将宝箱奖励
	HeroGachaRaceChestInfo   map[HeroGachaChestKey]uint32
	GdHGRRank                map[uint32][]*ProtobufGen.HGRRANK // activityTyp->
	GdHGRGacha               map[uint32]*ProtobufGen.HGRGACHAOPTION
	GdHGRGachas              map[uint32]struct{} // gachaid
	GdHGRConfig              *ProtobufGen.HGRCONFIG
}

type HeroGachaChestKey struct {
	ActivityType int64
	Index        uint32
}

func (d HeroGachaData) GetHeroGachaRaceChestInfo(activityID int64, index uint32) (uint32, bool) {
	info, had := GetHotDatas().Activity.activityIdInfoValid[uint32(activityID)]
	if !had {
		return 0, false
	}
	ret, had := d.HeroGachaRaceChestInfo[HeroGachaChestKey{
		ActivityType: int64(info.ActivityType), Index: index}]
	return ret, had
}

func (d HeroGachaData) GetHeroGachaRaceChestReward(activityID int64, index uint32) []*ProtobufGen.HGRBOX_LootRule {
	info := GetHotDatas().Activity.activityIdInfoValid[uint32(activityID)]
	return d.HeroGachaRaceChestReward[HeroGachaChestKey{
		ActivityType: int64(info.ActivityType), Index: index}]
}

func (d HeroGachaData) GetHeroGachaRaceChestRewardByType(activityType int64, index uint32) []*ProtobufGen.HGRBOX_LootRule {
	return d.HeroGachaRaceChestReward[HeroGachaChestKey{
		ActivityType: int64(activityType), Index: index}]
}

func (d HeroGachaData) GetHeroGachaRaceChestInfoByType(activityType int64, index uint32) (uint32, bool) {
	ret, had := d.HeroGachaRaceChestInfo[HeroGachaChestKey{
		ActivityType: int64(activityType), Index: index}]
	return ret, had
}

type hotHeroGachaRaceChest struct {
}

func (act *hotHeroGachaRaceChest) loadData(buffer []byte, datas *HotDatas) error {
	ar := &ProtobufGen.HGRBOX_ARRAY{}
	if err := proto.Unmarshal(buffer, ar); err != nil {
		return err
	}
	as := ar.GetItems()

	datas.HotLimitHeroGachaData.HeroGachaRaceChestInfo = make(map[HeroGachaChestKey]uint32, len(as))
	datas.HotLimitHeroGachaData.HeroGachaRaceChestReward = make(map[HeroGachaChestKey][]*ProtobufGen.HGRBOX_LootRule, len(as))
	for _, info := range as {
		key := HeroGachaChestKey{ActivityType: int64(info.GetActivityID()), Index: info.GetBoxID()}
		datas.HotLimitHeroGachaData.HeroGachaRaceChestInfo[key] = info.GetNeedPoint()

		tmp := make([]*ProtobufGen.HGRBOX_LootRule, 0, len(info.GetLoot_Table()))
		for i := 0; i < len(info.GetLoot_Table()); i++ {
			tmp = append(tmp, info.GetLoot_Table()[i])
		}
		datas.HotLimitHeroGachaData.HeroGachaRaceChestReward[key] = tmp
	}
	return nil
}

type hotHeroGachaRaceOption struct {
}

func (ho *hotHeroGachaRaceOption) loadData(buffer []byte, datas *HotDatas) error {
	ar := &ProtobufGen.HGRGACHAOPTION_ARRAY{}
	if err := proto.Unmarshal(buffer, ar); err != nil {
		return err
	}

	as := ar.GetItems()
	datas.HotLimitHeroGachaData.GdHGRGacha = make(map[uint32]*ProtobufGen.HGRGACHAOPTION, len(as))
	datas.HotLimitHeroGachaData.GdHGRGachas = make(map[uint32]struct{}, 5)
	for _, r := range as {
		datas.HotLimitHeroGachaData.GdHGRGacha[r.GetActivityID()] = r
		datas.HotLimitHeroGachaData.GdHGRGachas[r.GetGachaID()-1] = struct{}{}
	}
	return nil
}

func (d HeroGachaData) IsGachaHGR(gachaId uint32) bool {
	_, ok := d.GdHGRGachas[gachaId]
	return ok
}

func (d HeroGachaData) IsActivityGachaValid(activityId, gachaId uint32) bool {
	if activityId <= 0 {
		return false
	}
	info, ok := GetHotDatas().Activity.activityIdInfoValid[activityId]
	if !ok {
		return false
	}
	cfg, ok := d.GdHGRGacha[info.ActivityType]
	if !ok {
		return false
	}
	if cfg.GetGachaID()-1 != gachaId {
		return false
	}
	return true
}

type hotHeroGachaRaceConfig struct {
}

func (hc *hotHeroGachaRaceConfig) loadData(buffer []byte, datas *HotDatas) error {
	ar := &ProtobufGen.HGRCONFIG_ARRAY{}
	if err := proto.Unmarshal(buffer, ar); err != nil {
		return err
	}

	as := ar.GetItems()
	datas.HotLimitHeroGachaData.GdHGRConfig = as[0]
	return nil
}

func (d HeroGachaData) GetHGRConfig() *ProtobufGen.HGRCONFIG {
	return d.GdHGRConfig
}

func (d HeroGachaData) GetHGRCurrOptValidActivityId() uint32 {
	acts := GetShardValidAct(uint32(game.Cfg.ShardId[0]),
		GetHotDatas().Activity)
	if !game.Cfg.GetHotActValidData(game.Cfg.ShardId[0], uutil.Hot_Value_Limit_Hero) {
		return 0
	}
	if acts == nil || len(acts) <= 0 {
		return 0
	}
	now_t := time.Now().Unix()
	last_t := int64(d.GdHGRConfig.GetPublicityTime())
	for _, act := range acts {
		if now_t >= act.bt && now_t < act.et-last_t {
			return act.actid
		}
	}
	return 0
}
