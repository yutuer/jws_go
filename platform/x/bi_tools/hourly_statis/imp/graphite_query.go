package imp

import (
	"fmt"

	"time"

	"math"

	"strconv"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func collectAPCU() error {
	for channel, sidData := range channel2Sid2Data {
		for sid, _ := range sidData {
			data := channel2Sid2Data[channel][sid]
			gid := _sid2gid(sid)
			if gid < 0 {
				logs.Error("collectAPCU _sid2gid gid is nil, sid=%s", sid)
				continue
			}
			// acu
			ccu_k := fmt.Sprintf("gid%d.sid%s.gamex.gamex.%s.ccu.count", gid, sid, channel)
			acu_f := fmt.Sprintf("summarize(%s, '1h', 'avg', false)", ccu_k)
			acu_url := fmt.Sprintf("%s/render?target=%s&from=%s&format=json",
				Cfg.Graphite, acu_f, data_time_begin_graphite)

			v, err := _graphite(acu_url, channel, sid)
			if err != nil {
				continue
			}
			data.acu = int(math.Ceil(v))

			// pcu
			pcu_f := fmt.Sprintf("summarize(%s, '1h', 'max', false)", ccu_k)
			pcu_url := fmt.Sprintf("%s/render?target=%s&from=%s&format=json",
				Cfg.Graphite, pcu_f, data_time_begin_graphite)

			v, err = _graphite(pcu_url, channel, sid)
			if err != nil {
				continue
			}
			data.pcu = int(v)
		}
	}
	logs.Info("collectAPCU %v", channel2Sid2Data)
	return nil
}

func _graphite(url, channel, sid string) (float64, error) {
	logs.Info("graphite url %s", url)
	req := httplib.Get(url).SetTimeout(10*time.Second, 10*time.Second)

	res := make([]graphite_res, 0, 10)
	err := req.ToJson(&res)
	if err != nil {
		logs.Error("collectAPCU req.ToJson err %s", err.Error())
		return 0.0, err
	}

	if len(res) > 0 {
		for _, c := range res[0].Datapoints {
			if len(c) >= 2 {
				_v := c[0]
				v, ok := _v.(float64)
				if !ok {
					continue
				}
				_st := c[1]
				st, ok := _st.(float64)
				if !ok {
					continue
				}
				if int64(st) == begin_st {
					logs.Debug("_graphite got data %f", v)
					if v < 0 {
						return 0.0, nil
					}
					return v, nil
				}
			}
		}
	}
	return 0.0, nil
}

func _sid2gid(ssid string) int {
	sid, err := strconv.Atoi(ssid)
	if err != nil {
		logs.Error("_sid2gid sid strconv.Atoi err %v ", err)
		return -1
	}
	for gid, sids := range Cfg.Gid2Shard {
		if sid >= sids.From && sid <= sids.To {
			res, err := strconv.Atoi(gid)
			if err != nil {
				logs.Error("_sid2gid gid strconv.Atoi err %v ", err)
				return -1
			}
			return res
		}
	}

	return -1
}
