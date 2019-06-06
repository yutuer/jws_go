package db

import (
	"fmt"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/auth/models"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/config"
)

const (
	YYBGiftTable = "YYBGift"
	MongoDB      = "MongoDB"
	DynamoDB     = "DynamoDB"
)

// not support nested
type YYBGift struct {
	SdkUid  string `bson:"id,omitempty" json:"id"`
	GiftID  int    `bson:"gift_id,omitempty" json:"gift_id"`
	Times   int    `bson:"times,omitempty" json:"times"`
	GetTime int64  `bson:"get_time,omitempty" json:"get_time"`
}

var (
	GiftDB DB
	MailDB timail.Timail
	AuthDB models.DBInterface
)

type DB interface {
	Init(cfg config.Config) error
	checkSignCondition(sdkUid string, giftId uint32) (int, bool)
	checkLoginCondition(sdkUid string, acId string, giftId uint32) (int, bool)
	checkLevelCondition(sdkUid string, acId string, giftId, cond uint32) (int, bool)
	checkGSCondition(sdkUid string, acId string, giftId, cond uint32) (int, bool)
}

func Init() {
	cfg := config.CommonConfig
	switch cfg.DB {
	case DynamoDB:
		GiftDB = &DBByDynamoDB{}
		AuthDB = &models.DBByDynamoDB{}
	case MongoDB:
		GiftDB = &DBByMongoDB{}
		AuthDB = &models.DBByMongoDB{}
	}
	err := etcd.InitClient(cfg.EtcdEndPoint)
	if err != nil {
		panic(fmt.Sprintf("etcd InitClient err %s", err.Error()))
	}
	logs.Info("etcd Init client done.")

	if err := GiftDB.Init(cfg); err != nil {
		panic(fmt.Sprintf("GiftDB init err %v", err))
		return
	}

	_timail, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
		DBName:       cfg.MailDBName,
		MongoDBUrl:   cfg.MailMongoDBUrl,
		AWSRegion:    cfg.DynamoRegion,
		AWSAccessKey: cfg.DynamoAccessKey,
		AWSSecretKey: cfg.DynamoSecretAccess,
		DBDriver:     cfg.DB,
	})
	if err != nil {
		panic(fmt.Sprintf("MailDB init err %v", err))
	}
	MailDB = _timail

	if err := AuthDB.Init(models.DBConfig{
		MongoDBUrl:            cfg.AuthMongoDBUrl,
		MongoDBName:           cfg.AuthDBName,
		DynamoRegion:          cfg.DynamoRegion,
		DynamoAccessKeyID:     cfg.DynamoAccessKey,
		DynamoSecretAccessKey: cfg.DynamoSecretAccess,
		DynamoSessionToken:    cfg.DynamoSessionToken,
	}); err != nil {
		panic(fmt.Sprintf("AuthDB init err %v", err))
	}
}

func CheckSend(sdkUid string, giftId uint32) int {
	cfg, ok := config.GetGiftByID(giftId)
	if !ok {
		return config.RetCode_SendFailed
	}
	// 找accountid
	info, err, _ := AuthDB.GetDeviceInfo(sdkUid, true)
	if err != nil {
		logs.Error("DBByMongoDB CheckSend GetDeviceInfo err %v", err)
		return config.RetCode_NoRole
	}
	lastGidSid, err := AuthDB.GetLastLoginShard(info.UserId.String())
	if err != nil || lastGidSid == "" {
		logs.Error("DBByMongoDB CheckSend GetLastLoginShard err %v", err)
		return config.RetCode_NoRole
	}
	acid := fmt.Sprintf("%s:%s", lastGidSid, info.UserId)
	var ret int = config.RetCode_InvalidRequest
	var condOK bool
	switch cfg.Type {
	case config.Sign_Typ:
		ret, condOK = GiftDB.checkSignCondition(sdkUid, giftId)
	case config.Login_Typ:
		ret, condOK = GiftDB.checkLoginCondition(sdkUid, acid, giftId)
	case config.Level_Typ:
		ret, condOK = GiftDB.checkLevelCondition(sdkUid, acid, giftId, cfg.CondVal)
	case config.GS_Type:
		ret, condOK = GiftDB.checkGSCondition(sdkUid, acid, giftId, cfg.CondVal)
	}

	if !condOK {
		return ret
	}

	// send mail
	time_now := time.Now().Unix()
	ItemId := make([]string, len(cfg.GiftData))
	Count := make([]uint32, len(cfg.GiftData))
	for i, r := range cfg.GiftData {
		ItemId[i] = r.GetItemID()
		Count[i] = r.GetItemCount()
	}
	logs.Info("Info: acid: %s", acid)
	if err := MailDB.SendMail(fmt.Sprintf("profile:%s", acid), timail.MailReward{
		Param:     []string{},
		Idx:       timail.MkMailId(timail.Mail_Send_By_Sys, 0),
		TimeBegin: time_now,
		TimeEnd:   time_now + int64(util.DaySec*7),
		IdsID:     mail_sender.IDS_MAIL_ACTIVITY_YYB_GIFT_TITLE,
		ItemId:    ItemId,
		Count:     Count,
		Reason:    "YYBGift",
	}); err != nil {
		logs.Error("DBByMongoDB CheckSend MailDB.SendMail err %v", err)
		return config.RetCode_SendFailed
	}

	return config.RetCode_FinishAndSuccess
}
