package account

import (
	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
)

type Account7Day struct {
	Goods            map[uint32]Good7Day // PromotionID->Good7Day
	NextRefDailyTime int64
}

type Good7Day struct {
	PromotionID uint32 `json:"pid"`
	LeftTimes   uint32 `json:"lt"`
}

type Account7DayInDB struct {
	Goods            good7day_slice
	NextRefDailyTime int64
}

func (a7d *Account7Day) onAfterLogin() {
	lc := gamedata.GetAccount7DayLimitCount()
	if a7d.Goods == nil || len(a7d.Goods) <= 0 {
		a7d.Goods = make(map[uint32]Good7Day, len(lc))
		for id, cfg := range lc {
			a7d.Goods[id] = Good7Day{
				PromotionID: id,
				LeftTimes:   cfg.GetCountLimit(),
			}
		}
	} else if len(a7d.Goods) < len(lc) {
		ng := make(map[uint32]Good7Day, len(lc))
		for id, cfg := range lc {
			ov, ok := a7d.Goods[id]
			if ok {
				ng[id] = ov
			} else {
				ng[id] = Good7Day{
					PromotionID: id,
					LeftTimes:   cfg.GetCountLimit(),
				}
			}
			a7d.Goods = ng
		}
	}
}
func (a7d *Account7Day) UpdateGoods(now_time int64) {
	if now_time < a7d.NextRefDailyTime {
		return
	}
	a7d.NextRefDailyTime = util.GetNextDailyTime(
		gamedata.GetCommonDayBeginSec(now_time), now_time)
	for id, g := range a7d.Goods {
		cfg := gamedata.GetAccount7DayGood(id)
		if cfg.GetLimitType() == gamedata.LimitCountRefTyp_Daily {
			g.LeftTimes = cfg.GetCountLimit()
		}
	}
}

func (a7d *Account7Day) ToDB() Account7DayInDB {
	db := Account7DayInDB{
		Goods:            make(good7day_slice, 0, len(a7d.Goods)),
		NextRefDailyTime: a7d.NextRefDailyTime,
	}
	for _, v := range a7d.Goods {
		db.Goods = append(db.Goods, v)
	}
	sort.Sort(db.Goods)
	return db
}

func (a7d *Account7Day) FromDB(data *Account7DayInDB) error {
	a7d.NextRefDailyTime = data.NextRefDailyTime
	a7d.Goods = make(map[uint32]Good7Day, len(data.Goods))
	for _, v := range data.Goods {
		a7d.Goods[v.PromotionID] = v
	}
	return nil
}

type good7day_slice []Good7Day

func (a7d good7day_slice) Len() int {
	return len(a7d)
}

func (a7d good7day_slice) Less(i, j int) bool {
	return a7d[i].PromotionID < a7d[j].PromotionID
}

func (a7d good7day_slice) Swap(i, j int) {
	a7d[i], a7d[j] = a7d[j], a7d[i]
}
