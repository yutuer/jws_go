package alibaba_uc

import (
	"fmt"

	"time"

	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
)

const (
	signGift     = "0"
	weekSignGift = "1"
	levelGift    = "2"
	payGift      = "3"
	vipGift      = "4"
	vipSignGift  = "5"
)

type checkFunc func(info *SendGiftReq, db *dynamodb.DynamoDB) (int, error)

func checkLevelGift(info *SendGiftReq, db *dynamodb.DynamoDB) (int, error) {
	ret, err := db.Get(dbName, info.RoleId, info.KAID)
	if err != nil {
		return RetCode_InnerError, err
	}
	if len(ret) > 0 {
		return RetCode_AlreadyGot, fmt.Errorf("already got")
	}
	gid, sid, err := genGIDSID(info.ServerId)
	if err != nil {
		return RetCode_ServerInfoError, err
	}
	conn, err := core.GetGamexProfileRedisConn(gid, sid)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return RetCode_InnerError, err
	}
	reply, err := conn.Do("HGET", "profile:"+info.RoleId, "corp")
	if err != nil {
		return RetCode_InnerError, err
	}
	level, err := getLvFromJson(reply)
	if err != nil {
		return RetCode_InnerError, fmt.Errorf("getLvInfoFromJson err by %v ", err)
	}
	data := config.GetUCGiftData(info.KAID)
	if level < int(data.GetFCValue1()) {
		return RetCode_RoleInfoError, fmt.Errorf("level too low")
	}
	return 0, nil
}

func checkSignGift(info *SendGiftReq, db *dynamodb.DynamoDB) (int, error) {
	ret, err := db.Get(dbName, info.RoleId, info.KAID)
	if err != nil {
		return RetCode_InnerError, err
	}
	logs.Debug("get sign gift db data: %v", ret)
	giftInfo, err := getGiftInfo(ret)
	if err != nil {
		return RetCode_InnerError, err
	}
	t, err := time.Parse("2006-01-02", info.GetDate)
	if err != nil {
		return RetCode_InnerError, err
	}
	if len(ret) <= 0 || !isSameDay(fmt.Sprintf("%d", t.Unix()), giftInfo.GetTime) {
		return RetCode_Success, nil
	}
	return RetCode_AlreadyGot, fmt.Errorf("already got")
}

func checkWeekSignGift(info *SendGiftReq, db *dynamodb.DynamoDB) (int, error) {
	ret, err := db.Get(dbName, info.RoleId, info.KAID)
	if err != nil {
		return RetCode_InnerError, err
	}
	giftInfo, err := getGiftInfo(ret)
	if err != nil {
		return RetCode_InnerError, err
	}
	t, err := time.Parse("2006-01-02", info.GetDate)
	if err != nil {
		return RetCode_InnerError, err
	}
	if giftInfo == nil || !isSameWeekDay(fmt.Sprintf("%d", t.Unix()), giftInfo.GetTime) {
		return RetCode_Success, nil
	}
	return RetCode_AlreadyGot, fmt.Errorf("already got")
	return 0, nil
}

func checkPayGift(info *SendGiftReq, db *dynamodb.DynamoDB) (int, error) {
	ret, err := db.Get(dbName, info.RoleId, info.KAID)
	if err != nil {
		return RetCode_InnerError, err
	}
	if len(ret) > 0 {
		return RetCode_AlreadyGot, fmt.Errorf("already got")
	}
	gid, sid, err := genGIDSID(info.ServerId)
	if err != nil {
		return RetCode_ServerInfoError, err
	}
	conn, err := core.GetGamexProfileRedisConn(gid, sid)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return RetCode_InnerError, err
	}
	reply, err := conn.Do("HGET", "profile:"+info.RoleId, "v")
	if err != nil {
		return RetCode_InnerError, err
	}
	HCCount, err := getHCCountFromJson(reply)
	if err != nil {
		return RetCode_InnerError, fmt.Errorf("gethccount err by %v ", err)
	}
	data := config.GetUCGiftData(info.KAID)
	logs.Debug("HCCount: %v", HCCount)
	if HCCount < int(data.GetFCValue1()) {
		return RetCode_RoleInfoError, fmt.Errorf("pay too little")
	}
	return RetCode_Success, nil
}

func checkVIPGift(info *SendGiftReq, db *dynamodb.DynamoDB) (int, error) {
	ret, err := db.Get(dbName, info.RoleId, info.KAID)
	if err != nil {
		return RetCode_InnerError, err
	}
	if len(ret) > 0 {
		return RetCode_AlreadyGot, fmt.Errorf("already got")
	}
	gid, sid, err := genGIDSID(info.ServerId)
	if err != nil {
		return RetCode_ServerInfoError, err
	}
	conn, err := core.GetGamexProfileRedisConn(gid, sid)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return RetCode_InnerError, err
	}
	reply, err := conn.Do("HGET", "profile:"+info.RoleId, "v")
	if err != nil {
		return RetCode_InnerError, err
	}
	vipLevel, err := getVIPLvFromJson(reply)
	if err != nil {
		return RetCode_InnerError, fmt.Errorf("getLvInfoFromJson err by %v ", err)
	}
	if vipLevel <= 0 {
		return RetCode_RoleInfoError, fmt.Errorf("no vip")
	}
	return RetCode_Success, nil
}

func checkVIPSignGift(info *SendGiftReq, db *dynamodb.DynamoDB) (int, error) {
	ret, err := db.Get(dbName, info.RoleId, info.KAID)
	if err != nil {
		return RetCode_InnerError, err
	}
	giftInfo, err := getGiftInfo(ret)
	if err != nil {
		return RetCode_InnerError, err
	}
	if len(ret) > 0 && isSameDay(info.GetDate, giftInfo.GetTime) {
		return RetCode_AlreadyGot, fmt.Errorf("already got")
	}
	gid, sid, err := genGIDSID(info.ServerId)
	if err != nil {
		return RetCode_ServerInfoError, err
	}
	conn, err := core.GetGamexProfileRedisConn(gid, sid)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return RetCode_InnerError, err
	}
	reply, err := conn.Do("HGET", "profile:"+info.RoleId, "v")
	if err != nil {
		return RetCode_InnerError, err
	}
	vipLevel, err := getVIPLvFromJson(reply)
	if err != nil {
		return RetCode_InnerError, fmt.Errorf("getLvInfoFromJson err by %v ", err)
	}
	if vipLevel <= 0 {
		return RetCode_RoleInfoError, fmt.Errorf("no vip")
	}
	return RetCode_Success, nil
}

func commonCheck(info *SendGiftReq, db *dynamodb.DynamoDB) (int, error) {
	ret, err := db.Get(dbName, info.RoleId, info.KAID)
	if err != nil {
		return RetCode_InnerError, err
	}
	logs.Debug("get sign gift db data: %v", ret)
	giftInfo, err := getGiftInfo(ret)
	if err != nil {
		return RetCode_InnerError, err
	}
	if len(ret) <= 0 || !isSameDay2(info.GetDate, giftInfo.GetTime) {
		return RetCode_Success, nil
	}
	logs.Debug("get sign gift db data2: %v", *giftInfo)
	return RetCode_AlreadyGot, fmt.Errorf("already got")
}

func getCheckFunc(giftID string) checkFunc {
	data := config.GetUCGiftData(giftID)
	typ := data.GetGiftType()
	switch typ {
	case levelGift:
		return checkLevelGift
	case signGift:
		return checkSignGift
	case weekSignGift:
		return checkWeekSignGift
	case payGift:
		return checkPayGift
	case vipGift:
		return checkVIPGift
	case vipSignGift:
		return checkVIPSignGift
	}
	return nil
}
