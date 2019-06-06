package pay

import (
	"time"

	"sync"

	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/api_gateway/config"
)

const (
	year_time = 365 * day_time
	day_time  = 24 * 3600
)

//var MailModule MailSenderModule

type MailToUser struct {
	Mail timail.MailReward
	Typ  int64
	Uid  string
}

type MailSenderModule struct {
	shardId uint
	db      timail.Timail
}

func (r *MailSenderModule) SendAndroidIAPMail(accountId, info string) error {
	m := MailToUser{
		Typ: timail.Mail_Send_By_AndroidIAP,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(year_time)
	m.Mail.Tag = info

	return r.sendMail(m)
}

func (r *MailSenderModule) SendIOSIAPMail(accountId, info string) error {
	m := MailToUser{
		Typ: timail.Mail_Send_By_IOSIAP,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(year_time)
	m.Mail.Tag = info
	return r.sendMail(m)
}

func (r *MailSenderModule) sendMail(m MailToUser) error {
	// 支付gateway服务多个shard，所以这里不分shard了
	return sendMailImp(r.db, time.Now().Unix(), _addon_get(), &m)
}

var addon timail.MailAddonCounter
var addon_lock sync.Mutex

func _addon_get() int64 {
	addon_lock.Lock()
	i := addon.Get()
	addon_lock.Unlock()
	return i
}

func (r *MailSenderModule) Start(dc config.DBConfig) error {
	err := r.initMail(dc)
	if err != nil {
		logs.Error("MailSenderModule DB Open Err by %s", err.Error())
		return err
	}
	return nil
}

func (r *MailSenderModule) Stop() {
	r.db.Close()
}

func (r *MailSenderModule) initMail(dc config.DBConfig) error {
	db, err := mailhelper.NewMailDriver(dc.GetMailConfig())
	if err != nil {
		return err
	}
	r.db = db
	return nil
}

func sendMailImp(db timail.Timail, now_t, addon int64, data *MailToUser) error {
	data.Mail.Idx = timail.MkMailIdByTime(now_t, data.Typ, addon) // 防止一个人同时有奖励邮件
	err := db.SendMail("profile:"+data.Uid, data.Mail)
	//err := db.SyncMail("profile:"+data.Uid, []timail.MailReward{data.Mail}, []int64{})
	if err != nil {
		logs.SentryLogicCritical(data.Uid, "SendMailErr err: %v data: %v", err.Error(), *data)
		return err
	}

	return nil
}
