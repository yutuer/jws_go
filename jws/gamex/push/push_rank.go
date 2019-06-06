package push

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
)

type ch_rank_info chan rank_info

var (
	ch_boss_total  ch_rank_info
	ch_boss_singel ch_rank_info
	ch_simple_pvp  ch_rank_info
	timing         *timingwheel.TimingWheel
)

func init() {
	ch_boss_total = make(chan rank_info, 200)
	ch_boss_singel = make(chan rank_info, 200)
	ch_simple_pvp = make(chan rank_info, 1024)
	_push(ch_boss_total, bossTotalPushContent)
	_push(ch_boss_singel, bossSinglePushContent)
	_push(ch_simple_pvp, simplePvpPushContent)
}

const (
	boss_rank_limit       = 30
	Simple_pvp_rank_limit = 100

//	boss_rank_limit       = 2 // 临时修改 TDB by zhangzhen
//	simple_pvp_rank_limit = 2 // 临时修改
)

type rank_info struct {
	acid        string
	platformid  string
	devicetoken string
	rank        int
	param       []string
}

func BossFightTotal(rank int, acid string) {
	if rank <= boss_rank_limit {
		ch_boss_total <- rank_info{
			acid: acid,
			rank: rank,
		}
	}
}

func bossTotalPushContent(rank int, param ...string) string {
	if rank == 1 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_HFEATS_1")
	} else if rank == 2 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_HFEATS_2")
	} else if rank == 3 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_HFEATS_3")
	} else if rank <= 10 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_HFEATS_10")
	} else if rank <= 30 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_HFEATS_30")
	}
	return ""
}

func BossFightSingle(rank int, acid string, params ...string) {
	if rank <= boss_rank_limit {
		ch_boss_singel <- rank_info{
			acid:  acid,
			rank:  rank,
			param: params,
		}
	}
}

func bossSinglePushContent(rank int, param ...string) string {
	if rank == 1 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_ACCFEATS_1")
	} else if rank == 2 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_ACCFEATS_2")
	} else if rank == 3 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_ACCFEATS_3")
	} else if rank <= 10 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_ACCFEATS_10")
	} else if rank <= 30 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_BOSSFIGHT_ACCFEATS_30")
	}
	return ""
}

func SimplePvp(rank int, acid, platformid, devicetoken string, params ...string) {
	if uutil.IsOverseaVer() {
		return
	}
	if rank <= Simple_pvp_rank_limit {
		ch_simple_pvp <- rank_info{
			acid:        acid,
			platformid:  platformid,
			devicetoken: devicetoken,
			rank:        rank,
			param:       params,
		}
	}
}

func simplePvpPushContent(rank int, param ...string) string {
	if rank == 1 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_PVP_1", param...)
	} else if rank == 2 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_PVP_2", param...)
	} else if rank == 3 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_PVP_3", param...)
	} else if rank <= 10 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_PVP_10", param...)
	} else if rank <= 30 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_PVP_30", param...)
	} else if rank <= 100 {
		return gamedata.GetCommonIdsStr("IDS_PUSH_PVP_100", param...)
	}
	return ""
}

type fPushContent func(rank int, param ...string) string

func _push(ch ch_rank_info, f fPushContent) {
	acids := make([]string, 0, 100)
	acid2rank := make(map[string]*rank_info, 100)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logs.Error("[_push] recover error %v", err)
			}
		}()
		for {
			timerChan := uutil.TimerSec.After(time.Second)
			select {
			case info := <-ch:
				acid2rank[info.acid] = &info
				acids = append(acids, info.acid)
			case <-timerChan:
				if len(acids) <= 0 {
					continue
				}
				_acids := acids
				_acid2rank := acid2rank
				acids = make([]string, 0, 100)
				acid2rank = make(map[string]*rank_info, 100)
				// 查device token，在push
				//acid2device, err := account_info.AccountInfoDynamo.GetAccountDeviceInfos(_acids)
				//if err != nil {
				//	logs.Error("rank push err %s", err.Error())
				//	continue
				//}
				i := 0
				acid := ""
				for {
					if i < len(_acids) {
						acid = _acids[i]
					} else {
						break
					}
					rank, ok := _acid2rank[acid]
					if !ok {
						i++
						continue
					}
					content := f(rank.rank, rank.param...)
					if content == "" {
						i++
						continue
					}
					logs.Trace("[push_rank] %s %s", acid, content)
					//device, ok := acid2device[acid]
					//if !ok {
					//	continue
					//}
					retry := Push2Device(rank.platformid, rank.devicetoken, "", content)
					if retry {
						_acids = append(_acids, acid)
					}
					i++

					time.Sleep(10 * time.Millisecond)
				}

				timerChan = uutil.TimerSec.After(time.Second)
			}
		}
	}()
}
