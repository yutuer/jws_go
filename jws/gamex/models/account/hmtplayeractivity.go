package account

import (
	"fmt"
	"sort"
	"time"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type HmtPlayerActivityInfo struct {
	DateActivityInfo map[string]*DailyInfo `json:"date_activity_info"`
}

type DailyInfo struct {
	DailyDate     string  `json:"daily_date"`      //日期
	IsLogin       bool    `json:"is_login"`        //是否登录
	DungeonJy     []int   `json:"dungeon_jy"`      //精英副本通关次数
	DungeonJyTime []int64 `json:"dungeon_jy_time"` //精英副本通关时间戳
	DungeonDy     []int   `json:"dungeon_dy"`      //地狱副本通关次数
	DungeonDyTime []int64 `json:"dungeon_dy_time"` //地狱副本通关时间戳
	GachaNum      []int   `json:"gacha_num"`       //嘎查抽奖次数
	GachaNumTime  []int64 `json:"gacha_num_time"`  //嘎查抽奖时间戳
}

type HmtPlayerActivityInfoInDB struct {
	DateActivityInfo hmtActivityInfo_slice
}

func (pa *HmtPlayerActivityInfo) IsLogin(nowt int64) {
	date := formatDate(nowt)
	if pa.DateActivityInfo == nil || len(pa.DateActivityInfo) <= 0 {
		pa.DateActivityInfo = make(map[string]*DailyInfo)
	} else if _, ok := pa.DateActivityInfo[date]; ok {
		logs.Debug("[HMT Activity] Hmt Activity is login %v", pa.DateActivityInfo[date])
		return
	}
	pa.DateActivityInfo[date] = &DailyInfo{IsLogin: true,
		DailyDate: date}
}

func (pa *HmtPlayerActivityInfo) AddDungeonJy(nowt int64, num int) {
	date := formatDate(nowt)
	if tmp, ok := pa.DateActivityInfo[date]; ok {
		if tmp.DungeonJy == nil {
			tmp.DungeonJy = make([]int, 0)
			tmp.DungeonJyTime = make([]int64, 0)
		}
		tmp.DungeonJy = append(tmp.DungeonJy, num)
		tmp.DungeonJyTime = append(tmp.DungeonJyTime, nowt)
	}
	logs.Debug("[HMT Activity] Hmt Activity Date %s:DungeonJy record %v", date, pa.DateActivityInfo[date])
}

func (pa *HmtPlayerActivityInfo) AddDungeonDy(nowt int64, num int) {
	date := formatDate(nowt)
	if tmp, ok := pa.DateActivityInfo[date]; ok {
		if tmp.DungeonDy == nil {
			tmp.DungeonDy = make([]int, 0)
			tmp.DungeonDyTime = make([]int64, 0)
		}
		tmp.DungeonDy = append(tmp.DungeonDy, num)
		tmp.DungeonDyTime = append(tmp.DungeonDyTime, nowt)

	}
	logs.Debug("[HMT Activity] Hmt Activity Date %s:DungeonDy record %v", date, pa.DateActivityInfo[date])
}

func (pa *HmtPlayerActivityInfo) AddGachaNum(nowt int64, num int) {
	date := formatDate(nowt)
	if tmp, ok := pa.DateActivityInfo[date]; ok {
		if tmp.GachaNum == nil {
			tmp.GachaNum = make([]int, 0)
			tmp.GachaNumTime = make([]int64, 0)
		}
		tmp.GachaNum = append(tmp.GachaNum, num)
		tmp.GachaNumTime = append(tmp.GachaNumTime, nowt)
	}
	logs.Debug("[HMT Activity] Hmt Activity Date %s:GachaNum record %v", date, pa.DateActivityInfo[date])
}

func formatDate(nowt int64) string {
	tm := time.Unix(nowt, 0).In(util.ServerTimeLocal)
	year, month, day := tm.Date()
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}

func (pa *HmtPlayerActivityInfo) ToDB() HmtPlayerActivityInfoInDB {
	db := HmtPlayerActivityInfoInDB{
		DateActivityInfo: make(hmtActivityInfo_slice, 0, len(pa.DateActivityInfo)),
	}
	for _, v := range pa.DateActivityInfo {
		db.DateActivityInfo = append(db.DateActivityInfo, *v)
	}
	sort.Sort(db.DateActivityInfo)
	return db
}

func (pa *HmtPlayerActivityInfo) FromDB(data *HmtPlayerActivityInfoInDB) error {
	pa.DateActivityInfo = make(map[string]*DailyInfo, len(data.DateActivityInfo))
	for _, v := range data.DateActivityInfo {
		pa.DateActivityInfo[v.DailyDate] = &v
	}
	return nil
}

type hmtActivityInfo_slice []DailyInfo

func (pa hmtActivityInfo_slice) Len() int {
	return len(pa)
}

func (pa hmtActivityInfo_slice) Less(i, j int) bool {
	return pa[i].DailyDate < pa[j].DailyDate
}

func (pa hmtActivityInfo_slice) Swap(i, j int) {
	pa[i], pa[j] = pa[j], pa[i]
}
