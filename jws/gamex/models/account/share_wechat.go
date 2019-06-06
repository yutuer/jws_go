package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type ShareWeChatInfo struct {
	Items     []ShareWeChatItem `json:"s_wc_i"`
	ResetTime int64             `json:"s_wc_t"`
}
type ShareWeChatItem struct {
	Type  int `json:"s_wc_type" codec:"s_wc_type"`
	Times int `json:"s_wc_times" codec:"s_wc_times"`
}

func (swi *ShareWeChatItem) addCount() {
	swi.Times += 1
}

func (sw *ShareWeChatInfo) Init() {
	itemMap := gamedata.GetShareWeChatData()
	sw.Items = make([]ShareWeChatItem, 0, len(itemMap))
	for k, _ := range itemMap {
		sw.Items = append(sw.Items, ShareWeChatItem{Type: int(k.Type), Times: 0})
	}
	logs.Debug("Init sharewechatInfo")
}

func (sw *ShareWeChatInfo) GetItems() []ShareWeChatItem {
	return sw.Items
}

func (sw *ShareWeChatInfo) AddTimesByType(t int) {
	if len(sw.Items) == 0 {
		sw.Init()
	}
	for i := 0; i < len(sw.Items); i++ {
		item := &sw.Items[i]
		if item.Type == t {
			item.addCount()
			return
		}
	}

	logs.Error("No type: %d WeChat rewards", t)
}

func (sw *ShareWeChatInfo) GetTimesByType(t int) (int, bool) {
	if len(sw.Items) == 0 {
		sw.Init()
	}
	for i := 0; i < len(sw.Items); i++ {
		item := sw.Items[i]
		if item.Type == t {
			return item.Times, true
		}
	}
	logs.Error("No type: %d WeChat Type", t)
	return 0, false
}

func (sw *ShareWeChatInfo) UpdateTimesAndRest(now_t int64) {
	if now_t < sw.ResetTime {
		return
	}
	sw.ResetTime = util.DailyBeginUnixByStartTime(now_t,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	sw.ResetTime += util.DaySec
	sw.Init()
}
