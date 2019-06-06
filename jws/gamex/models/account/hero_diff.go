package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/hero_diff"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type HeroDiff struct {
	UsedHero         []int                `json:"used_hero" codec:"used_hero"`
	CurStageID       int                  `json:"-" codec:"cur_stage_id"`
	CurStageMaxScore [AVATAR_NUM_CURR]int `json:"-" codec:"hero_max_score"`

	Score         [gamedata.HeroDiff_Count]int                  `json:"score" codec:"-"`
	ScoreHero     [gamedata.HeroDiff_Count][AVATAR_NUM_CURR]int `json:"freq_hero" codec:"-"`
	LastMaxScore  [gamedata.HeroDiff_Count][AVATAR_NUM_CURR]int `json:"last_max_score" codec:"-"`
	LastResetTime int64                                         `json:"last_reset_time" codec: "-"`
	TodayStage    []int                                         `json:"today_stage" codec:"-"`
	CurStageIndex int                                           `json:"cur_stage_index" codec:"cur_stage_index"`
	LastCleanTime int64                                         `json:'last_clean_time" codec: "-"`
	IsLastStage   bool                                          `json:"-" codec:"is_last_stage"`
}

func (hd *HeroDiff) OnPassStage(usedHeroID int, score int) {
	hd.judgeInit()
	hd.UsedHero = append(hd.UsedHero, usedHeroID)
	hd.updateCurStageID()
	hd.Score[hero_diff.HeroDiffID2Index(hd.CurStageID)] += score
	hd.ScoreHero[hero_diff.HeroDiffID2Index(hd.CurStageID)][usedHeroID] += score
	if score > hd.LastMaxScore[hero_diff.HeroDiffID2Index(hd.CurStageID)][usedHeroID] {
		hd.LastMaxScore[hero_diff.HeroDiffID2Index(hd.CurStageID)][usedHeroID] = score
	}
	hd.CurStageIndex++
	hd.updateCurStageID()
	logs.Debug("next hero diff stage id is %d", hd.CurStageID)
}

func (hd *HeroDiff) GetLastMaxScore(heroID int) int {
	return hd.LastMaxScore[hero_diff.HeroDiffID2Index(hd.CurStageID)][heroID]
}

func (hd *HeroDiff) ClearMaxScore() {
	hd.LastMaxScore = [gamedata.HeroDiff_Count][AVATAR_NUM_CURR]int{}
}

func (hd *HeroDiff) GetCurStageID() int {
	hd.updateCurStageID()
	return hd.CurStageID
}

func (hd *HeroDiff) GetHeroDiffScore() [gamedata.HeroDiff_Count]int {
	return hd.Score
}

func (hd *HeroDiff) judgeInit() {
	if hd.UsedHero == nil {
		hd.UsedHero = make([]int, 0)
	}
	if hd.TodayStage == nil {
		hd.TodayStage = make([]int, 0)
	}
}

func (hd *HeroDiff) updateCurStageID() {
	if len(hd.TodayStage) <= 0 {
		return
	}
	if hd.CurStageIndex >= len(hd.TodayStage) {
		// 最后一关，没有下一关ID，仍传给前端最后一关ID
		hd.CurStageID = hd.TodayStage[len(hd.TodayStage)-1]
		hd.IsLastStage = true
	} else {
		hd.CurStageID = hd.TodayStage[hd.CurStageIndex]
		hd.IsLastStage = false
	}
	hd.CurStageMaxScore = hd.LastMaxScore[hero_diff.HeroDiffID2Index(hd.CurStageID)]
}

func (hd *HeroDiff) IsUsedHero(heroID int) bool {
	for _, item := range hd.UsedHero {
		if item == heroID {
			return true
		}
	}
	return false
}

func (hd *HeroDiff) UpdateTodayInfo(nowT int64) (isNew bool) {
	updateTime := util.DailyBeginUnixByStartTime(nowT,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	if updateTime != hd.LastResetTime {
		hd.CurStageIndex = 0
		hd.LastResetTime = updateTime
		hd.UsedHero = hd.UsedHero[:0]
		logs.Debug("update hero diff today info, nextTime is: %d, stageSeq is: %v", updateTime, hd.TodayStage)
		isNew = true
	}
	if len(hd.TodayStage) <= 0 {
		isNew = true
	}
	cleanTime := util.DailyBeginUnixByStartTime(nowT, gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypHeroDiffReset))
	if cleanTime != hd.LastCleanTime {
		hd.CleanStageData()
		logs.Debug("clean stage data at time: %d", nowT)
		hd.LastCleanTime = cleanTime
	}
	hd.updateCurStageID()
	return
}

func (hd *HeroDiff) UpdateTodayStage(stageSeq []int) {
	hd.TodayStage = stageSeq
	hd.updateCurStageID()
}

func (hd *HeroDiff) CleanStageData() {
	hd.UsedHero = hd.UsedHero[:0]
	hd.Score = [gamedata.HeroDiff_Count]int{}
	hd.ScoreHero = [gamedata.HeroDiff_Count][AVATAR_NUM_CURR]int{}
	logs.Debug("reset hero diff self data")
}

func (hd *HeroDiff) GetTopNFreqHero(stageIndex int, n int) []int {
	if n > AVATAR_NUM_CURR {
		n = AVATAR_NUM_CURR
	}
	hd.judgeInit()
	freqHero := hd.ScoreHero[stageIndex][:]
	topNHeroIdxs := make([]int, n) // topN heroIdx sort by desc
	for i := 0; i < n; i++ {
		topNHeroIdxs[i] = -1
	}
	for i, score := range freqHero {
		for j, idx := range topNHeroIdxs {
			if idx == -1 {
				topNHeroIdxs[j] = i
				break
			}
			if score > freqHero[idx] {
				rear := append([]int{}, topNHeroIdxs[j:]...)
				topNHeroIdxs = append(topNHeroIdxs[0:j], i)
				topNHeroIdxs = append(topNHeroIdxs, rear...)
				topNHeroIdxs = topNHeroIdxs[0:n]
				break
			}
		}
	}
	logs.Debug("hero diff top n freq avatar: %v", topNHeroIdxs)
	freqHeroNo0 := make([]int, 0)
	for _, i := range topNHeroIdxs {
		if freqHero[i] > 0 {
			freqHeroNo0 = append(freqHeroNo0, i)
		}
	}
	return freqHeroNo0
}
