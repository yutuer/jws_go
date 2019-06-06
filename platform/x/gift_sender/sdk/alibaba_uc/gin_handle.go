package alibaba_uc

import (
	"fmt"

	"time"

	"sync"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/models"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
)

const dbName = "UCGift"

type UCGiftInfo struct {
	AcID    string
	GiftID  string
	GetTime string
}

var Handler UCHandler

type UCHandler struct {
	core.Handler
	db           dynamodb.DynamoDB
	cfg          config.UCConfig
	AuthDB       models.DBInterface
	timeTicker   *time.Ticker
	mutex        sync.RWMutex
	allShardInfo []ShardInfo
}

func (u *UCHandler) GenError() {
	u.ErrorInfo = GetErrorCode()
}
func initAuthDB(cfg config.UCConfig) {
	authConfig.Cfg.Dynamo_NameDevice = cfg.Dynamo_db_Device
	authConfig.Cfg.Dynamo_NameName = cfg.Dynamo_db_Name
	authConfig.Cfg.Dynamo_NameUserInfo = cfg.Dynamo_db_UserInfo
	authConfig.Cfg.Dynamo_GM = cfg.Dynamo_db_GM
	authConfig.Cfg.Dynamo_UserShardInfo = cfg.Dynamo_db_UserShardInfo
	authConfig.Cfg.Runmode = "test"
	authConfig.Cfg.EtcdRoot = config.CommonConfig.EtcdRoot
}

func (u *UCHandler) Reg(e *gin.Engine, cfg config.Config) error {
	u.cfg = cfg.UCConfig
	err := u.initDB(cfg)
	if err != nil {
		return err
	}
	e.POST("getRoleShard", u.getRoleShard)
	e.POST("getAllShard", u.getAllShard)
	e.POST("sendGift", u.sendGift)
	e.POST("checkGift", u.checkGift)
	initAuthDB(u.cfg)
	u.AuthDB = &models.DBByDynamoDB{}
	if err := u.AuthDB.Init(models.DBConfig{
		MongoDBUrl:            "",
		MongoDBName:           "",
		DynamoRegion:          cfg.AWS_Region,
		DynamoAccessKeyID:     cfg.AWS_AccessKey,
		DynamoSecretAccessKey: cfg.AWS_SecretKey,
		DynamoSessionToken:    "",
	}); err != nil {
		panic(fmt.Sprintf("AuthDB init err %v", err))
	}
	err = u.LoadAllShardInfoFromRemote(cfg.UCConfig.GID)
	if err != nil {
		panic(fmt.Sprintf("get shard info err %v", err))
	}
	go func() {
		u.timeTicker = time.NewTicker(30 * time.Minute)
		for {
			<-u.timeTicker.C
			logs.Debug("get all shardInfo")
			err = u.LoadAllShardInfoFromRemote(cfg.UCConfig.GID)
			if err != nil {
				logs.Error("get shard info err %v", err)
			}
		}
	}()
	return nil
}

func init() {
	initErrorCode()
}

func (u *UCHandler) initDB(cfg config.Config) error {
	db := dynamodb.DynamoDB{}
	err := db.Connect(
		cfg.AWS_Region,
		cfg.AWS_AccessKey,
		cfg.AWS_SecretKey,
		"",
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
	if !db.IsTableHasCreate(dbName) {
		return fmt.Errorf("no table named %v in dynamoDB, break!", dbName)
	}
	u.db = db
	return nil
}

func (u *UCHandler) commonCheck(req *Req) int {
	sign, err := GenSign(req.Data.Params, u.cfg.APIKey, u.cfg.Caller)
	logs.Debug("gensign: %v", sign)
	if err != nil {
		return RetCode_InnerError
	}
	if sign != req.Sign {
		return RetCode_MD5SignError
	}
	return RetCode_Success
}

func (u *UCHandler) getRoleShard(c *gin.Context) {
	var req Req
	var reqData GetRoleShardReqData
	err := u.DecodeReq(c, &req, &reqData)
	if err != nil {
		logs.Error("decode req err by %v", err)
		c.JSON(200, u.EncodeRsp(RetCode_AESError, req.ID, nil))
		return
	}
	if code := u.commonCheck(&req); code != RetCode_Success {
		logs.Error("common check err by %v", err)
		c.JSON(200, u.EncodeRsp(code, req.ID, nil))
		return
	}
	uid, err := u.GetUid(reqData.AccountID + "@9@quick.com")
	//uid, err := u.GetUid("3d120cf0e9e8caf9d3dd2caa2236ef1323f81906@06121736168415@apple.com")
	if err != nil {
		logs.Error("acid err by %v", err)
		c.JSON(200, u.EncodeRsp(RetCode_ACIDError, req.ID, nil))
		return
	}
	info, err := u.GetPlayerShardInfo(uid)
	if err != nil {
		logs.Error("get player shard info err by %v", err)
		c.JSON(200, u.EncodeRsp(RetCode_InnerError, req.ID, nil))
		return
	}

	roleInfo := make([]GetRoleShardRspRoleInfoData, 0)
	for _, item := range info {
		roleInfo = append(roleInfo, GetRoleShardRspRoleInfoData{
			ServerID:   item.ServerID,
			ServerName: item.ServerName,
			RoleID:     item.AcID,
			RoleName:   item.RoleName,
			RoleLevel:  item.RoleLevel,
		})
	}
	rspInfo := GetRoleShardRspData{
		AccountID: reqData.AccountID,
		RoleInfos: roleInfo,
	}
	c.JSON(200, u.EncodeRsp(RetCode_Success, req.ID, rspInfo))
	return
}

func (u *UCHandler) getAllShard(c *gin.Context) {
	var req Req
	var reqData GetAllShardInfoReq
	//reqData.Page = 1
	//reqData.Count = 2
	err := u.DecodeReq(c, &req, &reqData)
	if err != nil {
		logs.Error("decode req err by %v", err)
		c.JSON(200, u.EncodeRsp(RetCode_AESError, req.ID, nil))
		return
	}
	if code := u.commonCheck(&req); code != RetCode_Success {
		logs.Error("common check err by %v", err)
		c.JSON(200, u.EncodeRsp(code, req.ID, nil))
		return
	}

	if reqData.Page <= 0 || reqData.Count <= 0 {
		logs.Error("req arg err by %v")
		c.JSON(200, u.EncodeRsp(RetCode_ArgError, req.ID, nil))
		return
	}
	//info := u.GetAllShardInfo(0)
	info := u.GetAllShardInfo(u.cfg.GID)
	serverInfo := make([]GetAllShardRspInfoData, 0)

	j := 0

	for i := int(reqData.Count) * int(reqData.Page-1); i < len(info) && j < int(reqData.Count); i++ {
		serverInfo = append(serverInfo, GetAllShardRspInfoData{
			ServerID:   info[i].ServerID,
			ServerName: info[i].ServerName,
		})
		j++
	}
	rspInfo := GetAllShardInfoRsp{
		RecordCount: len(info),
		List:        serverInfo,
	}
	c.JSON(200, u.EncodeRsp(RetCode_Success, req.ID, rspInfo))
	return
}

func (u *UCHandler) sendGift(c *gin.Context) {
	var req Req
	var reqData SendGiftReq
	err := u.DecodeReq(c, &req, &reqData)
	if err != nil {
		logs.Error("decode req err by %v", err)
		c.JSON(200, u.EncodeRsp(RetCode_AESError, req.ID, nil))
		return
	}
	if code := u.commonCheck(&req); code != RetCode_Success {
		logs.Error("common check err by %v", err)
		c.JSON(200, u.EncodeRsp(code, req.ID, nil))
		return
	}
	if !isLegalGiftID(reqData.KAID) {
		logs.Error("kaid err")
		c.JSON(200, u.EncodeRsp(RetCode_GiftIDError, req.ID, nil))
		return
	}
	//reqData.KAID = "1017.0"
	//reqData.GetDate = "2017-06-14"
	//reqData.RoleId = "0:10:17ea2256-ffb2-4aab-acc9-6d88143cbb6c"
	//reqData.ServerId = "0:10"
	if code, err := u.CheckGiftValid(&reqData); err != nil {
		logs.Error("check gift valid err by %v", err)
		c.JSON(200, u.EncodeRsp(code, req.ID, nil))
		return
	}
	if code, err := u.CheckAcIDValid(reqData.ServerId, reqData.RoleId); err != nil {
		logs.Error("check acid valid err by %v", err)
		c.JSON(200, u.EncodeRsp(code, req.ID, nil))
		return
	}
	nowT := time.Now().Unix()
	err = u.SendGift2Player(reqData.RoleId, reqData.KAID, nowT)
	if err != nil {
		logs.Error("send gift email err by %v", err)
		c.JSON(200, u.EncodeRsp(RetCode_InnerError, req.ID, nil))
		return
	}
	err = u.RecordGift(reqData.RoleId, reqData.KAID, reqData.GetDate)
	if err != nil {
		logs.Error("record gift err by %v", err)
		c.JSON(200, u.EncodeRsp(RetCode_InnerError, req.ID, nil))
		return
	}
	rspInfo := SendGiftRsp{
		Result: "true",
	}
	c.JSON(200, u.EncodeRsp(RetCode_Success, req.ID, rspInfo))
}

func (u *UCHandler) checkGift(c *gin.Context) {
	var req Req
	var reqData CheckGiftReq
	err := u.DecodeReq(c, &req, &reqData)
	if err != nil {
		logs.Error("decode req err by %v", err)
		c.JSON(200, u.EncodeRsp(RetCode_AESError, req.ID, nil))
		return
	}
	if code := u.commonCheck(&req); code != RetCode_Success {
		logs.Error("common check err by %v", err)
		c.JSON(200, u.EncodeRsp(code, req.ID, nil))
		return
	}
	//reqData.KAID = "1010"
	if !isLegalGiftID(reqData.KAID) {
		logs.Error("kaid err")
		c.JSON(200, u.EncodeRsp(RetCode_GiftIDError, req.ID, nil))
		return
	}
	rspInfo := CheckGiftRsp{
		Result: "true",
	}
	c.JSON(200, u.EncodeRsp(RetCode_Success, req.ID, rspInfo))
	return
}

func (u *UCHandler) CheckAcIDValid(serverID string, acID string) (int, error) {
	gid, sid, err := genGIDSID(serverID)
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
	r, err := redis.Int(conn.Do("EXISTS", "profile:"+acID))
	if err != nil {
		return RetCode_InnerError, err
	}
	if r == 1 {
		return RetCode_Success, nil
	} else {
		return RetCode_ACIDError, fmt.Errorf("acid illegal")
	}
}

func (u *UCHandler) CheckGiftValid(info *SendGiftReq) (int, error) {
	checkFunc := commonCheck
	if checkFunc == nil {
		return RetCode_InnerError, fmt.Errorf("no gift type")
	}
	code, err := checkFunc(info, &u.db)
	if err != nil {
		return code, err
	}
	return RetCode_Success, nil
}

func (u *UCHandler) RecordGift(acID string, giftID string, date string) error {
	giftInfo := UCGiftInfo{
		AcID:    acID,
		GiftID:  giftID,
		GetTime: date,
	}
	mapData, err := getGiftMapInfo(&giftInfo)
	if err != nil {
		return err
	}
	err = u.db.SetByHashM(dbName, acID, mapData)
	if err != nil {
		return err
	}
	return nil
}

func (u *UCHandler) SendGift2Player(acID string, giftID string, nowT int64) error {
	tm := core.GetTiMail(fmt.Sprintf("%d", u.cfg.GID))
	if tm == nil {
		return fmt.Errorf("No mailDB for gid: %v", u.cfg.GID)
	}
	data := config.GetUCGiftData(giftID)
	if len(data.GetItem_Table()) <= 0 {
		return fmt.Errorf("no gift id for %v", giftID)
	}
	itemid := make([]string, 0)
	itemc := make([]uint32, 0)
	for _, item := range data.GetItem_Table() {
		itemid = append(itemid, item.GetItemID())
		itemc = append(itemc, item.GetItemCount())
	}
	if err := tm.SendMail(fmt.Sprintf("profile:%s", acID), timail.MailReward{
		Param:     []string{},
		Idx:       timail.MkMailId(timail.Mail_Send_By_Sys, 0),
		TimeBegin: nowT,
		TimeEnd:   nowT + int64(util.DaySec*7),
		IdsID:     data.GetMailType(),
		ItemId:    itemid,
		Count:     itemc,
		Reason:    "UCGift",
	}); err != nil {
		return fmt.Errorf("sendMail err %v", err)
	}
	return nil
}

func isLegalGiftID(giftID string) bool {
	return config.GetUCGiftData(giftID) != nil
}
