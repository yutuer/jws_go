package logiclog

import (
	"time"

	"vcs.taiyouxi.net/platform/planx/util"
)

type DailyStatistic struct {
	TS        int64    `json:"ts"`
	JoinCount int      `json:"j_c"`
	JoinMem   []string `json:"j_ms"`
}

type DailyStatistics struct {
	Infos []DailyStatistic `json:"infos"`
}

func (ds *DailyStatistics) JoinStic(acid string) {
	tb := util.DailyBeginUnix(time.Now().Unix())
	if ds.Infos == nil {
		ds.Infos = make([]DailyStatistic, 0, 3)
	}
	found := false
	for i, info := range ds.Infos {
		if info.TS == tb {
			_info := &ds.Infos[i]
			_info.JoinCount++
			had := false
			for _, id := range _info.JoinMem {
				if id == acid {
					had = true
					break
				}
			}
			if !had {
				_info.JoinMem = append(_info.JoinMem, acid)
			}
			found = true
			break
		}
	}
	if !found {
		_info := DailyStatistic{
			TS:        tb,
			JoinCount: 1,
			JoinMem:   make([]string, 0, 8),
		}
		_info.JoinMem = append(_info.JoinMem, acid)
		ds.Infos = append(ds.Infos, _info)
	}
	if len(ds.Infos) > 2 {
		ds.Infos = ds.Infos[1:]
	}
}
