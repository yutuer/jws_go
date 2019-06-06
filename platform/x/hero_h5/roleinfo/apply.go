package roleinfo

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/utils"
	"github.com/cenk/backoff"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/herogacharace"
	"vcs.taiyouxi.net/jws/gamex/modules/ws_pvp"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
	h5Config "vcs.taiyouxi.net/platform/x/hero_h5/config"
	"vcs.taiyouxi.net/platform/x/hero_h5/serverinfo"
)

/*测试base.go中的函数*/
type severgroup struct {
	// * 服务器ID
	SID uint32 `protobuf:"varint,1,req,def=0" json:"SID,omitempty"`
	// * 跨服同组（从1开始）
	GroupID uint32 `protobuf:"varint,2,opt,def=0" json:"GroupID,omitempty"`
	// * 无双竞技场服务器分组
	WspvpGroupID uint32 `protobuf:"varint,5,opt,def=0" json:"WspvpGroupID,omitempty"`
	// * 无双竞技场机器人配置
	WspvpBot uint32 `protobuf:"varint,6,opt,def=0" json:"WspvpBot,omitempty"`
	// * 劫营夺粮服务器分组
	RobCropsGroupID uint32 `protobuf:"varint,7,opt,def=0" json:"RobCropsGroupID,omitempty"`
	// * 服务器付费批次
	Sbatch uint32 `protobuf:"varint,3,opt,def=0" json:"Sbatch,omitempty"`
	// * 批次生效时间
	EffectiveTime string `protobuf:"bytes,4,opt,def=" json:"EffectiveTime,omitempty"`
	// * 世界boss服务器分组
	WorldBossGroupID uint32 `protobuf:"varint,8,opt,def=0" json:"WorldBossGroupID,omitempty"`
	// * 单服限时神将配置
	HGRHotID         uint32 `protobuf:"varint,9,opt,def=0" json:"HGRHotID,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

//策划的通关次数表
const (
	releation0  = 148
	releation1  = 168
	releation2  = 178
	releation3  = 188
	releation4  = 198
	releation5  = 218
	releation6  = 238
	releation7  = 248
	releation8  = 278
	releation9  = 308
	releation10 = 348
	releation11 = 358
	releation12 = 388
	releation13 = 418
	releation14 = 448
	releation15 = 528
	releation16 = 558
)

type accountInfo struct {
	//建号时间
	CreateTime string `json:"CreateTime"`
	//玩家姓名
	Name string `json:"name"`
	//服务器名字
	SeverName string `json:"severname"`
	//第一个英雄
	FirstHero string `json:"firstHero"`
	//战力最高的英雄
	BestHero string `json:"bestHero"`
	//英雄数量
	HeroNum uint `json:"heroNum"`
	//工会名字
	Gname string `json:"gname"`
	//玩家等级
	Level uint `json:"level"`
	//玩家战役数量
	BattleNum uint `json:"battleNum"`
	//个人排行榜排名
	PvpRank int `json:"pvpRank"`
	//跨服排行榜排名
	A_pvpRank int `json:"A_pvpRank"`
	Result    int `json:"result"`
}

//获取玩家的ACID,若未找到则报错
func getACID(name string, gid, sid int) (string, error) {
	_realSid, err := etcd.Get(fmt.Sprintf("%s/%d/%d/gm/mergedshard", h5Config.CommonConfig.Etcd_Root, gid, sid))
	if err != nil {
		logs.Error("get realSid err :%v", err)
		return "", err
	}
	realSid, err := strconv.Atoi(_realSid)
	if err != nil {
		logs.Error("change realSid from string to int err :%v", err)
		return "", err
	}
	acID, err := getACIDFromName(name, gid, sid, realSid)
	if err != nil {
		logs.Error("getAcIDFromName err :%v", err)
		return "", err
	}
	return acID, nil
}

//获取信息，通过acid将玩家所有信息查找出来，放在结构体内
func getInfo(acid string, gid, sid int) (accountInfo, error) {
	profileID := "profile:" + acid
	var result accountInfo = accountInfo{}
	//链接玩家个人信息数据库db
	conn1, err := core.GetGamexProfileRedisConn(uint(gid), uint(sid))
	defer func() {
		if conn1 != nil {
			conn1.Close()
		}
	}()
	if err != nil {
		return result, err
	}
	//链接玩家排行榜信息数据库db
	conn5, err := core.GetGamexRankReidsConn(uint(gid), uint(sid))
	defer func() {
		if conn5 != nil {
			conn5.Close()
		}
	}()
	if err != nil {
		return result, err
	}
	//获取建号时间
	timestamp, err := redis.Int64(conn1.Do("HGET", profileID, "createtime"))
	if err != nil {
		return result, err
	}
	result.CreateTime = time.Unix(timestamp, 0).Format("2006年01月02日")

	//玩家昵称
	name, err := redis.String(conn1.Do("HGET", profileID, "name"))
	if err != nil {
		return result, err
	}
	result.Name = name
	//服务器名
	severname, err := getEtcdCfg1(uint(gid), uint(sid))
	if err != nil {
		return result, err
	} else {
		result.SeverName = severname
	}
	//第一个英雄
	result.FirstHero = "关平"
	//战力最高英雄
	besthero, err := conn1.Do("HGET", profileID, "data")
	if err != nil {
		return result, err
	} else if besthero != nil {
		var herodata account.ProfileData
		err := json.Unmarshal(besthero.([]byte), &herodata)
		if err != nil {
			return result, err
		} else {
			max_gs := 0     //最高战斗力武将的战斗力
			index_gs := 0   //最高战斗力武将id
			s_index_gs := 0 //第二高战斗力武将的id
			for i := 0; i < helper.AVATAR_NUM_MAX; i++ {
				if herodata.HeroGs[i] > max_gs {
					s_index_gs = index_gs
					max_gs = herodata.HeroGs[i]
					index_gs = i
				}
			}
			if index_gs == serverinfo.HERO_GY { //如果战斗力最高的英雄为关平，则输出第二战斗力武将
				index_gs = s_index_gs

			}

			result.BestHero = serverinfo.GetHeroName(index_gs)
		}
	}
	//总英雄数量
	heros, err := conn1.Do("HGET", profileID, "hero")
	if err != nil {
		return result, err
	} else if heros != nil {

		//武将数量
		var herodata account.PlayerHero

		err := json.Unmarshal(heros.([]byte), &herodata)
		if err != nil {
			return result, err
		} else {
			heronum := 0
			for i := 0; i < helper.AVATAR_NUM_MAX; i++ {
				if herodata.HeroStarLevel[i] > 0 {
					heronum++
				}
			}

			result.HeroNum = uint(heronum)
		}
	}
	//工会名
	gname, err := redis.String(conn1.Do("HGET", "pguild:"+acid, "gname"))
	if err != nil {
		logs.Error("get gname err :%v", err)
		result.Gname = ""
	} else {
		result.Gname = gname
	}
	//玩家等级

	corpdata, err := conn1.Do("HGET", profileID, "corp")
	if err != nil {
		return result, err
	} else {
		var corp account.Corp
		err = json.Unmarshal(corpdata.([]byte), &corp)
		if err != nil {
			return result, err
		}

		result.Level = uint(corp.Level)
	}

	//计算游戏天数
	now := time.Now().Unix()
	ct := timestamp
	if err != nil {
		return result, err
	}
	pro_days := (now-int64(ct))/(24*3600) + 1
	//玩家vip信息
	vipdata, err := conn1.Do("HGET", profileID, "v")
	if err != nil {
		return result, err
	}
	var vip account.VIP
	err = json.Unmarshal(vipdata.([]byte), &vip)

	if err != nil {
		return result, err
	} else {
		//计算通关数量
		valueVip := 0
		switch vip.V {
		case 0:
			valueVip = releation0
		case 1:
			valueVip = releation1
		case 2:
			valueVip = releation2
		case 3:
			valueVip = releation3
		case 4:
			valueVip = releation4
		case 5:
			valueVip = releation5
		case 6:
			valueVip = releation6
		case 7:
			valueVip = releation7
		case 8:
			valueVip = releation8
		case 9:
			valueVip = releation9
		case 10:
			valueVip = releation10
		case 11:
			valueVip = releation11
		case 12:
			valueVip = releation12
		case 13:
			valueVip = releation13
		case 14:
			valueVip = releation14
		case 15:
			valueVip = releation15
		case 16:
			valueVip = releation16
		}
		result.BattleNum = uint(pro_days * int64(valueVip))
	}
	//个人竞技排行榜

	pvprank, err := redis.Int(conn5.Do("zrevrank", fmt.Sprintf("%d:%d:RankSimplePvp", gid, sid), acid))
	logs.Debug("info: %v, %v", fmt.Sprintf("%d:%d", gid, sid)+":RankSimplePvp", acid)
	pvprank++
	if err != nil {
		logs.Error("get pvprank err by %v", err)
		result.PvpRank = -1
	} else {
		result.PvpRank = pvprank
	}

	result.A_pvpRank, err = getwspvpinfo(acid, sid, gid)
	logs.Error("跨服排行榜获取出错 %v", err)
	logs.Debug("map长度为：%v", len(gdServerGroupSbatch))
	return result, nil

}

func getwspvpinfo(acid string, sid, gid int) (int, error) {
	//跨服争霸排行榜
	groupString := (fmt.Sprintf("%d", gdServerGroupSbatch[uint32(sid)].WspvpGroupID))
	ws_pvpmap, err := getEtcdCfg(uint(gid))
	if err != nil {
		logs.Error("跨服排行榜获取出错 %v", err)
		return -1, err
	}
	logs.Debug("ws_pvpmap_length,groupString, %d,%s", len(ws_pvpmap), groupString)
	for k, v := range ws_pvpmap {
		logs.Debug("ws_pvpmap,key:%v,value:%v", k, v)
	}
	wsConfig, ok := ws_pvpmap[groupString]
	logs.Debug("ok:%v", ok)
	if !ok {
		wsConfig = ws_pvpmap["default"]
	}
	logs.Debug("DB:%v", wsConfig.DB)
	connx, err := getRedisConn(wsConfig.AddrPort, wsConfig.DB, wsConfig.Auth)
	defer func() {
		if connx != nil {
			connx.Close()
		}
	}()
	if err != nil {
		logs.Error("跨服排行榜获取出错 %v", err)
		return -1, err
	}
	logs.Debug("groupString:%s", groupString)
	A_rankpvp, err := redis.Int(connx.Do("zrank", "rank:"+groupString, acid))
	if err != nil {
		logs.Error("跨服排行榜获取出错 %v", err)
		return -1, err
	} else {
		return A_rankpvp + 1, nil
	}
}

func Init(e *gin.Engine) {
	config.CommonConfig.EtcdRoot = h5Config.CommonConfig.Etcd_Root
	//加载data文件
	dataRelPath := "data"
	dataAbsPath := filepath.Join(getMyDataPath(), dataRelPath)
	load := func(dfilepath string, loadfunc func(string)) {
		loadfunc(filepath.Join("", dataAbsPath, dfilepath))
		logs.Trace("LoadGameData %s success", dfilepath)
	}
	load("level_info.data", loadStageData)
	//加载之后，全局变量gdStageData中为全部已通关关卡的详细信息
	//全局变量gdStagePurchasePolicy中为所有已通关关卡的关卡重置次数购买规则
	//全局变量gdStageTimeLimit中为限时
	load("severgroup.data", loadServeGroup)
	e.GET("/getRoleInfo", getRoleInfo)
}

func getRoleInfo(c *gin.Context) {

	name := c.Query("name")

	serverID := c.Query("server_id")

	logs.Debug("server_id为:%s", serverID)
	g_sid := strings.Split(serverID, ":")

	if g_sid[0] == "" {
		logs.Debug("putin wrong gid")
		c.JSON(202, accountInfo{Result: -1})
		return
	}
	gid, err := strconv.Atoi(g_sid[0])
	if err != nil {
		logs.Error("putin wrong gid error %v", err)
	}
	if g_sid[1] == "" {
		logs.Debug("putin wrong sid")
		c.JSON(202, accountInfo{Result: -1})
		return
	}
	sid, err := strconv.Atoi(g_sid[1])
	if err != nil {
		logs.Error("putin wrong sid error %v", err)
	}
	acid, err := getACID(name, gid, sid)
	if err != nil {
		logs.Error("get ACID by name err: %v", err)
		c.JSON(202, accountInfo{Result: -1})
		return
	}
	info, err := getInfo(acid, gid, sid)
	if err != nil {
		logs.Error("get Info err by: %v", err)
		c.JSON(201, accountInfo{Result: -1})
		return
	}
	logs.Debug("%v", info)
	info.Result = 1
	c.JSON(200, info)

}

var (
	gdStageData           map[string]gamedata.StageData
	gdStagePurchasePolicy map[string]string
	gdStageTimeLimit      map[int32]int32
	gdServerGroupSbatch   map[uint32]severgroup
)

func getEtcdCfg(gid uint) (map[string]ws_pvp.RedisDBSetting, error) {
	var etcdMap map[string]ws_pvp.RedisDBSetting
	if jsonValue, err := etcd.Get(fmt.Sprintf("%s/%d/WSPVP/dbs", h5Config.CommonConfig.Etcd_Root, gid)); err != nil {
		return nil, fmt.Errorf("wspvp etcd get key failed. %v", err)
	} else {
		if err := json.Unmarshal([]byte(jsonValue), &etcdMap); err != nil {
			return nil, fmt.Errorf("wspvp json.Unmarshal key failed. %v", err)
		}
	}
	return etcdMap, nil
}
func getEtcdCfg1(gid, sid uint) (string, error) {
	var etcdString string
	if jsonValue, err := etcd.Get(fmt.Sprintf("%s/%d/%d/dn", h5Config.CommonConfig.Etcd_Root, gid, sid)); err != nil {
		return "", fmt.Errorf("etcd get key failed. %v", err)
	} else {
		return jsonValue, nil
	}
	return etcdString, nil
}

//用于解析servergroup.data文件
func loadServeGroup(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.SEVERGROUP_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	as := ar.GetItems()
	gdServerGroupSbatch = make(map[uint32]severgroup, len(as))
	for _, r := range as {
		gdServerGroupSbatch[r.GetSID()] = severgroup{
			SID:              r.GetSID(),
			GroupID:          r.GetGroupID(),
			WspvpGroupID:     r.GetWspvpGroupID(),
			WspvpBot:         r.GetWspvpBot(),
			RobCropsGroupID:  r.GetRobCropsGroupID(),
			Sbatch:           r.GetSbatch(),
			EffectiveTime:    r.GetEffectiveTime(),
			WorldBossGroupID: r.GetWorldBossGroupID(),
			HGRHotID:         r.GetHGRHotID(),
		}

	}
}

//用于解析level_info.data文件
func loadStageData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	//将.data文件解析为[]byte
	buffer, err := loadBin(filepath)
	errcheck(err)
	//将解析出的[]byte传给stage_data
	stage_data := &ProtobufGen.LEVEL_INFO_ARRAY{}
	err = proto.Unmarshal(buffer, stage_data)
	errcheck(err)
	//获取玩家等级信息
	datas := stage_data.GetItems()
	gdStageData = make(map[string]gamedata.StageData, len(datas))
	gdStagePurchasePolicy = make(map[string]string)
	for _, r := range datas {
		pre_lv := strings.Split(r.GetPreLevelID(), ",")
		gdStageData[r.GetLevelID()] = gamedata.StageData{
			r.GetLevelID(),
			r.GetChapterID(),
			r.GetLevelType(),
			r.GetEnergy(),
			r.GetHighEnergy(),
			r.GetMaxDailyAccess(),
			r.GetTimeLimit(),
			r.GetTimeGoal(),
			r.GetHpGoal(),
			pre_lv,
			r.GetTeamRequirement(),
			r.GetLevelRequirement(),
			r.GetRoleOnly(),
			uint32(r.GetGameModeID()),
			r.GetLevelIndex(),
			r.GetEliteLevelIndex(),
			r.GetHardLevelIndex(),
		}
		if r.GetLevelPurchasePolicy() != "" {
			gdStagePurchasePolicy[r.GetLevelID()] = r.GetLevelPurchasePolicy()
		}
	}
	gdStageTimeLimit = make(map[int32]int32, 0)
	for _, r := range datas {
		gdStageTimeLimit[r.GetLevelType()] = r.GetTimeLimit()
	}
}

func loadBin(cfgname string) ([]byte, error) {
	errgen := func(err error, extra string) error {
		return fmt.Errorf("gamex.models.gamedata loadbin Error, %s, %s", extra, err.Error())
	}

	//	path := GetDataPath()
	//	appConfigPath := filepath.Join(path, cfgname)

	file, err := os.Open(cfgname)
	if err != nil {
		return nil, errgen(err, "open")
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, errgen(err, "stat")
	}

	buffer := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buffer) //read all content
	if err != nil {
		return nil, errgen(err, "readfull")
	}

	return buffer, nil
}
func getRedisConn(redisUrl string, redisDB int, redisAuth string) (redis.Conn, error) {
	logs.Debug("redis conn nil for param: %v, %v, %v", redisUrl, redisAuth, redisDB)
	var redisConn redis.Conn
	err := backoff.Retry(func() error {
		conn, err := redis.Dial("tcp", redisUrl,
			redis.DialConnectTimeout(5*time.Second),
			redis.DialReadTimeout(5*time.Second),
			redis.DialWriteTimeout(5*time.Second),
			redis.DialPassword(redisAuth),
			redis.DialDatabase(redisDB),
		)
		if conn != nil {
			redisConn = conn
		}
		return err
	}, herogacharace.New2SecBackOff())

	if err != nil {
		return nil, err
	}
	if redisConn == nil {
		return nil, fmt.Errorf("redis conn nil for param: %v, %v, %v", redisUrl, redisAuth, redisDB)
	}
	return redisConn, nil
}

//自用获取路径函数
func getMyDataPath() string {
	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	appConfigPath := filepath.Join(AppPath, "conf")
	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "conf")
		}
	}
	return appConfigPath
}
func getACIDFromName(name string, gid, sid int, realSid int) (string, error) {
	conn, err := core.GetGamexProfileRedisConn(uint(gid), uint(sid))
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		logs.Error("getACIDFromName err when get conn:%v", err)
		return "", err
	}
	acID, err := redis.String(conn.Do("HGET", fmt.Sprintf("names:%d:%d", gid, realSid), name))
	if err != nil {
		logs.Error("getACIDFromName err when get acid:%v", err)
		return "", err
	}
	if strings.Split(acID, ":")[1] != strconv.Itoa(sid) {
		logs.Error("sid illegal for numberID: %d, gid: %d, sid: %d, acID: %s", acID)
		return "", fmt.Errorf("sid illegal for numberID: %d, gid: %d, sid: %d, acID: %s", acID)
	}
	return acID, nil
}
