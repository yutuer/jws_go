package itemHistory

import (
	"math"
	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type ItemHistory struct {
	// 对应表中的ID位置存储累计掉落量
	History   [][]int64 `json:"history"`
	LastLoot  []string
	LastCount []uint32
}

func (h *ItemHistory) Init() {
	if h.History == nil || len(h.History) == 0 {
		h.History = make([][]int64, 1, 1)
		h.History[0] = make([]int64, 64, 64)
		h.LastLoot = make([]string, 0, 64)
		h.LastCount = make([]uint32, 0, 64)
	}
}

func (h *ItemHistory) Add(id int, count uint32) {
	for id >= len(h.History[0]) {
		h.History[0] = append(h.History[0], 0)
	}

	h.History[0][id] += int64(count)
}

func (h *ItemHistory) AddLastLoot(itemId string, count uint32) {
	for i := 0; i < len(h.LastLoot); i++ {
		if h.LastLoot[i] == itemId {
			h.LastCount[i]++
			return
		}
	}
	h.LastLoot = append(h.LastLoot, itemId)
	h.LastCount = append(h.LastCount, count)
}

func (h *ItemHistory) CleanLastLoot() {
	h.LastLoot = h.LastLoot[0:0]
	h.LastCount = h.LastCount[0:0]
}

func (h *ItemHistory) GetRandLootByHistory(rd *rand.Rand) (string, int) {
	datas := gamedata.GetUniversalMaterialDatas()
	pool := make([]int, 0, 32)
	var lvInPool int = math.MaxInt32

	for _, item := range datas {
		var hasLoot int64
		if item.ID >= 0 && item.ID < len(h.History[0]) {
			hasLoot = h.History[0][item.ID]
		}

		if hasLoot >= item.AllNeed {
			continue
		}

		if item.Lv < lvInPool {
			pool = pool[0:0]
			lvInPool = item.Lv
		}

		if item.Lv == lvInPool {
			pool = append(pool, item.ID)
		}
	}

	if len(pool) == 0 {
		// 所有的都满了
		pool, lvInPool = gamedata.GetUniversalMaterialMaxLvData()
	}

	logs.Trace("lv %d in pool %v", lvInPool, pool)
	idx := pool[rd.Int()%len(pool)]
	return datas[idx].ItemID, idx
}
