package hero

import (
	"fmt"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/astaxie/beego/httplib"
	"github.com/gin-gonic/gin"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	u_cfg "vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/hero_gag"
)

const ac_iota = "ac_iota"

var Handler HeroHandler

type HeroHandler struct {
	core.Handler
}

func (h *HeroHandler) GenError() {
	h.ErrorInfo = GetErrorCode()
}

func (h *HeroHandler) Reg(e *gin.Engine, cfg config.Config) error {
	// 限定IP
	var IPLimit iplimitconfig.IPLimits
	u_cfg.NewConfigToml("iplimit.toml", &IPLimit)
	hero_gag.SetCommonIPLimitCfg(IPLimit.IPLimit)
	g := e.Group("", hero_gag.CommonIPLimit())

	g.POST("user/sendprize.html", handle)
	g.POST("user/costdiamond.html", CostDiamond)
	return nil
}

func init() {
	initErrorCode()
}

type GiftInfo struct {
	RoleID      int    `json:"roleid"`
	ServerID    int    `json:"serverid"`
	MailTitle   string `json:"mailTitle"`
	MailContent string `json:"mailContent"`
	ItemInfo    []ItemInfo
	ItemInfoSrc string `json:"itemInfo"`
}

type ItemInfo struct {
	PropID  string `json:"propid"`
	PropNum int    `json:"pronum"`
}

type Ret struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func genRet(code int) string {
	ret := Ret{
		Code: code,
		Msg:  Handler.GetMsg(code),
	}
	jsonStr, err := json.Marshal(ret)
	if err == nil {
		return string(jsonStr)
	} else {
		return ""
	}
}

func handle(c *gin.Context) {
	logs.Info("receive timestamp: %s request", c.PostForm("timestamp"))
	if sign := CalSign("data="+c.PostForm("data")+"&"+
		"timestamp="+c.PostForm("timestamp"), config.CommonConfig.SecretKey); sign != c.PostForm("sign") {
		logs.Debug("illegal sign: %s cmp right sign: %s", sign, c.PostForm("sign"))
		c.String(200, genRet(RetCode_InvalidSign))
		return
	}
	logs.Info("sign verified")
	srcData, err := DecodeData([]byte(c.PostForm("data")))
	if err != nil {
		logs.Error("decodeData err by %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	logs.Info("decode data: %v", srcData)
	giftInfo, err := ParseGiftInfoFromJson(srcData)
	if err != nil {
		logs.Error("ParseGiftInfoFromJson err by %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	logs.Info("unmarshal giftInfo: %v", giftInfo)

	gid, sid, err := getGIDFromeNumberGID(giftInfo.ServerID)
	if err != nil {
		logs.Error("getGIDFromeNumberGID err by %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	acID, err := getAcIDFromNumberID(giftInfo.RoleID, gid, sid)
	if err != nil {
		logs.Error("getAcIDFromNumberID err by %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	logs.Info("parse acID is %s", acID)
	err = sendMail2Player(acID, giftInfo, gid)
	if err != nil {
		logs.Error("sendMail2Player err by %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	c.String(200, genRet(RetCode_Success))
}

func getAcIDFromNumberID(numberID int, gid, sid int) (string, error) {
	conn, err := core.GetAuthRedisConn(uint(gid), uint(sid))
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return "", err
	}
	acID, err := redis.String(conn.Do("HGET", ac_iota, numberID))
	if err != nil {
		return "", err
	}
	if strings.Split(acID, ":")[1] != strconv.Itoa(sid) {
		return "", fmt.Errorf("sid illegal for numberID: %d, gid: %d, sid: %d, acID: %s", acID)
	}
	return acID, nil
}

func getGIDFromeNumberGID(serverNumberID int) (int, int, error) {
	serverID, ok := ParseServerID(serverNumberID)
	if !ok {
		return 0, 0, fmt.Errorf("parse server id error by serverNum: %d", serverNumberID)
	}
	gidsid := strings.Split(serverID, ":")
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

func sendMail2Player(acID string, gift GiftInfo, gid int) error {
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
		Reason:       "HeroGift",
		Tag:          "{}",
		Account2Send: []string{acID},
		Idx:          timail.MkMailId(timail.Mail_Send_By_Sys, 0),
		Param:        []string{gift.MailTitle, gift.MailContent},
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

func ParseGiftInfoFromJson(data []byte) (GiftInfo, error) {
	info := GiftInfo{}
	err := json.Unmarshal(data, &info)
	if err != nil {
		return GiftInfo{}, err
	}
	err = json.Unmarshal([]byte(info.ItemInfoSrc), &info.ItemInfo)
	if err != nil {
		return info, err
	}
	return info, nil
}

type RspInfo struct {
	Ret     int    `json:"ret"`
	Diamond string `json:"diamond"`
}

type ProInfo struct {
	Roleid   int `json:"roleid"`
	Serverid int `json:"serverid"`
	DiamNum  int `json:"diamond_num"`
}

type Rsp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    int    `json:"data"`
	Diamond int    `json:"diamond"`
}

func CostDiamond(c *gin.Context) {
	param := make(map[string]string, 3)
	param["data"] = c.PostForm("data")
	//signOrg := c.PostForm("sign")
	param["timestamp"] = c.PostForm("timestamp")
	logs.Debug("receive params %v", param)
	Key := config.CommonConfig.SecretKey
	logs.Debug(Key)

	if sign := CalSign("data="+c.PostForm("data")+"&"+
		"timestamp="+c.PostForm("timestamp"), Key); sign != c.PostForm("sign") {
		logs.Error("illegal sign: %s cmp right sign: %s", sign, c.PostForm("sign"))
		c.String(200, genRet(RetCode_InvalidSign))
		return
	}

	bsOut, err := DecodeData([]byte(param["data"]))
	if err != nil {
		logs.Error("Decode data error %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	//进行解密
	logs.Debug("%s\n", bsOut)
	info := ProInfo{}

	err = json.Unmarshal(bsOut, &info)
	if err != nil {
		logs.Error("json: error CostDiamInfo decoding json binary '%s': %v", bsOut, err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}

	logs.Debug("After all unmarshal:%v", info)

	serverID, ok := ParseServerID(info.Serverid)
	if !ok {
		logs.Error("parse server id error by serverNum: %d", serverID)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}

	g_sid := strings.Split(serverID, ":")
	if g_sid[0] == "" {
		logs.Error("wrong gid")
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	gid, err := strconv.Atoi(g_sid[0])
	if err != nil {
		logs.Error("wrong gid error %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	if g_sid[1] == "" {
		logs.Error("putin wrong sid")
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	sid, err := strconv.Atoi(g_sid[1])
	if err != nil {
		logs.Error("putin wrong sid error %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	logs.Debug("%s/%d/%d/internalip", config.CommonConfig.EtcdRoot, gid, sid)
	_redis_internalip, err := etcd.Get(fmt.Sprintf("%s/%d/%d/internalip",
		config.CommonConfig.EtcdRoot,
		gid, sid))
	if err != nil {
		logs.Error("Get ip error %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	redis_ip := strings.Split(_redis_internalip, ":")
	logs.Debug("%s/%d/%d/internalip:%s", config.CommonConfig.EtcdRoot, gid, sid, _redis_internalip)
	_redis_listen_post_port, err := etcd.Get(fmt.Sprintf("%s/%d/%d/listen_post_port",
		config.CommonConfig.EtcdRoot,
		gid, sid))
	if err != nil {
		logs.Error("Get port error %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	logs.Debug("%s/%d/%d/internalip:%s", config.CommonConfig.EtcdRoot, gid, sid, "http://"+redis_ip[0]+":"+_redis_listen_post_port)

	roleId, err := getAcIDFromNumberID(info.Roleid, gid, sid)
	if err != nil {
		logs.Error("Get acid error: %d,  %v", info.Roleid, err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}

	resp := httplib.Post("http://"+redis_ip[0]+":"+_redis_listen_post_port+"/h5shop/CostDiamondGamex").SetTimeout(5*time.Second, 5*time.Second).
		Param("id", roleId).Param("serverid", serverID).Param("num", strconv.Itoa(info.DiamNum))
	res, err := resp.String()
	if err != nil {
		logs.Error("%v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	logs.Debug("Rsp Info from gamex %s", res)
	tfo := RspInfo{}
	json.Unmarshal([]byte(res), &tfo)
	if err != nil {
		logs.Error("Get rep from gamex error %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	num, err := strconv.Atoi(tfo.Diamond)
	if err != nil {
		logs.Error("Convert diamond number string to int error %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	ret, err := json.Marshal(Rsp{
		RetCode_Success,
		Handler.GetMsg(RetCode_Success),
		tfo.Ret,
		num,
	})
	if err != nil {
		logs.Error("Json Marshal error %v", err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	c.String(200, string(ret))
}
