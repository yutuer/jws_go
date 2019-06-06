package mail

import (
	//"encoding/json"
	"errors"
	//"time"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	gmTimeUtil "vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/planx/util/tipay/dynamopay"
	"vcs.taiyouxi.net/platform/x/gm_tools/util"
)

type privateMailReword struct {
	IdsID uint32
	Param []string

	ItemId []string
	Count  []uint32

	TimeBegin          int64
	TimeBeginString    string
	TimeEnd            int64
	TimeEndString      string
	CreateBefore       int64
	CreateBeforeString string
	CreateEnd          int64
	CreateEndString    string

	Reason string
	Tag    string

	IsGetted bool
	IsRead   bool

	Idx   int64
	state int

	Account2Send []string
}

func GetAllMail(server_name, mail_profile_name string) (string, error) {
	cfg := CommonCfg.GetServerCfgFromName(server_name)

	if cfg.RedisName == "" {
		logs.Error("GetAllMail Err No Name %v", server_name)
		return "", errors.New("NoServerName")
	}
	m, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
		AWSRegion:    CommonCfg.AWS_Region,
		DBName:       cfg.MailDBName,
		AWSAccessKey: CommonCfg.AWS_AccessKey,
		AWSSecretKey: CommonCfg.AWS_SecretKey,

		MongoDBUrl: cfg.MailMongoUrl,
		DBDriver:   cfg.MailDBDriver,
	})

	if err != nil {
		return "", err
	}

	res, err := m.LoadAllMailByGM(mail_profile_name)
	if err != nil {
		return "", err
	}
	var help_res []privateMailReword = make([]privateMailReword, 0, len(res))
	for _, v := range res {
		arg := privateMailReword{
			v.IdsID,
			v.Param,
			v.ItemId,
			v.Count,
			v.TimeBegin,
			time.Unix(v.TimeBegin, 0).In(gmTimeUtil.ServerTimeLocal).Format("2006/01/02 15:04:05"),
			v.TimeEnd,
			time.Unix(v.TimeEnd, 0).In(gmTimeUtil.ServerTimeLocal).Format("2006/01/02 15:04:05"),
			v.CreateBefore,
			time.Unix(v.CreateBefore, 0).In(gmTimeUtil.ServerTimeLocal).Format("2006/01/02 15:04:05"),
			v.CreateEnd,
			time.Unix(v.CreateEnd, 0).In(gmTimeUtil.ServerTimeLocal).Format("2006/01/02 15:04:05"),
			v.Reason,
			v.Tag,
			v.IsGetted,
			v.IsRead,
			v.Idx,
			v.GetState(),
			v.Account2Send,
		}
		if arg.CreateBefore == 0 {
			arg.CreateBeforeString = ""
		}
		if arg.CreateEnd == 0 {
			arg.CreateEndString = ""
		}
		help_res = append(help_res, arg)
	}
	str, err := util.ToJson(help_res)

	if err != nil {
		return "", err
	}

	logs.Trace("GetAllMail Res %v", help_res)

	return string(str), nil
}

func SendMail(server_name, mail_profile_name string, mail_data *timail.MailReward) error {
	return SendMailWithAddon(server_name, mail_profile_name, mail_data, 0)
}

func SendMailWithAddon(server_name, mail_profile_name string, mail_data *timail.MailReward, addon int) error {

	cfg := CommonCfg.GetServerCfgFromName(server_name)

	if cfg.RedisName == "" ||
		CommonCfg.AWS_Region == "" ||
		cfg.MailDBName == "" {
		logs.Error("SendMail Err No Name %v %s|%s|%s|%s", server_name,
			CommonCfg.AWS_Region,
			cfg.MailDBName,
			CommonCfg.AWS_AccessKey,
			CommonCfg.AWS_SecretKey)
		return errors.New("NoServerName")
	}

	m, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
		AWSRegion:    CommonCfg.AWS_Region,
		DBName:       cfg.MailDBName,
		AWSAccessKey: CommonCfg.AWS_AccessKey,
		AWSSecretKey: CommonCfg.AWS_SecretKey,
		MongoDBUrl:   cfg.MailMongoUrl,
		DBDriver:     cfg.MailDBDriver,
	})

	if err != nil {
		logs.Error("DelMail Error:%v", err)
		return err
	}

	mail_data.Idx = timail.MkMailId(timail.Mail_Send_By_Sys, int64(addon%1000))

	logs.Debug("sendmail %s %s", server_name, mail_profile_name)
	return m.SyncMail(mail_profile_name, []timail.MailReward{*mail_data}, []int64{})
}

func SetVirTrueIapPayAndroid(value map[string]interface{}) {
	dbPayMgr := dynamopay.NewPayDynamoDB(
		CommonCfg.PayAndroidDBName,
		"cn-north-1",
		CommonCfg.AWS_AccessKey,
		CommonCfg.AWS_SecretKey)

	err := dbPayMgr.Open()
	if err != nil {
		logs.Error("Open  err %v", err)
		return
	}

	err = dbPayMgr.SetByHashM_IfNoExist(value["order_no"], value)
	if err != nil {
		logs.Error("SetByHashM 1 err %v", err)
		return
	}
	return

}

func SetVirTrueIapPayIOS(value map[string]interface{}) {
	dbPayMgr := dynamopay.NewPayDynamoDB(
		CommonCfg.PayIOSDBName,
		"cn-north-1",
		CommonCfg.AWS_AccessKey,
		CommonCfg.AWS_SecretKey)

	err := dbPayMgr.Open()
	if err != nil {
		logs.Error("Open  err %v", err)
		return
	}

	err = dbPayMgr.SetByHashM_IfNoExist(value["order_no"], value)
	if err != nil {
		logs.Error("SetByHashM 1 err %v", err)
		return
	}
	return

}

func SendMailVirtualIap(server_name string, data *Mail2User, addon int) error {
	cfg := CommonCfg.GetServerCfgFromName(server_name)

	if cfg.RedisName == "" ||
		CommonCfg.AWS_Region == "" ||
		cfg.MailDBName == "" {
		logs.Error("SendMail Err No Name %v %s|%s|%s|%s", server_name,
			CommonCfg.AWS_Region,
			cfg.MailDBName,
			CommonCfg.AWS_AccessKey,
			CommonCfg.AWS_SecretKey)
		return errors.New("NoServerName")
	}

	m, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
		AWSRegion:    CommonCfg.AWS_Region,
		DBName:       cfg.MailDBName,
		AWSAccessKey: CommonCfg.AWS_AccessKey,
		AWSSecretKey: CommonCfg.AWS_SecretKey,
		MongoDBUrl:   cfg.MailMongoUrl,
		DBDriver:     cfg.MailDBDriver,
	})

	if err != nil {
		return err
	}
	mailIdx := timail.MkMailIdByTime(time.Now().Unix(), data.Typ, int64(addon%1000))
	data.Mail.Idx = mailIdx
	err = m.SendMail("profile:"+data.Uid, data.Mail)

	if err != nil {
		logs.SentryLogicCritical(data.Uid, "SendMailErr %v", *data)
		return err
	}

	logs.Debug("maosi succes")
	return nil
}

func DelMail(server_name, mail_profile_name string, id_to_del int64) error {
	cfg := CommonCfg.GetServerCfgFromName(server_name)

	if cfg.RedisName == "" {
		logs.Error("SendMail Err No Name %v", server_name)
		return errors.New("NoServerName")
	}

	m, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
		AWSRegion:    CommonCfg.AWS_Region,
		DBName:       cfg.MailDBName,
		AWSAccessKey: CommonCfg.AWS_AccessKey,
		AWSSecretKey: CommonCfg.AWS_SecretKey,
		MongoDBUrl:   cfg.MailMongoUrl,
		DBDriver:     cfg.MailDBDriver,
	})

	if err != nil {
		return err
	}

	return m.SyncMail(mail_profile_name, []timail.MailReward{}, []int64{id_to_del})
}
