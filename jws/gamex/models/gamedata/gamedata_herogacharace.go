package gamedata

import (
	"time"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func GetHGRCurrValidActivityId() uint32 {
	acts := GetShardValidAct(uint32(game.Cfg.ShardId[0]),
		GetHotDatas().Activity)
	if !game.Cfg.GetHotActValidData(game.Cfg.ShardId[0], uutil.Hot_Value_Limit_Hero) {
		return 0
	}
	if acts == nil || len(acts) <= 0 {
		return 0
	}
	now_t := time.Now().Unix()
	for _, act := range acts {
		if now_t >= act.bt && now_t < act.et {
			return act.actid
		}
	}
	return 0
}

func GetShardValidAct(shard uint32, hotActivity hotActivityData) time_couples {
	activitys := make([]uint32, 0)
	_acts := hotActivity.GetShardActivities(uint32(shard))
	if _acts != nil || len(_acts) > 0 {
		for _, value := range _acts {
			activitys = append(activitys, value)
		}
	}

	hgrHotId := hotActivity.GetHGRHotID(shard)
	if hgrHotId != 0 {
		activitys = append(activitys, hgrHotId)
	}

	if len(activitys) == 0 {
		return nil
	}

	acts := make(map[uint32]struct{}, len(_acts))
	for _, a := range activitys {
		acts[a] = struct{}{}
	}

	tsc := make(time_couples, 0, len(acts))
	for typ, v := range hotActivity.activityTypeInfoValid {
		if typ < ActHeroGachaRace_Begin || typ > ActHeroGachaRace_End {
			continue
		}
		for _, act := range v {
			_, ok := acts[act.ActivityId]
			if ok {
				tsc = append(tsc, time_couple{
					actid: act.ActivityId,
					bt:    act.StartTime,
					et:    act.EndTime,
				})
			}
		}
	}
	return tsc
}

func (d HeroGachaData) GetHGRRankConfig(activityId uint32, rank, score uint64) *ProtobufGen.HGRRANK {
	info := GetHotDatas().Activity.activityIdInfoValid[activityId]
	for _, r := range d.GdHGRRank[info.ActivityType] {
		if score >= uint64(r.GetRankNeedPoint()) &&
			rank <= uint64(r.GetRankDown()) {
			return r
		}
	}
	return nil
}

type hotHeroGachaRaceRank struct {
}

func (gr *hotHeroGachaRaceRank) loadData(buffer []byte, datas *HotDatas) error {
	ar := &ProtobufGen.HGRRANK_ARRAY{}
	if err := proto.Unmarshal(buffer, ar); err != nil {
		return err
	}
	as := ar.GetItems()
	datas.HotLimitHeroGachaData.GdHGRRank = make(map[uint32][]*ProtobufGen.HGRRANK, len(as))
	for _, r := range as {
		data, ok := datas.HotLimitHeroGachaData.GdHGRRank[r.GetActivityID()]
		if !ok {
			data = make([]*ProtobufGen.HGRRANK, 0, 10)
			datas.HotLimitHeroGachaData.GdHGRRank[r.GetActivityID()] = data
		}
		data = append(data, r)
		datas.HotLimitHeroGachaData.GdHGRRank[r.GetActivityID()] = data
	}
	return nil
}
