package rank

import (
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func SendMail(shardId uint, uid string, rank_id int64, rank, class int, ids uint32, param []string, reason string,
	item_id []string, count []uint32, mail_type int64) bool {

	time_now := game.GetNowTimeByOpenServer(shardId)

	mail := timail.MailReward{
		IdsID:     ids,
		Param:     param,
		TimeBegin: time_now,
		Reason:    reason,
		TimeEnd:   time_now + util.WeekSec, // TODO 需要策划确认邮件的有效期和内容 TBDYZH
	}
	for idx, itemid := range item_id {
		mail.AddReward(itemid, count[idx])
	}
	// 防止一个人在两张榜中同时有奖励邮件
	// 防止一个人在一张榜中同时有两个奖励邮件
	mail.Idx = timail.MkMailId(mail_type, rank_id*int64(10)+int64(class))

	mail_sender.AddMailBatchSend(shardId, uid, mail, false, mail_type)

	return true
}
