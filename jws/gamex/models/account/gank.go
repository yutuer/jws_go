package account

import (
	"time"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/message"
	"vcs.taiyouxi.net/platform/planx/util"
)

const (
	GankMsgTableKey = "gank"
	GankMsgCount    = 5
)

// 切磋
type PlayerGank struct {
	GankWinSum          int64 `json:"gank_sum"`
	GankTodayWin        int64 `json:"gank_day_win"`
	GankTodayTime       int64 `json:"gank_day_t"`
	GankDayWinMax       int64 `json:"gank_win_max"`
	GankLastReviewLogTS int64 `json:"gank_l_revw_log_ts"`
	GankNewestLogTS     int64 `json:"gank_n_log_ts"`
}

type GankRecord struct {
	Time            int64  `json:"t"`
	IDS             int    `json:"ids"`
	FighterRoleId   string `json:"rid"`
	FighterRoleName string `json:"rnm"`
}

func (g *PlayerGank) OnAfterLogin(acid string) {
	err, recs := g.GetMsgLogs(acid)
	if err != nil {
		return
	}
	for _, m := range recs {
		g.SetNewestLogTS(m.Time)
	}
}

func (g *PlayerGank) GetMsgLogs(acid string) (error, []GankRecord) {
	msgs, err := message.LoadPlayerMsgs(acid,
		GankMsgTableKey, GankMsgCount)
	if err != nil {
		return err, nil
	}

	recs := make([]GankRecord, 0, GankMsgCount)
	for _, msg := range msgs {
		m := GankRecord{}
		err := json.Unmarshal([]byte(msg.Params[0]), &m)
		if err != nil {
			continue
		}
		recs = append(recs, m)
	}
	return nil, recs
}

func (g *PlayerGank) WinLog() {
	g.GankWinSum++
	zt := util.DailyBeginUnix(time.Now().Unix())
	if util.GetDayBefore(time.Unix(g.GankTodayTime, 0), time.Unix(zt, 0)) > 0 {
		g.GankTodayWin = 1
		g.GankTodayTime = zt
	} else {
		g.GankTodayWin++
	}
	if g.GankTodayWin > g.GankDayWinMax {
		g.GankDayWinMax = g.GankTodayWin
	}
}

func (g *PlayerGank) GetTodayWinCount() int64 {
	zt := util.DailyBeginUnix(time.Now().Unix())
	if util.GetDayBefore(time.Unix(g.GankTodayTime, 0), time.Unix(zt, 0)) > 0 {
		return 0
	}
	return g.GankTodayWin
}

func (g *PlayerGank) SetLastReviewLogTS(ts int64) {
	if ts > g.GankLastReviewLogTS {
		g.GankLastReviewLogTS = ts
	}
}

func (g *PlayerGank) SetNewestLogTS(ts int64) {
	if ts > g.GankNewestLogTS {
		g.GankNewestLogTS = ts
	}
}

func (g *PlayerGank) IsRedPoint() bool {
	return g.GankNewestLogTS > g.GankLastReviewLogTS
}
