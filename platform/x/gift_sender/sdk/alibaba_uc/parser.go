package alibaba_uc

import (
	"fmt"
	"strconv"
	"strings"

	"encoding/json"

	"crypto/aes"
	"crypto/cipher"

	"encoding/base64"

	"crypto/md5"
	"encoding/hex"

	"time"

	"bytes"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/auth/models"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
)

// 网络数据 marshal struct
type Req struct {
	ID      int64      `json:"long"`
	Client  CallerData `json:"client"`
	Encrypt string     `json:"encrypt"`
	Sign    string     `json:"sign"`
	Data    ReqParam   `json:"data"`
}

type ReqParam struct {
	Params string `json:"params"`
}

type Rsp struct {
	ID    int64    `json:"id"`
	State RspState `json:"state"`
	Data  string   `json:"data,omitempty"`
}

type CallerData struct {
	Caller string `json:"caller"`
	Ex     string `json:"ex"`
}

type RspState struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type GetRoleShardReqData struct {
	AccountID string  `json:"accountId"`
	GameID    float64 `json:"gameId"`
	Platform  float64 `json:"platform"`
}

type GetRoleShardRspData struct {
	AccountID string                        `json:"accountId"`
	RoleInfos []GetRoleShardRspRoleInfoData `json:"roleInfos"`
}

type GetRoleShardRspRoleInfoData struct {
	ServerID   string `json:"serverId"`
	ServerName string `json:"serverName"`
	RoleID     string `json:"roleId"`
	RoleName   string `json:"roleName"`
	RoleLevel  int    `json:"roleLevel"`
}

type GetAllShardInfoReq struct {
	GameID   float64 `json:"gameId"`
	Platform float64 `json:"platform"`
	Page     float64 `json:"page"`
	Count    float64 `json:"count"`
}

type GetAllShardInfoRsp struct {
	RecordCount int                      `json:"recordCount"`
	List        []GetAllShardRspInfoData `json:"list"`
}

type GetAllShardRspInfoData struct {
	ServerID   string `json:"serverId"`
	ServerName string `json:"serverName"`
}

type SendGiftReq struct {
	AccountID string  `json:"accountId"`
	GameID    float64 `json:"gameId"`
	KAID      string  `json:"kaId"`
	ServerId  string  `json:"serverId"`
	RoleId    string  `json:"roleId"`
	GetDate   string  `json:"getDate"`
}

type SendGiftRsp struct {
	Result string `json:"result"`
}

type CheckGiftReq struct {
	GameID float64 `json:"gameId"`
	KAID   string  `json:"kaId"`
}

type CheckGiftRsp struct {
	Result string `json:"result"`
}

func (u *UCHandler) DecodeReq(c *gin.Context, req *Req, reqData interface{}) error {
	err := c.BindJSON(&req)
	if err != nil {
		return err
	}
	logs.Debug("decode req: %v", req)
	jsonData, err := aesDecode(req.Data.Params, u.cfg.SecretKey)
	if err != nil {
		return err
	}
	logs.Debug("aesDecode data: %v", string(jsonData))
	err = json.Unmarshal(jsonData, reqData)
	if err != nil {
		return err
	}
	logs.Debug("unmarshal json: %v", reqData)
	return nil
}

// TODO by ljz  handle error
func (u *UCHandler) EncodeRsp(code int, ID int64, data interface{}) *Rsp {
	logs.Debug("rsp data: %v", data)
	var rsp *Rsp
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil
		}
		encodeData, err := aesEncode(jsonData, u.cfg.SecretKey)
		if err != nil {
			return nil
		}
		rsp = &Rsp{
			ID: ID,
			State: RspState{
				Code: code,
				Msg:  u.ErrorInfo[code],
			},
			Data: encodeData,
		}
	} else {
		rsp = &Rsp{
			ID: ID,
			State: RspState{
				Code: code,
				Msg:  u.ErrorInfo[code],
			},
		}
	}
	logs.Debug("encode rsp: %v", *rsp)
	return rsp

}

/**********************************************************************/

// 内部struct
type RoleShardInfo struct {
	ServerName string
	ServerID   string
	RoleName   string
	RoleLevel  int
	AcID       string
}

type ShardInfo struct {
	ServerName string
	ServerID   string
}

func (u *UCHandler) GetUid(sdkUID string) (string, error) {
	info, err, _ := u.AuthDB.GetDeviceInfo(sdkUID, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", info.UserId), nil
}

func (u *UCHandler) GetPlayerShardInfo(uid string) ([]RoleShardInfo, error) {
	info, err := u.AuthDB.GetUserShardInfo(uid)
	if err != nil {
		return nil, err
	}
	logs.Debug("user shard info: %v", info)
	ret := make([]RoleShardInfo, 0)
	for _, item := range info {
		gsid := strings.Split(item.GidSid, ":")
		if len(gsid) < 2 {
			logs.Error("get usershardinfo ret value format error")
			continue
		}
		gid, err := strconv.Atoi(gsid[0])
		if err != nil {
			logs.Error("parse gid strconv atoi error by %v", err)
			continue
		}
		sid, err := strconv.Atoi(gsid[1])
		if err != nil {
			logs.Error("parse sid strconv atoi error by %v", err)
			continue
		}
		if sid == 1000 {
			logs.Info("pass channel server")
			continue
		}
		name := etcd.GetSidDisplayName(config.CommonConfig.EtcdRoot, uint(gid), uint(sid))
		acid := item.GidSid + ":" + uid
		logs.Debug("get role info, acid: %v, gid: %v, sid: %v", acid, gid, sid)
		roleName, roleLevel, err := u.getRoleInfo(acid, uint(gid), uint(sid))
		if err != nil {
			logs.Error("get role info err by %v", err)
			continue
		}
		ret = append(ret, RoleShardInfo{
			ServerName: name,
			ServerID:   item.GidSid,
			RoleName:   roleName,
			RoleLevel:  roleLevel,
			AcID:       acid,
		})
	}
	return ret, nil
}

func (u *UCHandler) getRoleInfo(acid string, gid, sid uint) (name string, level int, err error) {
	conn, err := core.GetGamexProfileRedisConn(gid, sid)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return "", 0, err
	}
	reply, err := conn.Do("HMGET", "profile:"+acid, "name", "corp")
	if err != nil {
		return "", 0, err
	}
	values, err := redis.Values(reply, nil)
	if err != nil {
		return "", 0, err
	}
	if len(values) < 2 {
		return "", 0, fmt.Errorf("redis.Value err by illegal format(len < 2)")
	}
	if values[0] == nil && values[1] == nil {
		return "", 0, fmt.Errorf("redis.Value err by nil return")

	}
	name, err = getNameFromJson(values[0])
	if err != nil {
		return "", 0, fmt.Errorf("getLvInfoFromJson err by %v ", err)
	}
	level, err = getLvFromJson(values[1])
	if err != nil {
		return "", 0, fmt.Errorf("getLvInfoFromJson err by %v ", err)
	}
	return name, level, nil
}

func (u *UCHandler) GetAllShardInfo(gid uint) []ShardInfo {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.allShardInfo
}

func (u *UCHandler) LoadAllShardInfoFromRemote(gid uint) error {
	info, err := models.GetAllShard(gid)
	if err != nil {
		return err
	}
	logs.Debug("get all shard info: %v", info)
	ret := make([]ShardInfo, 0)
	for _, item := range info {
		ret = append(ret, ShardInfo{
			ServerName: item.DName,
			ServerID:   fmt.Sprintf("%d:%d", item.Gid, item.Sid),
		})
	}
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.allShardInfo = ret
	return nil
}

func aesDecode(data string, secretKey string) ([]byte, error) {
	src := make([]byte, len(data))
	_, err := base64.StdEncoding.Decode(src, []byte(data))
	if err != nil {
		return nil, err
	}
	fmt.Println(src)
	trimByte := bytes.TrimRight(src, "\x00")
	fmt.Println(trimByte)
	keyBytes := []byte(secretKey)
	block, err := aes.NewCipher(keyBytes) //选择加密算法
	if err != nil {
		return nil, err
	}
	blockModel := cipher.NewCBCDecrypter(block, keyBytes)
	plantText := make([]byte, len(trimByte))
	blockModel.CryptBlocks(plantText, trimByte)
	return bytes.TrimRight(plantText, "\x00"), nil
}

func aesEncode(data []byte, secretKey string) (string, error) {
	block, err := aes.NewCipher([]byte(secretKey)) //选择加密算法
	if err != nil {
		return "", err
	}
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padtext := strings.Repeat("\x00", padding) //用0去填充
	src := append(data, []byte(padtext)...)
	blockModel := cipher.NewCBCEncrypter(block, []byte(secretKey))
	ciphertext := make([]byte, len(src))
	blockModel.CryptBlocks(ciphertext, src)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func GenSign(data string, apiKey string, caller string) (string, error) {
	content := caller + "params=" + data + apiKey
	logs.Debug("sign content: %v", content)
	_md5 := md5.New()
	_md5.Write([]byte(content))
	ret := _md5.Sum(nil)
	hexText := make([]byte, 32)
	hex.Encode(hexText, ret)
	return string(hexText), nil
}

func getLvFromJson(v interface{}) (int, error) {
	jsonStr, err := redis.String(v, nil)
	if err != nil {
		return 0, err

	}
	lv := struct {
		Level int `json:"lv"`
	}{}
	err = json.Unmarshal([]byte(jsonStr), &lv)
	if err != nil {
		return 0, err
	}
	return lv.Level, nil
}

func getNameFromJson(v interface{}) (string, error) {
	jsonStr, err := redis.String(v, nil)
	if err != nil {
		return "", err
	}
	return jsonStr, nil
}

func getVIPLvFromJson(v interface{}) (int, error) {
	jsonStr, err := redis.String(v, nil)
	if err != nil {
		return 0, err
	}

	lv := struct {
		Level int `json:"v"`
	}{}
	err = json.Unmarshal([]byte(jsonStr), &lv)
	if err != nil {
		return 0, err
	}
	return lv.Level, nil
}

func getHCCountFromJson(v interface{}) (int, error) {
	jsonStr, err := redis.String(v, nil)
	if err != nil {
		return 0, err
	}

	lv := struct {
		RmbPoint int `json:"rmb"`
	}{}
	logs.Debug("reply: %v", jsonStr)
	err = json.Unmarshal([]byte(jsonStr), &lv)
	if err != nil {
		return 0, err
	}
	return lv.RmbPoint, nil
}

func genGIDSID(gidsid string) (uint, uint, error) {
	gsid := strings.Split(gidsid, ":")
	if len(gsid) < 2 {
		return 0, 0, fmt.Errorf("get usershardinfo ret value format error")
	}
	gid, err := strconv.Atoi(gsid[0])
	if err != nil {
		return 0, 0, err
	}
	sid, err := strconv.Atoi(gsid[1])
	if err != nil {
		return 0, 0, err
	}
	return uint(gid), uint(sid), nil
}

func getGiftInfo(mapInfo map[string]interface{}) (*UCGiftInfo, error) {
	data, err := json.Marshal(mapInfo)
	if err != nil {
		return nil, err
	}
	var giftInfo UCGiftInfo
	err = json.Unmarshal(data, &giftInfo)
	if err != nil {
		return nil, err
	}
	return &giftInfo, nil
}

func getGiftMapInfo(info *UCGiftInfo) (map[string]interface{}, error) {
	data, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	var mapData map[string]interface{}
	err = json.Unmarshal(data, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}

func isSameDay(t1str, t2str string) bool {
	t1, err := strconv.ParseInt(t1str, 10, 0)
	if err != nil {
		// TODO by ljz handle error
		return true
	}
	t2, err := strconv.ParseInt(t2str, 10, 0)
	if err != nil {
		// TODO by ljz handle error
		return true
	}
	time1 := time.Unix(t1, 0)
	time2 := time.Unix(t2, 0)
	return time1.Year() == time2.Year() && time1.Month() == time2.Month() && time1.Day() == time2.Day()
}

func isSameDay2(t1str, t2str string) bool {
	t1, err := time.Parse("2006-01-02", t1str)
	if err != nil {
		logs.Error("time parse err by %v", err)
		return false
	}
	t2, err := time.Parse("2006-01-02", t2str)
	if err != nil {
		logs.Error("time parse err by %v", err)
		return false
	}
	return t1 == t2
}

func isSameWeekDay(t1str, t2str string) bool {
	t1, err := strconv.ParseInt(t1str, 10, 0)
	if err != nil {
		// TODO by ljz handle error
		return true
	}
	t2, err := strconv.ParseInt(t2str, 10, 0)
	if err != nil {
		// TODO by ljz handle error
		return true
	}
	year1, week1 := time.Unix(t1, 0).ISOWeek()
	year2, week2 := time.Unix(t2, 0).ISOWeek()
	logs.Debug("year1: %d, week1: %d, year2: %d, week2: %d", year1, week1, year2, week2)
	return year1 == year2 && week1 == week2
}
