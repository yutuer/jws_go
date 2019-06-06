package gamedata

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type hotGvgData struct {
	hotGvg *ProtobufGen.GVGHOTDATA
}

type hotGvgMng struct {
}

func (sg *hotGvgMng) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.GVGHOTDATA_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	items := dataList.GetItems()
	if items == nil || items[0] == nil {
		return fmt.Errorf("gvg nil data")
	}
	datas.GvgConfig.hotGvg = items[0]
	logs.Debug("Load Hot Data hotGvg Success， notice time week: %d, time: %d",
		datas.GvgConfig.hotGvg.GetReportWeek(), datas.GvgConfig.hotGvg.GetAnnounceTime())
	return nil
}

func (gvg hotGvgData) GetHotGvgConfig() *ProtobufGen.GVGHOTDATA {
	return gvg.hotGvg
}

type GVGInfo2Client struct {
	GVGNoticeTime     int64 `codec:"nt"`
	GVGResetTime      int64 `codec:"rt"`
	GVGOpeningTime    int64 `codec:"ot"`
	GVGEndTime        int64 `codec:"st"`
	GVGBalanceEndTime int64 `codec:"ste"`
}

func (gi *GVGInfo2Client) GetNextResetTitleTime(now_t int64) int64 {
	if now_t > gi.GVGResetTime {
		return gi.GVGResetTime + util.WeekSec
	} else {
		return gi.GVGResetTime
	}
}

func (gvg hotGvgData) GetGVGEndTime(sid uint, now_t int64) int64 {
	ts, _ := gvg.GetGVGTime(sid, now_t)
	return ts.GVGEndTime
}

func (gvg hotGvgData) GetGVGTime(sid uint, now_t int64) (GVGInfo2Client, int64) {
	var res GVGInfo2Client
	st := game.ServerStartTime(sid)
	_sus_t := st + int64(gvg.hotGvg.GetSuspensionHour())*Hour2Second
	if now_t < _sus_t {
		now_t = _sus_t
	}
	_ts := util.GetWeekTime(now_t,
		int(gvg.hotGvg.GetRestartAndStartWeek()),
		gvg.hotGvg.GetStartTime())
	end_ts := _ts + int64(gvg.hotGvg.GetGVGOpeningTime()+
		gvg.hotGvg.GetGVGStatementTime())*Minute2Second
	var offset int64
	if end_ts <= now_t {
		offset = util.WeekSec
	}
	res.GVGBalanceEndTime = end_ts + offset
	res.GVGEndTime = _ts +
		int64(gvg.hotGvg.GetGVGOpeningTime())*Minute2Second + offset
	res.GVGOpeningTime = _ts + offset
	res.GVGResetTime = util.GetWeekTime(now_t,
		int(gvg.hotGvg.GetRestartAndStartWeek()),
		gvg.hotGvg.GetRestartCityTime()) + offset
	res.GVGNoticeTime = util.GetWeekTime(now_t,
		int(gvg.hotGvg.GetReportWeek()),
		gvg.hotGvg.GetAnnounceTime()) + offset
	return res, end_ts
}
