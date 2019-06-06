package rules

import (
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/youmi_imcc/base"
)

// 将20s之内在世界频道连续发言的人禁言
func queryInterval(infos *base.MsgInfo) (ret []*base.MuteInfo) {
	ret = make([]*base.MuteInfo, 0)
	for k, v := range infos.MsgInfoMap {
		msgValue := base.ParseMsgValue(v.Value)
		if len(msgValue) < 2 {
			logs.Debug("first talk")
			return nil
		}
		index := len(msgValue) - 1
		if isLegal(msgValue[index].CreateTime, msgValue[index-1].CreateTime) {
			return nil
		}
		logs.Info("mute user: %v", k)
		ret = append(ret, base.GenMuteInfo(k, "世界频道发言时间间隔过短", -1, "", false))
	}
	return ret
}

func isLegal(s1, s2 string) bool {
	time1, err := time.Parse("2006-01-02 15:04:05", s1)
	if err != nil {
		logs.Error("arg format error for str: %s by err %v", s1, err)
		return true
	}
	time2, err := time.Parse("2006-01-02 15:04:05", s2)
	if err != nil {
		logs.Error("arg format error for str: %s by err %v", s2, err)
		return true
	}
	i1 := time1.Unix()
	i2 := time2.Unix()
	return i1-i2 >= base.Cfg.LegalInterval
}
