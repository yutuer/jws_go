package general

import (
	"fmt"
	"time"

	"sync/atomic"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
)

type GeneralQuestInList struct {
	QuestId    int64  `json:"qid"`
	QuestCfgId string `json:"qcfgid"`
	IsRec      bool   `json:"rev"`
}

type GeneralQuestRec struct {
	QuestId    int64    `json:"qid"`
	QuestCfgId string   `json:"qcfgid"`
	FinishTime int64    `json:"ft"`
	GeneralIds []string `json:"gids"`
}

func (pg *PlayerGenerals) GQListUpdate(now_time int64) {
	if now_time < pg.QuestListNextRefTime {
		return
	}
	// 需要刷新任务
	t := time.Unix(now_time, 0).In(util.ServerTimeLocal)
	var tod bool
	cRefTs := gamedata.GeneralQuestRefreshTime()
	for _, chm := range cRefTs {
		ct, _ := time.ParseInLocation("2006-1-2 15:04",
			fmt.Sprintf("%d-%d-%d %s", t.Year(), t.Month(), t.Day(), chm),
			util.ServerTimeLocal)
		ctu := ct.Unix()
		if ctu > now_time { // 未跨天
			pg.QuestListNextRefTime = ctu
			tod = true
			break
		}
	}
	if !tod { // 跨天
		chm := cRefTs[0]
		ct, _ := time.ParseInLocation("2006-1-2 15:04",
			fmt.Sprintf("%d-%d-%d %s", t.Year(), t.Month(), t.Day(), chm),
			util.ServerTimeLocal)
		pg.QuestListNextRefTime = ct.Unix() + util.DaySec
	}
	pg.GQListUpdateForce()
}

func (pg *PlayerGenerals) GQListUpdateForce() {
	// 刷新任务
	rare2num := pg.generalRareLvlNum()
	allQ := make(map[string]struct{}, 256)
	qsCond := gamedata.GeneralQuestCondCfg()
	for cond, qs := range qsCond {
		n := rare2num[uint32(cond.Param2)]
		if n >= cond.Param1 {
			for _, q := range qs {
				allQ[q] = struct{}{}
			}
		}
	}

	// 利用map的随机，来随机任务
	pg.QuestList = make([]GeneralQuestInList, 0, gamedata.GeneralQuestListMax)
	i := 0
	for k, _ := range allQ {
		pg.QuestList = append(pg.QuestList, GeneralQuestInList{
			QuestId:    pg.nextQuestId(),
			QuestCfgId: k,
		})
		i++
		if i >= gamedata.GeneralQuestListMax {
			break
		}
	}
}

func (pg *PlayerGenerals) GetQuestInList(qid int64) *GeneralQuestInList {
	for i := 0; i < len(pg.QuestList); i++ {
		q := &pg.QuestList[i]
		if q.QuestId == qid {
			return q
		}
	}
	return nil
}

func (pg *PlayerGenerals) GetQuestInRec(qid int64) (*GeneralQuestRec, int) {
	for i := 0; i < len(pg.QuestRec); i++ {
		q := &pg.QuestRec[i]
		if q.QuestId == qid {
			return q, i
		}
	}
	return nil, 0
}

func (pg *PlayerGenerals) DelQuestInRec(idx int) {
	if len(pg.QuestRec) <= 0 {
		return
	}
	if idx == len(pg.QuestRec)-1 {
		pg.QuestRec = pg.QuestRec[:idx]
	} else {
		tq := make([]GeneralQuestRec, 0, len(pg.QuestRec)-1)
		a := pg.QuestRec[:idx]
		b := pg.QuestRec[idx+1:]
		tq = append(tq, a...)
		tq = append(tq, b...)
		pg.QuestRec = tq
	}
}

func (pg *PlayerGenerals) GeneralUsedByQuest(gens []string) bool {
	if len(gens) <= 0 {
		return false
	}
	gs := make(map[string]struct{}, len(gens))
	for _, g := range gens {
		gs[g] = struct{}{}
	}
	for _, q := range pg.QuestRec {
		for _, g := range q.GeneralIds {
			if _, ok := gs[g]; ok {
				return true
			}
		}
	}
	return false
}

func (pg *PlayerGenerals) ReceiveQuest(q *GeneralQuestInList,
	gens []string, finishTime int64) {
	q.IsRec = true
	pg.QuestRec = append(pg.QuestRec, GeneralQuestRec{
		QuestId:    q.QuestId,
		QuestCfgId: q.QuestCfgId,
		FinishTime: finishTime,
		GeneralIds: gens,
	})
}

func (pg *PlayerGenerals) generalRareLvlNum() map[uint32]int64 {
	res := make(map[uint32]int64, 8)
	for _, g := range pg.GeneralAr {
		data := gamedata.GetGeneralInfo(g.Id)
		if data != nil {
			if g.IsHas() {
				n := res[data.GetRareLevel()]
				n += 1
				res[data.GetRareLevel()] = n
			}
		}
	}
	return res
}

func (pg *PlayerGenerals) nextQuestId() int64 {
	return atomic.AddInt64(&pg.QuestNextId, 1)
}
