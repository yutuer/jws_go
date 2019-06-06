package hmt_gift

import (
	"fmt"
	"strconv"
	"strings"

	"encoding/json"

	"time"

	"github.com/gin-gonic/gin"
	"sort"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
	"sync/atomic"
)

const ac_iota = "ac_iota"
const dbName = "HMTGift"
const cpUidLen = 8

var mailId int64

var Handler HmtHandler

var title_info = []string{"cp_uid", "server_id", "role_id", "reward_list", "reward_name", "description", "time", "record_id"}

var title_heroInfo = []string{"info_name", "start_time", "end_time", "cp_uid", "server_id", "role_id"}

type HmtHandler struct {
	core.Handler
	db dynamodb.DynamoDB
}

func (h *HmtHandler) GenError() {
	h.ErrorInfo = GetErrorCode()
}

func (h *HmtHandler) Reg(e *gin.Engine, cfg config.Config) error {
	err := h.initDB(cfg)
	if err != nil {
		return err
	}
	e.POST("sendBigGift", h.sendBigGifg)
	e.POST("getHeroInfo", h.getHeroInfo)

	return nil
}

func init() {
	initErrorCode()
}

type GiftInfomation struct {
	CpUid       int        `json:"cp_uid"`
	ServerId    string     `json:"server_id"`
	RoleId      string     `json:"role_id"`
	ItemInfo    []ItemInfo `json:"item_info"`
	RewardName  string     `json:"reward_name"`
	Description string     `json:"description"`
	Time        string     `json:"time"`
	RewandId    string     `json:"rewand_id"`
}
type ItemInfo struct {
	PropID  string `json:"propid"`
	PropNum int    `json:"pronum"`
}

type Ret struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
}

type HeroInfomation struct {
	Info_name  string
	Start_time int64
	End_time   int64
	CpUid      int
	ServerId   string
	RoleId     string
}

func genRet(code int) string {
	if code == 0 {
		ret := Ret{
			Code:   code,
			Reason: "success",
		}
		jsonStr, err := json.Marshal(ret)
		if err == nil {
			return string(jsonStr)
		} else {
			return ""
		}
	} else {
		ret := Ret{
			Code:   -1,
			Reason: Handler.GetMsg(code),
		}
		jsonStr, err := json.Marshal(ret)
		if err == nil {
			return string(jsonStr)
		} else {
			return ""
		}
	}
}

func (h *HmtHandler) initDB(cfg config.Config) error {
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
	h.db = db
	return nil
}

func (h *HmtHandler) sendBigGifg(c *gin.Context) {
	logs.Info("receive timestamp: %s request", c.PostForm("timestamp"))
	logs.Info("Get cp_uid: %d", c.PostForm("cp_uid"))

	if len(c.PostForm("cp_uid")) != cpUidLen {
		logs.Debug("cp_uid Is Wrong")
		c.String(200, genRet(RetCode_UserInfoError))
		return
	}

	data := giftInfoSort(c, title_info)
	if data == "wrong" {
		logs.Debug("Missing parameter")
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	if sign, _ := CalSign(data); sign != c.PostForm("sign") {
		logs.Debug("illegal sign: %s cmp right sign: %s", sign, c.PostForm("sign"))
		c.String(200, genRet(RetCode_InvalidSign))
		return
	}
	logs.Info("sign verified")

	giftInfo, err := ParseGiftInfoFromJson(c)
	if err != nil {
		logs.Error("ParseGiftInfoFromJson err by %v", err)
		c.String(200, genRet(RetCode_NoItem))
		return
	}
	logs.Info("unmarshal giftInfo: %v", giftInfo)
	ret, err := h.db.Get(dbName, giftInfo.RoleId, giftInfo.RewandId)
	if err != nil {
		logs.Debug("DynamoDB Has Problem %v", err)
		c.String(200, genRet(RetCode_Default))
		return
	}
	if len(ret) > 0 {
		logs.Debug("already got: %v", ret)
		c.String(200, genRet(RetCode_Success))
		return
	}

	ids, _ := getFormatItem(giftInfo.ItemInfo)
	if !config.IsHmtGiftItem(ids) {
		logs.Error("check item valid err")
		c.String(200, genRet(RetCode_NoItem))
		return
	}
	logs.Info("Gift Item Is ok %v", giftInfo.ItemInfo)
	gid, sid, err := getGIDFromeNumberGID(giftInfo.ServerId)
	if err != nil {
		logs.Error("getGIDFromeNumberGID err by %v", err)
		c.String(200, genRet(RetCode_UserInfoError))
		return
	}
	logs.Info("parse gid: %d sid: %d", gid, sid)
	if _, err := CheckAcIDValid(uint(gid), uint(sid), giftInfo.RoleId); err != nil {
		logs.Error("check acid valid err by %v", err)
		c.String(200, genRet(RetCode_UserInfoError))
		return
	}
	logs.Info("parse acID is %s", giftInfo.RoleId)

	err = sendMail2Player(giftInfo.RoleId, giftInfo, gid)
	if err != nil {
		logs.Error("sendMail2Player err by %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	dataMap := make(map[string]interface{})
	dataMap["GiftID"] = giftInfo.RewandId
	err = h.db.SetByHashM(dbName, giftInfo.RoleId, dataMap)
	if err != nil {
		logs.Error("sendHMTGift err by %v", err)
		c.String(200, genRet(RetCode_Default))
		return
	}
	c.String(200, genRet(RetCode_Success))
}

func (h *HmtHandler) getHeroInfo(c *gin.Context) {
	logs.Info("Get cp_uid: %d", c.PostForm("cp_uid"))

	if len(c.PostForm("cp_uid")) != cpUidLen {
		logs.Debug("cp_uid Is Wrong")
		c.String(200, genRet(RetCode_UserInfoError))
		return
	}

	data := giftInfoSort(c, title_heroInfo)
	if data == "wrong" {
		logs.Debug("Missing parameter")
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	if sign, _ := CalSign(data); sign != c.PostForm("sign") {
		logs.Debug("illegal sign: %s cmp right sign: %s", sign, c.PostForm("sign"))
		c.String(200, genRet(RetCode_InvalidSign))
		return
	}
	logs.Info("sign verified")

	heroInfo, err := ParseHeroInfoFromJson(c)
	if err != nil {
		logs.Error("ParseHeroInfoFromJson err by %v", err)
		c.String(200, genRet(RetCode_NoItem))
		return
	}
	logs.Info("unmarshal heroInfo: %v", heroInfo)

	gid, sid, err := getGIDFromeNumberGID(heroInfo.ServerId)
	if err != nil {
		logs.Error("getGIDFromeNumberGID err by %v", err)
		c.String(200, genRet(RetCode_UserInfoError))
		return
	}
	logs.Info("parse gid: %d sid: %d", gid, sid)
	if _, err := CheckAcIDValid(uint(gid), uint(sid), heroInfo.RoleId); err != nil {
		logs.Error("check acid valid err by %v", err)
		c.String(200, genRet(RetCode_UserInfoError))
		return
	}
	logs.Info("parse acID is %s", heroInfo.RoleId)
	heroHmtInfoData, err := getHmtHeroInfo(uint(gid), uint(sid), heroInfo.RoleId)
	if err != nil {
		logs.Error("get data from redis err by %v", err)
		c.String(200, genRet(RetCode_GetInfoError))
		return
	}
	result, infoName, codeError := getHmtActivityResult(heroInfo, heroHmtInfoData)
	if codeError > 0 {
		logs.Error("get Error By %s", genRet(codeError))
		c.String(200, genRet(codeError))
		return
	}
	type r struct {
		Type string `json:"type"`
		Num  int    `json:"num"`
	}
	Result := r{infoName, result}

	f := struct {
		Code   int `json:"code"`
		Result r   `json:"result"`
	}{0, Result}
	logs.Info("Get Info Success!")
	c.JSON(200, f)
}

func CheckAcIDValid(gid, sid uint, acID string) (int, error) {
	conn, err := core.GetGamexProfileRedisConn(gid, sid)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return 0, err
	}
	r, err := redis.Int(conn.Do("EXISTS", "profile:"+acID))
	if err != nil {
		return 0, err
	}
	if r == 1 {
		return RetCode_Success, nil
	} else {
		return 1, fmt.Errorf("acid illegal")
	}
}

func getGIDFromeNumberGID(serverNumberID string) (int, int, error) {
	//serverID, ok := ParseServerID(serverNumberID)
	//if !ok {
	//	return 0, 0, fmt.Errorf("parse server id error by serverNum: %d", serverNumberID)
	//}
	if !strings.Contains(serverNumberID, ":") {
		return 0, 0, fmt.Errorf("parse server id error by serverNum: %s", serverNumberID)
	}
	gidsid := strings.Split(serverNumberID, ":")
	gid, err := strconv.Atoi(gidsid[0])
	if err != nil {
		return 0, 0, err
	}
	sid, err := strconv.Atoi(gidsid[1])
	if err != nil {
		return 0, 0, err
	}
	return gid, sid, nil
}

func getFormatItem(info []ItemInfo) ([]string, []uint32) {
	id := make([]string, 0)
	count := make([]uint32, 0)
	for _, item := range info {
		id = append(id, item.PropID)
		count = append(count, uint32(item.PropNum))
	}
	return id, count
}

func sendMail2Player(acID string, gift GiftInfomation, gid int) error {
	atomic.AddInt64(&mailId,1)
	tm := core.GetTiMail(strconv.Itoa(gid))
	if tm == nil {
		return fmt.Errorf("No mailDB for gid: %v", gid)
	}
	for _, item := range gift.ItemInfo {
		if !CheckPropExist(item.PropID) {
			return fmt.Errorf("ItemInfo error")
		}
	}
	item, count := getFormatItem(gift.ItemInfo)
	timeNow := time.Now().Unix()
	s := timail.MailReward{
		ItemId:       item,
		Count:        count,
		Reason:       gift.RewardName,
		Tag:          "{}",
		Account2Send: []string{acID},
		Idx:          timail.MkMailId(timail.Mail_Send_By_Sys, atomic.LoadInt64(&mailId)),
		Param:        []string{gift.RewardName, gift.Description},
		TimeBegin:    timeNow,
		TimeEnd:      timeNow + int64(util.DaySec*7),
	}
	err := tm.SendMail(fmt.Sprintf("profile:%s", acID), s)
	if err != nil {
		return err
	}
	return nil
}

func CheckPropExist(itemID string) bool {
	_, ok := gamedata.GetProtoItem(itemID)
	return ok
}

func ParseGiftInfoFromJson(c *gin.Context) (GiftInfomation, error) {
	cpUid, _ := strconv.Atoi(c.PostForm("cp_uid"))
	info := GiftInfomation{
		CpUid:       cpUid,
		ServerId:    c.PostForm("server_id"),
		RoleId:      c.PostForm("role_id"),
		RewardName:  c.PostForm("reward_name"),
		Description: c.PostForm("description"),
		Time:        c.PostForm("time"),
		RewandId:    c.PostForm("record_id"),
	}
	var dat map[string]interface{}
	err := json.Unmarshal([]byte(c.PostForm("reward_list")), &dat)
	info.ItemInfo = make([]ItemInfo, 0)
	if err != nil {
		return info, err
	}
	for v, x := range dat {
		info.ItemInfo = append(info.ItemInfo, ItemInfo{
			PropID:  v,
			PropNum: int(x.(float64)),
		})
	}
	//err := json.Unmarshal([]byte(c.PostForm("reward_list")), &info.ItemInfo)
	logs.Debug("gift Info %s", info)
	return info, nil
}

func ParseHeroInfoFromJson(c *gin.Context) (HeroInfomation, error) {
	s := HeroInfomation{}
	cpUid, err := strconv.Atoi(c.PostForm("cp_uid"))
	if err != nil {
		return s, err
	}
	endTime, err := strconv.ParseInt(c.PostForm("end_time"), 10, 64)
	if err != nil {
		return s, err
	}
	startTime, err := strconv.ParseInt(c.PostForm("start_time"), 10, 64)
	if err != nil {
		return s, err
	}
	info := HeroInfomation{
		Info_name:  c.PostForm("info_name"),
		Start_time: startTime,
		End_time:   endTime,
		CpUid:      cpUid,
		ServerId:   c.PostForm("server_id"),
		RoleId:     c.PostForm("role_id"),
	}
	logs.Debug("[Require Hero] Info %s", info)
	return info, nil
}

func giftInfoSort(c *gin.Context, titleInfo []string) string {
	var data string
	isInfo := make([]string, 0)
	for _, x := range titleInfo {
		if c.PostForm(x) != "" {
			isInfo = append(isInfo, x)
		}
	}
	if len(isInfo) != len(titleInfo) {
		return "wrong"
	}
	sort.Strings(isInfo)
	logs.Debug("Sort IsInfo %s", isInfo)
	for _, x := range isInfo {
		if data == "" {
			data = x + "=" + c.PostForm(x)
		} else {
			data = data + "&" + x + "=" + c.PostForm(x)
		}
	}
	logs.Debug("receive Data %s", data)
	return data
}

func getHmtHeroInfo(gid, sid uint, acid string) (account.HmtPlayerActivityInfoInDB, error) {
	profileID := "profile:" + acid
	//链接玩家个人信息数据库db
	var herodata account.HmtPlayerActivityInfoInDB
	conn1, err := core.GetGamexProfileRedisConn(gid, sid)
	defer func() {
		if conn1 != nil {
			conn1.Close()
		}
	}()
	if err != nil {
		return herodata, err
	}
	logs.Debug("Begin get hero activity info")
	heros, err := conn1.Do("HGET", profileID, "hmt_activity_info")
	if err != nil {
		return herodata, err
	} else if heros != nil {
		logs.Debug("Begin unmarshal hero activity info")
		err := json.Unmarshal(heros.([]byte), &herodata)
		if err != nil {
			return herodata, err
		}
	}
	return herodata, nil
}

func getHmtActivityResult(info HeroInfomation, activityInfo account.HmtPlayerActivityInfoInDB) (int, string, int) {
	var result int
	var tempDay int64
	tempDay = info.Start_time
	date := make([]string, 0)
	logs.Debug("time start %d, end %d,tempDay %d", info.Start_time, info.End_time, tempDay)
	if info.Start_time > info.End_time {
		return 0, "", RetCode_TimeError
	}
	logs.Debug("tempDay %s,StartTime %s, EndTime %s", formatDate(tempDay), formatDate(info.Start_time), formatDate(info.End_time))
	for {
		if formatDate(tempDay) == formatDate(info.End_time) {
			date = append(date, formatDate(tempDay))
			break
		}
		date = append(date, formatDate(tempDay))
		tempDay += int64(util.DaySec)
	}
	logs.Debug("Serach Date %v", date)
	for _, hDate := range date {
		for _, dDate := range activityInfo.DateActivityInfo {
			if hDate == dDate.DailyDate {
				switch info.Info_name {
				case "login_daily":
					if dDate.IsLogin {
						result += 1
					}
				case "dungeon_jy":
					pNum := sumList(dDate.DungeonJy, dDate.DungeonJyTime, info.End_time, info.Start_time)
					result += pNum
				case "dungeon_dy":
					pNum := sumList(dDate.DungeonDy, dDate.DungeonDyTime, info.End_time, info.Start_time)
					result += pNum
				case "draw":
					pNum := sumList(dDate.GachaNum, dDate.GachaNumTime, info.End_time, info.Start_time)
					result += pNum
				default:
					return 0, "", RetCode_GetInfoError
				}
			}

		}
	}
	return result, info.Info_name, RetCode_Success
}

func formatDate(nowt int64) string {
	tm := time.Unix(nowt, 0).In(util.ServerTimeLocal)
	year, month, day := tm.Date()
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}

func sumList(palyNum []int, palyTime []int64, endTime, startTime int64) int {
	var result int
	if len(palyTime) == 0 {
		return 0
	}
	for i, pTime := range palyTime {
		if pTime > endTime || pTime < startTime {
			continue
		}
		result += palyNum[i]
	}
	return result
}
