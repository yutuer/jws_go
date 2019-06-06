package gamedata

import (
	"sort"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func (hd HotDatas) checkHeroGachaRace() error {
	for _, shard := range game.Cfg.ShardId {
		tsc := GetShardValidAct(uint32(shard), hd.Activity)
		if tsc == nil || len(tsc) <= 0 {
			continue
		}
		sort.Sort(tsc)
		for i := 0; i < len(tsc)-1; i++ {
			if tsc[i].et > tsc[i+1].bt {
				return fmt.Errorf("checkHeroGachaRace time overlap, shard %d actId %d %d",
					shard, tsc[i].actid, tsc[i+1].actid)
			}
		}
	}
	return nil
}

type time_couple struct {
	actid uint32
	bt    int64
	et    int64
}

type time_couples []time_couple

func (pq time_couples) Len() int { return len(pq) }

func (a time_couples) Less(i, j int) bool {
	return a[i].bt < a[j].bt
}

func (pq time_couples) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
