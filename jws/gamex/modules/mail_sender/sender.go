package mail_sender

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

var mail_db timail.Timail

func initMail() error {
	db, err := mailhelper.NewMailDriver(cfg)
	if err == nil {
		mail_db = db
	}
	return err
}

func sendMailImp(now_t, addon int64, data *MailToUser) bool {
	data.Mail.Idx = timail.MkMailIdByTime(now_t, data.Typ, addon) // 防止一个人同时有奖励邮件

	err := mail_db.SyncMail("profile:"+data.Uid, []timail.MailReward{data.Mail}, []int64{})
	if err != nil {
		logs.SentryLogicCritical(data.Uid, "SendMailErr %v %s", *data, err.Error())
		return false
	}
	// logic log
	item := data.Mail
	logiclog.LogSendMail(data.Uid, item.IdsID, item.Reason, item.ItemId, item.Count)
	return true
}

//func MailExist(user_id string, idx int64) bool {
//	return mail_db.MailExist(user_id, idx)
//}
