package guild_info

import (
	"fmt"
	"time"

	"strconv"
	"strings"

	"encoding/json"

	"github.com/astaxie/beego/cache"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/guild_boss"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/planx/util/safecache"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

const (
	MaxGuildMember       = helper.MaxGuildMember
	MaxGuildScienceCount = helper.MaxGuildScienceCount

	Table_Guild = "guild:Info" // 公会信息表
)

var (
	guildCache cache.Cache
)

func init() {
	_guildCache, err := safecache.NewSafeCache("guild_cache", `{"internal:"300"}`)
	if err != nil {
		logs.Critical("GuildInfo Startup Failed, due to Cache: %s", err.Error())
		panic("GuildInfo Startup Failed, by lockCache init failed")
	}
	guildCache = _guildCache
}

// 从数据库加载一个公会的信息,主要用于排行榜,只加载,不存储
// 优化：不加载整个工会，只加载simpleinfo。因为频繁加载导致性能下降严重
func LoadGuildInfo(guildUUID string, db redispool.RedisPoolConn) *GuildSimpleInfo {
	_g := guildCache.Get(guildUUID)
	if _g != nil {
		g, ok := _g.(GuildSimpleInfo)
		if ok {
			return &g
		} else {
			logs.Error("LoadGuildInfo convert to GuildSimpleInfo fail")
		}
	}

	s, err := redis.String(db.Do("HGET", GuildDBName(guildUUID), "Base"))
	if err != nil || s == "" {
		logs.Error("LoadGuildInfo HGET Base err %s or res %s empty  %s",
			err.Error(), s, guildUUID)
		return nil
	}
	si := &GuildSimpleInfo{}
	if err := json.Unmarshal([]byte(s), si); err != nil {
		logs.Error("LoadGuildInfo json.Unmarshal Base err %s", err.Error())
		return nil
	}

	// 当是工会老档，找出会长，补充在simpleinfo里
	if si.LeaderAcid == "" {
		sm, err := redis.String(db.Do("HGET", GuildDBName(guildUUID), "Members"))
		if err != nil || sm == "" {
			logs.Error("LoadGuildInfo HGET Members err %s or res %s empty", err.Error(), sm)
			return nil
		}

		mems := &[MaxGuildMember]helper.AccountSimpleInfo{}
		if err := json.Unmarshal([]byte(sm), mems); err != nil {
			logs.Error("LoadGuildInfo json.Unmarshal Members err %s", err.Error())
			return nil
		}
		for _, m := range mems {
			if m.GuildPosition == gamedata.Guild_Pos_Chief {
				si.LeaderAcid = m.AccountID
				si.LeaderName = m.Name
				break
			}
		}
		logs.Debug("LoadGuildInfo frommem %v", si)
	}

	// 存缓存
	if err := guildCache.Put(guildUUID, *si, 300); err != nil {
		logs.Error("guildCache.Put err %s", err.Error())
	}
	return si
}

type GuildSimpleInfo struct {
	GuildUUID                 string `json:"uuid"`
	GuildID                   int64  `json:"gid"`
	Name                      string `json:"name"`
	LeaderAcid                string `json:"leader"`
	LeaderName                string `json:"leader_n"`
	CreateTS                  int64  `json:"create_ts"`
	Level                     uint32 `json:"level"`
	MemNum                    int    `json:"mem_num"`
	MaxMemNum                 int    `json:"max_mem_num"`
	ApplyGsLimit              int    `json:"aply_gs"`   // 申请公会gs限制
	ApplyAuto                 bool   `json:"aply_auto"` // 申请是否自动加入
	XpCurr                    int64  `json:"actxp"`     // 当前活跃度 升级用
	GEPointWeek               int64  `json:"pweek"`
	GEPointWeekUpdateTime     int64  `json:"pweekt"`
	Icon                      string `json:"icon"`
	Notice                    string `json:"notc"`
	JoinType                  int    `json:"join_t"`
	GatesEnemyCount           int    `json:"gec"`
	GatesEnemyCountUpdateTime int64  `json:"gect"`
	isNowInGEAct              bool
	GuildGSSum                int64 `json:"gssum"`
	RenameTimes               int   `json:"renamet"` // 合服后清0
	GuildTmpVer               int   `json:"guildtmpver"`
	Guild2LvlTimes            int64 `json:"guild_2_lvl_times"`
}

func (g *GuildSimpleInfo) GetGEPointWeek() int64 {
	nowT := time.Now().Unix()
	if util.IsSameUnixByStartTime(
		nowT, g.GEPointWeekUpdateTime,
		gamedata.GetGatesEnemyWeekBalanceTime()) {
		g.GEPointWeek = 0
		g.GEPointWeekUpdateTime = nowT
	}

	return g.GEPointWeek
}

func (g *GuildSimpleInfo) AddGEPointWeek(p int64) {
	nowT := time.Now().Unix()
	if util.IsSameUnixByStartTime(
		nowT, g.GEPointWeekUpdateTime,
		gamedata.GetGatesEnemyWeekBalanceTime()) {
		g.GEPointWeek = 0
		g.GEPointWeekUpdateTime = nowT
	}

	g.GEPointWeek += p
}

func (g *GuildSimpleInfo) GetGateEnemyCount() int {
	nowT := time.Now().Unix()
	if !gamedata.IsSameDayCommon(nowT, g.GatesEnemyCountUpdateTime) {
		g.GatesEnemyCount = int(gamedata.GetGEConfig().GetDailyFightTime())
		g.GatesEnemyCountUpdateTime = nowT
	}

	return g.GatesEnemyCount
}

func (g *GuildSimpleInfo) DebugResetGateEnemyCount() {
	g.GatesEnemyCount = int(gamedata.GetGEConfig().GetDailyFightTime())
}

func (g *GuildSimpleInfo) SetInGE(b bool) {
	g.isNowInGEAct = b
}

func (g *GuildSimpleInfo) GetInGE() bool {
	return g.isNowInGEAct
}

type GuildScience struct {
	Lvl uint32 `json:"lvl" codec:"lvl"`
	Exp uint32 `json:"exp" codec:"exp"`
}

// 工会换团长的信息
type GuildChangeLeader struct {
	IsSleep        bool  `json:"issleep" codec:"issleep"`               // 沉睡状态
	LastUpdateTime int64 `json:"lastupdatetime" codec:"lastupdatetime"` // 上次状态发生改变的时间, 只是用于唤醒24小时的判断
}

type GuildInfoBase struct {
	Base                  GuildSimpleInfo                          `json:"base"`
	Members               [MaxGuildMember]helper.AccountSimpleInfo `json:"members"`  // MemNum: 成员实际数量在Base中
	Sciences              [MaxGuildScienceCount]GuildScience       `json:"sciences"` // 科技树，0号无用
	GatesEnemyData        player_msg.GatesEnemyData                `json:"gedata"`   // 用来在兵临城下活动后发奖和看排行的数据
	Inventory             GuildInventory                           `json:"inventory"`
	ActBoss               guild_boss.ActivityState                 `json:"actboss"`
	GSTTodayResetTime     int64                                    `json:"gsttdrt"`
	GSTWeekResetTime      int64                                    `json:"gstwkrt"`
	DebugTimeAbsolute     int64                                    `json:"_dt"`
	ActivePlayerStatistic logiclog.DailyStatistics                 `json:"stic"` // 为bi信息的加的统计信息
	GuildChangeChief      GuildChangeLeader                        `json:"chleader"`
	GuildRedPacket        GuildRedPacketInfo                       `json:"grpi"`
	GuildLog              GuildLogs                                `json:"glog"`
	GuildWorship          GuildWorshipInfo                         `json:"g_worship"`
	LostInventory         GuildInventory                           `json:"lost_inventory"`
}

func (g *GuildInfoBase) TryRefresh(shardId uint) {
	g.ActBoss.Refersh(g.Base.GuildUUID, g.Base.Name, g.GetDebugNowTime(shardId))
	g.GuildRedPacket.CheckDailyReset(g.GetDebugNowTime(shardId))
	g.TryResetGuildWorship(g.GetDebugNowTime(shardId))
}

func (guildInfo *GuildInfoBase) GetGuildChief() *helper.AccountSimpleInfo {
	for i := 0; i < int(guildInfo.Base.MemNum); i++ {
		if guildInfo.Members[i].GuildPosition == gamedata.Guild_Pos_Chief {
			return &guildInfo.Members[i]
		}
	}
	return nil
}

func (p *GuildInfoBase) GetMember(acID string) *helper.AccountSimpleInfo {
	if acID == "" {
		return nil
	}

	for i := 0; i < len(p.Members) && i < p.Base.MemNum; i++ {
		if p.Members[i].AccountID == acID {
			return &p.Members[i]
		}
	}

	return nil
}

func GuildDBName(guuid string) string {
	return fmt.Sprintf("%s:%s", Table_Guild, guuid)
}

func (p *GuildInfoBase) DBSave(cb redis.CmdBuffer) error {
	logs.Debug("GuildInfoBase DBSave")
	key := GuildDBName(p.Base.GuildUUID)
	return driver.DumpToHashDBCmcBuffer(cb, key, p)
}

func (p *GuildInfoBase) DBLoad(logInfo bool) error {
	logs.Debug("GuildInfoBase DBLoad")
	key := GuildDBName(p.Base.GuildUUID)

	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return err
}

func (g *GuildInfoBase) GetDebugNowTime(shard uint) int64 {
	return g.DebugTimeAbsolute + game.GetNowTimeByOpenServer(shard)
}

func (g *GuildInfoBase) SetDebugTime(t int64, shard uint) {
	g.DebugTimeAbsolute = t - game.GetNowTimeByOpenServer(shard)
}

func GenGuildUUidByPlayer(gid, sid uint) string {
	return fmt.Sprintf("%d:%d:%s", gid, sid, uuid.NewV4().String())
}

func GetShardIdByGuild(guildId string) (uint, error) {
	infos := strings.Split(guildId, ":")
	if len(infos) != 3 {
		return 0, fmt.Errorf("GuildId format ilegal, Parse %s failed!", guildId)
	}
	id, err := strconv.Atoi(infos[1])
	if err != nil {
		return 0, fmt.Errorf("GuildId format ilegal, Parse %s failed!", guildId)
	}
	return uint(id), nil
}

func CheckGuildUuid(guuid string) bool {
	_, err := db.ParseAccount(guuid)
	return err == nil
}
