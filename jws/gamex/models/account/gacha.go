package account

import (
	"math/rand"

	"strconv"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/mn_selector"
)

const ( // 注意实际请求的id是从0开始 所以比较时要减一
	GachaHCOneID = 1 // 注意实际请求的id是从0开始 所以比较时要减一
	GachaHCTenID = 4 // 注意实际请求的id是从0开始 所以比较时要减一
	GachaVIP     = 5 // 注意实际请求的id是从0开始 所以比较时要减一
)

type GachaState struct {
	LastFreeTime   int64 `json:"ft"`
	TodayFreeCount int   `json:"fc"`

	RewardExtRewardCount int64 `json:"erc"`

	RewardSerialCount int64 `json:"rc"`

	SpecInNormalMN   mnSelector.MNSelectorState `json:"snmn"`
	TreasureInSpecMN mnSelector.MNSelectorState `json:"tsmn"`

	HistoryCount int64 `json:"hc"`

	HeroGachaRaceCount int64 `json:"cc"`
}

func (g *GachaState) AddRewardExtRewardCount(player_corp_lv uint32, idx int) {
	if g.RewardExtRewardCount == 0 {
		data := gamedata.GetGachaData(player_corp_lv, idx)
		if data == nil {
			logs.Error("Gacha Data Err By %d", idx)
		} else {
			g.RewardExtRewardCount = int64(data.ExtraStartNum) + 1
		}
	}

	g.RewardExtRewardCount++
}

func (g *GachaState) GetRewardExtRewardCount(player_corp_lv uint32, idx int) int64 {
	if g.RewardExtRewardCount == 0 {
		data := gamedata.GetGachaData(player_corp_lv, idx)
		if data == nil {
			logs.Error("Gacha Data Err By %d", idx)
		} else {
			g.RewardExtRewardCount = int64(data.ExtraStartNum) + 1
		}
	}

	return g.RewardExtRewardCount - 1
}

func (g *GachaState) IsCanFree(player_corp_lv uint32, now_time int64, idx int) bool {
	data := gamedata.GetGachaData(player_corp_lv, idx)
	if data == nil {
		logs.Error("Gacha Data Err By %d", idx)
		return false
	}

	if data.FreeCoolTime <= 0 {
		return false
	}

	if data.FreeCountEveryOneDay <= g.TodayFreeCount {
		return false
	}

	return now_time >= data.FreeCoolTime+g.LastFreeTime
}

func (g *GachaState) SetUseFreeNow(now_time int64) {
	g.LastFreeTime = now_time
	g.TodayFreeCount += 1
}

func (g *GachaState) GetSerialRewardInfo(player_corp_lv uint32, idx, avatar_id int) (string, uint32) {
	data := gamedata.GetGachaData(player_corp_lv, idx)
	if data == nil {
		logs.Error("Gacha Data Err By %d", idx)
		return "", 0
	}

	if len(data.RewardSerial[avatar_id]) == 0 {
		g.RewardSerialCount = 0
		return "", 0
	}

	if g.RewardSerialCount < 0 {
		// 这个要很久很久很久以后才会出现
		g.RewardSerialCount = 0
	}

	ridx := g.RewardSerialCount % int64(len(data.RewardSerial[avatar_id]))
	reward := data.RewardSerial[avatar_id][int(ridx)]
	return reward.Id, reward.Count
}

func (g *GachaState) GetExtReward(player_corp_lv uint32, idx, avatar_id int, rd *rand.Rand) *gamedata.GachaReward {
	data := gamedata.GetGachaData(player_corp_lv, idx)
	if data == nil {
		logs.Error("Gacha Data Err By %d", idx)
		return nil
	}

	if data.ExtraSpace == 0 {
		return nil
	}
	g.AddRewardExtRewardCount(player_corp_lv, idx)
	extCount := g.GetRewardExtRewardCount(player_corp_lv, idx)
	if extCount < 0 {
		return nil
	}

	isExt := (extCount % int64(data.ExtraSpace)) == 0

	if isExt {
		pool := data.ExtraGroupRewardPool.GetByAvatar()
		ridx := pool.Rander.Rand(rd)
		if ridx < 0 || ridx >= len(pool.ToSelect) {
			logs.Error("Gacha ridx Err By %d %v %d",
				ridx, pool.ToSelect, pool.PoolId)
			return nil
		}

		logs.Trace("ridx %v %v", ridx, *pool)

		return &(pool.ToSelect[ridx])
	}
	return nil
}

// 注意这个会将现有Count计数加一
func (g *GachaState) GetSerialReward(player_corp_lv uint32, idx, avatar_id int) *gamedata.GachaReward {
	data := gamedata.GetGachaData(player_corp_lv, idx)
	if data == nil {
		logs.Error("Gacha Data Err By %d", idx)
		return nil
	}

	if len(data.RewardSerial[avatar_id]) == 0 {
		g.RewardSerialCount = 0
		return nil
	}

	if g.RewardSerialCount < 0 {
		// 这个要很久很久很久以后才会出现
		g.RewardSerialCount = 0
	}

	ridx := g.RewardSerialCount % int64(len(data.RewardSerial[avatar_id]))

	logs.Trace("SerialReward %d %v", int(ridx), data.RewardSerial[avatar_id][int(ridx)])

	return &data.RewardSerial[avatar_id][int(ridx)]
}

func (g *GachaState) Gacha(acid string, player_corp_lv uint32, avatar_id, idx, hc_t, gacha_count int, rd *rand.Rand) *gamedata.GachaReward {
	data := gamedata.GetGachaData(player_corp_lv, idx)
	if data == nil {
		logs.Error("Gacha Data Err By %d", idx)
		return nil
	}

	if g.HistoryCount == 0 && data.FirstGive.Id != "" {
		// 首次必掉
		g.HistoryCount++
		return &data.FirstGive
	}

	logs.Trace("[%s]selectPool %d %d %v %d", acid, avatar_id, gacha_count, data, hc_t)
	pool := g.selectPool(acid, avatar_id, gacha_count, data, hc_t, rd)
	if pool == nil {
		logs.Error("Gacha selectPool Err By %d", idx)
		return nil
	}

	ridx := pool.Rander.Rand(rd)
	if ridx < 0 || ridx >= len(pool.ToSelect) {
		logs.Error("Gacha ridx Err By %d %v %d",
			ridx, pool.ToSelect, pool.PoolId)
		return nil
	}

	g.HistoryCount++
	return &pool.ToSelect[ridx]
}

func (g *GachaState) selectPool(
	acid string,
	avatar_id, gacha_count int,
	data *gamedata.GachaData,
	hc_t int,
	rd *rand.Rand) *gamedata.GachaRewardPoolForOneAvatar {

	if g.SpecInNormalMN.IsNowNeedNewTurn() {
		g.SpecInNormalMN.Reset(
			int64(data.SpecInNormalNum),
			int64(data.SpecInNormalSpace))
	}

	isSelected := g.SpecInNormalMN.Selector(rd)

	g.SpecInNormalMN.LogicLog(
		acid,
		logiclog.LogType_GachaMN,
		strconv.Itoa(gacha_count))

	if isSelected {
		/*
			玩家使用免费钻进行抽奖的时候，若随机到了珍稀物品组，则需要再进行一次概率判定，
			若通过则继续随机珍稀物品，否则以普通物品组继续往下走。
			将钻石分为付费钻（A）、返利钻（B）、和补偿钻（C）三类，
			// 不一样 -> 每次随机按照每种钻当前的数量得到其被随机到的概率，根据随机结果扣除相应类型的钻。
			若随机到珍稀物品，则按照每种钻的权重判定是否将珍稀物品发放给玩家，
			判定失败则进行一次普通随机补偿。
			若发生被随机到的钻数量不足一次抽奖，则按照优先级的从高到低补齐差额（默认A>B>C），
			当取权重时，以消耗的最高优先级的钻石所对应的权重为最终判断所使用的权重。
		*/
		is_hc_spec := gamedata.IsGachaToSpecPool(hc_t, rd)
		logs.Trace("IsGachaToSpecPool %v to Spec!", is_hc_spec)
		return g.selectPoolSpecOrTreasure(acid, avatar_id, data, is_hc_spec, rd)
	} else {
		return g.selectPoolNormal(data, rd)
	}

}

func (g *GachaState) selectPoolSpecOrTreasure(acid string, avatar_id int, data *gamedata.GachaData, is_hc_spec bool, rd *rand.Rand) *gamedata.GachaRewardPoolForOneAvatar {
	if g.TreasureInSpecMN.IsNowNeedNewTurn() {
		g.TreasureInSpecMN.Reset(
			int64(data.TreasureInSpaceNum),
			int64(data.TreasureInSpaceSpace),
		)
	}

	isSelected := g.TreasureInSpecMN.Selector(rd)

	g.TreasureInSpecMN.LogicLog(
		acid,
		logiclog.LogType_GachaMN,
		strconv.Itoa(avatar_id))

	if isSelected && is_hc_spec {
		return g.selectPoolTreasure(data, rd)
	} else {
		return g.selectPoolSpec(data, rd)
	}

}

func (g *GachaState) selectPoolNormal(data *gamedata.GachaData, rd *rand.Rand) *gamedata.GachaRewardPoolForOneAvatar {
	idx := data.NormalPoolRander.Rand(rd)
	logs.Trace("selectPoolNormal %d, %v", idx, data.NormalPool)
	if idx < uint32(len(data.NormalPool)) {
		return data.NormalPool[idx].GetByAvatar()
	}
	return nil
}

func (g *GachaState) selectPoolSpec(data *gamedata.GachaData, rd *rand.Rand) *gamedata.GachaRewardPoolForOneAvatar {
	return data.SpecPool.GetByAvatar()
}

func (g *GachaState) selectPoolTreasure(data *gamedata.GachaData, rd *rand.Rand) *gamedata.GachaRewardPoolForOneAvatar {
	return data.TreasurePool.GetByAvatar()
}

type PlayerGacha struct {
	Gacha          [gamedata.GachaMaxCount]GachaState `json:"g"`
	LastUpdateTime int64                              `json:"t"`
}

func (p *PlayerGacha) Update(now_time int64) {
	logs.Trace("PlayerGacha Update %d %d", now_time, p.LastUpdateTime)
	if !gamedata.IsSameDayCommon(now_time, p.LastUpdateTime) {
		logs.Trace("PlayerGacha clean Free Count")
		for i := 0; i < len(p.Gacha); i++ {
			p.Gacha[i].TodayFreeCount = 0
		}
		p.LastUpdateTime = now_time
	}
}

func (p *PlayerGacha) GetGachaStat(idx int) *GachaState {
	if idx < 0 || idx >= len(p.Gacha) {
		return nil
	}

	return &p.Gacha[idx]
}
