package db

import (
	"fmt"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/codec"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/config"
)

type DBByDynamoDB struct {
	db dynamodb.DynamoDB
}

func (ddb *DBByDynamoDB) Init(cfg config.Config) error {
	logs.Info("DBByDynamoDB Init %v", cfg)
	db := dynamodb.DynamoDB{}
	authConfig.Cfg.Dynamo_NameDevice = cfg.Dynamo_NameDevice
	authConfig.Cfg.Dynamo_NameName = cfg.Dynamo_NameName
	authConfig.Cfg.Dynamo_NameUserInfo = cfg.Dynamo_NameUserInfo
	authConfig.Cfg.Dynamo_GM = cfg.Dynamo_GM

	err := db.Connect(
		cfg.DynamoRegion,
		cfg.DynamoAccessKey,
		cfg.DynamoSecretAccess,
		cfg.DynamoSessionToken,
	)
	if err != nil {
		return err
	}
	logs.Info("connect dynamoDB success")
	err = db.InitTable()
	if err != nil {
		return err
	}
	logs.Info("DB table inited")
	if !db.IsTableHasCreate(YYBGiftTable) {
		return fmt.Errorf("no table named %v in dynamoDB, break!", YYBGiftTable)
	}
	ddb.db = db
	return nil
}

func (ddb *DBByDynamoDB) checkGSCondition(sdkUid string, acId string, giftId, cond uint32) (int, bool) {
	ret, err := ddb.db.Get(YYBGiftTable, sdkUid, giftId)
	if err != nil {
		logs.Error("DBByDynamoDB Find err %v", err)
		return config.RetCode_SendFailed, false
	}
	// no record
	if len(ret) > 0 {
		logs.Error("Already sent the gift")
		return config.RetCode_HadSend, false
	}
	account := loadProfile(acId)
	if account == nil {
		logs.Error("LoadProfile err %v", err)
		return config.RetCode_SendFailed, false
	}
	if account.Profile.GetData().CorpCurrGS < int(cond) {
		logs.Error("Hero GS err")
		return config.RetCode_NoCondition, false
	}
	yybGift := &YYBGift{
		SdkUid:  sdkUid,
		GiftID:  int(giftId),
		Times:   1,
		GetTime: time.Now().Unix(),
	}
	infoMap, err := ddb.struct2map(yybGift)
	if err != nil {
		logs.Debug("convert map err by %v", err)
		return config.RetCode_SendFailed, false
	}
	if err := ddb.db.SetByHashM(YYBGiftTable, sdkUid, infoMap); err != nil {
		logs.Error("Save DB err %v", err)
		return config.RetCode_SendFailed, false
	}
	return 0, true
}

func (ddb *DBByDynamoDB) map2struct(data map[string]interface{}) (*YYBGift, error) {
	giftData := YYBGift{}
	byteData := codec.Encode(data)
	codec.Decode(byteData, &giftData)
	logs.Debug("map2struct: %v", giftData)
	return &giftData, nil
}

func (ddb *DBByDynamoDB) struct2map(yybGift *YYBGift) (map[string]interface{}, error) {
	mapData := make(map[string]interface{})
	byteData := codec.Encode(yybGift)
	codec.Decode(byteData, &mapData)
	logs.Debug("map2struct: %v", mapData)
	return mapData, nil
}

func (ddb *DBByDynamoDB) checkLevelCondition(sdkUid string, acId string, giftId, cond uint32) (int, bool) {
	ret, err := ddb.db.Get(YYBGiftTable, sdkUid, giftId)
	if err != nil {
		logs.Error("DBByDynamoDB Find err %v", err)
		return config.RetCode_SendFailed, false
	}
	// no record
	if len(ret) > 0 {
		logs.Error("Already sent the gift")
		return config.RetCode_HadSend, false
	}
	account := loadProfile(acId)
	if account == nil {
		logs.Error("LoadProfile err %v", err)
		return config.RetCode_SendFailed, false
	}
	if account.Profile.GetCorp().Level < cond {
		logs.Error("Hero Level err")
		return config.RetCode_NoCondition, false
	}
	yybGift := &YYBGift{
		SdkUid:  sdkUid,
		GiftID:  int(giftId),
		Times:   1,
		GetTime: time.Now().Unix(),
	}
	infoMap, err := ddb.struct2map(yybGift)
	if err != nil {
		logs.Debug("convert map err by %v", err)
		return config.RetCode_SendFailed, false
	}
	if err := ddb.db.SetByHashM(YYBGiftTable, sdkUid, infoMap); err != nil {
		logs.Error("Save DB err %v", err)
		return config.RetCode_SendFailed, false
	}
	return 0, true
}

func (ddb *DBByDynamoDB) checkLoginCondition(sdkUid string, acId string, giftId uint32) (int, bool) {
	ret, err := ddb.db.Get(YYBGiftTable, sdkUid, giftId)
	if err != nil {
		logs.Error("DBByDynamoDB Find err %v", err)
		return config.RetCode_SendFailed, false
	}

	account := loadProfile(acId)
	if account == nil {
		logs.Error("LoadProfile err %v", err)
		return config.RetCode_SendFailed, false
	}
	lastLoginTime := account.Profile.LoginTime
	now_t := time.Now()
	if len(ret) == 0 && now_t.Day() == time.Unix(lastLoginTime, 0).Day() {
		yybGift := &YYBGift{
			SdkUid:  sdkUid,
			GiftID:  int(giftId),
			Times:   1,
			GetTime: time.Now().Unix(),
		}
		infoMap, err := ddb.struct2map(yybGift)
		if err != nil {
			logs.Debug("convert map err by %v", err)
			return config.RetCode_SendFailed, false
		}
		if err := ddb.db.SetByHashM(YYBGiftTable, sdkUid, infoMap); err != nil {
			logs.Error("Save DB err %v", err)
			return config.RetCode_SendFailed, false
		}
	} else if len(ret) > 0 {
		giftInfo, err := ddb.map2struct(ret)
		if err != nil {
			logs.Error("map2struct convert err by %v", err)
			return config.RetCode_SendFailed, false
		}
		if now_t.Day() != time.Unix(giftInfo.GetTime, 0).Day() && now_t.Day() == time.Unix(lastLoginTime, 0).Day() {
			giftInfo.Times += 1
			giftInfo.GetTime = now_t.Unix()
			infoMap, err := ddb.struct2map(giftInfo)
			if err != nil {
				logs.Debug("convert map err by %v", err)
				return config.RetCode_SendFailed, false
			}
			if err := ddb.db.SetByHashM(YYBGiftTable, sdkUid, infoMap); err != nil {
				logs.Error("Save DB err %v", err)
				return config.RetCode_SendFailed, false
			}

		} else if now_t.Day() == time.Unix(giftInfo.GetTime, 0).Day() && now_t.Day() == time.Unix(lastLoginTime, 0).Day() {
			logs.Error("Get already err")
			return config.RetCode_HadSend, false
		} else {
			return config.RetCode_NoCondition, false
		}
	}
	return 0, true
}

func (ddb *DBByDynamoDB) checkSignCondition(sdkUid string, giftId uint32) (int, bool) {
	ret, err := ddb.db.Get(YYBGiftTable, sdkUid, giftId)
	if err != nil {
		logs.Error("DBByDynamoDB Find err %v", err)
		return config.RetCode_SendFailed, false
	}
	if len(ret) <= 0 {
		yybGift := &YYBGift{
			SdkUid:  sdkUid,
			GiftID:  int(giftId),
			Times:   1,
			GetTime: time.Now().Unix(),
		}
		infoMap, err := ddb.struct2map(yybGift)
		if err != nil {
			logs.Debug("convert map err by %v", err)
			return config.RetCode_SendFailed, false
		}
		logs.Debug("infmap: %v", infoMap)
		if err := ddb.db.SetByHashM(YYBGiftTable, sdkUid, infoMap); err != nil {
			logs.Error("Save DB err %v", err)
			return config.RetCode_SendFailed, false
		}
		return 0, true
	}
	giftInfo, err := ddb.map2struct(ret)
	if err != nil {
		logs.Error("map2struct convert err by %v", err)
		return config.RetCode_SendFailed, false
	}
	now_t := time.Now().Unix()
	if time.Now().Day() != time.Unix(giftInfo.GetTime, 0).Day() {
		giftInfo.Times += 1
		giftInfo.GetTime = now_t
		infoMap, err := ddb.struct2map(giftInfo)
		if err != nil {
			logs.Debug("convert map err by %v", err)
			return config.RetCode_SendFailed, false
		}
		if err := ddb.db.SetByHashM(YYBGiftTable, sdkUid, infoMap); err != nil {
			logs.Error("Save DB err %v", err)
			return config.RetCode_SendFailed, false
		}
	} else {
		logs.Error("Get already err %v", err)
		return config.RetCode_HadSend, false
	}
	return 0, true
}
