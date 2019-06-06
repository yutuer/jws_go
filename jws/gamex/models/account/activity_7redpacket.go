package account

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

const SevenDays = 7

type RedPacket7Days struct {
	CreateTime int64
	SaveHc     [SevenDays]int64
	AllHc      int64
}

func (re *RedPacket7Days) GetDay(nTime int64) int64 {
	beginDay := util.DailyBeginUnixByStartTime(re.CreateTime, gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	today := util.DailyBeginUnixByStartTime(nTime, gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	return (today - beginDay) / int64(util.DaySec)
}

func (re *RedPacket7Days) GetRelativeTime(nTime int64) int64 {
	beginDay := util.DailyBeginUnixByStartTime(re.CreateTime, gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	endDay := beginDay + int64(gamedata.GetCommonCfg().GetReceiveDay())*int64(util.DaySec)
	return endDay - nTime
}

func (re *RedPacket7Days) SetCreatTime(nTime int64) {
	if re.CreateTime == 0 {
		re.CreateTime = nTime
	}

}

func (re *RedPacket7Days) UpdateSaveHc(nTime, saveHc int64) {
	day := re.GetDay(nTime)
	if day < 0 {
		return
	}
	if day < int64(gamedata.GetCommonCfg().GetReserveDays()) {
		re.SaveHc[int(day)] += saveHc
	}
}

func (re *RedPacket7Days) GetInfo2Client() []int64 {
	redPacket2cliet := make([]int64, len(re.SaveHc))
	for x, hc := range re.SaveHc {
		if hc == -1 {
			redPacket2cliet[x] = -1
		} else {
			redPacket2cliet[x] = hc / int64(gamedata.GetCommonCfg().GetRedpackeTratio())
		}
	}
	return redPacket2cliet
}

func (re *RedPacket7Days) GetTotalHc() int64 {
	var hcNum int64
	for _, x := range re.SaveHc {
		hcNum += x / int64(gamedata.GetCommonCfg().GetRedpackeTratio())
	}
	if hcNum > re.AllHc {
		re.AllHc = hcNum
	}
	return re.AllHc
}

func (re *RedPacket7Days) GetDayHcNum(day int) uint32 {
	return uint32(re.SaveHc[day])
}

func (re *RedPacket7Days) SetPacket2Done(day int) {
	re.SaveHc[day] = -1
}

func (re *RedPacket7Days) SendRedPacket7daysMail(accountId string, nTime int64) {
	if re.GetDay(nTime)+1 > int64(gamedata.GetCommonCfg().GetReceiveDay()) {
		for i, num := range re.SaveHc {
			if num > 0 {
				mail_sender.BatchSendMail2Account(accountId,
					timail.Mail_send_By_RedPacket7Days,
					mail_sender.IDS_MAIL_REPACKT_TITLE,
					[]string{fmt.Sprintf("%d", i+1)},
					map[string]uint32{
						gamedata.VI_Hc: uint32(num) / gamedata.GetCommonCfg().GetRedpackeTratio(),
					},
					"RedPacket 7 days", false)
				re.SetPacket2Done(i)
			}
		}
	}
}
