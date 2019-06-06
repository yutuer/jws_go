package ws_pvp

import (
	"encoding/json"
	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type WSPVPInfo struct {
	// 角色基础属性
	Acid      string        `redis:"acid"`  // account id
	Name      string        `redis:"name"`  // name
	AvatarId  int           `redis:"aid"`   // 当前使用形象ID
	CorpLv    uint32        `redis:"clv"`   // 战队等级
	GuildName string        `redis:"gname"` // 公会名字
	ServerId  int           `redis:"sid"`
	VipLevel  int           `redis:"vlv"`
	TitleId   string        `redis:"title_id"`
	AllGs     int64         `redis:"allgs"`
	ExtraAttr ExtraAttr     `redis:"extra_attr"`
	BestHeros BestHeroArray `redis:"best_hero"`
	// 9个武将信息
	Formation WSPVPFormationInfo `redis:"formation"` // 布阵信息
}

type ExtraAttr struct {
	EquipAttr   [3]float32 `json:"equip_attr"`
	DestinyAttr [3]float32 `json:"des_attr"`
	JadeAttr    [3]float32 `json:"jade_attr"`
}

type BestHeroArray struct {
	BestHeroInfo []BestHeroInfo
}

type BestHeroInfo struct {
	Idx       int64 `json:"idx"`
	Level     int   `json:"lv"`
	StarLevel int   `json:star_lv`
	BaseGs    int64 `json:base_gs`
	ExtraGs   int64 `json:extra_gs`
}

type WSPVPFormationInfo struct {
	Avatar    []WSPVPHeroInfo `json:"avatar"` // 9个武将的战斗属性, 下标并不是一一对应关系
	Group     []int64         `json:"group"`  // 9个位置的武将ID 如果更改阵型产生新武将需要同步attr
	DesSkills []int64         `json:"des_skills"`
}

type WSPVPHeroInfo struct {
	Idx       int    // id
	Attr      []byte // 战斗属性
	StarLevel int    // 星级
	Gs        int64  // 战力

	AvatarSkills   []uint32 `json:"skills"`         // 武将技能等级
	AvatarFashion  []string `json:"avatar_equips"`  // 武将时装
	HeroSwing      int      `json:"swing"`          // 翅膀
	MagicPetfigure uint32   `json:"magicpetfigure"` // 灵宠形象
	PassiveSkillId []string `json:"pskillid"`       // 被动技能
	CounterSkillId []string `json:"cskillid"`
	TriggerSkillId []string `json:"tskillid"`
}

type WSPVPBattleLog struct {
	Attack            bool   `json:"att"`            // 是否是进攻方
	Result            bool   `json:"result"`         // 己方是否胜利
	RankChange        int64  `json:"rank_chg"`       // 排名变化值
	OpponentName      string `json:"opp_name"`       // 对手名字
	OpponentGuildName string `json:"opp_guild_name"` // 对手公会名字
	Time              int64  `json:"time"`           // 挑战时间
	Rank              int    `json:"rank"`
}

type WspvpLogArray []*WSPVPBattleLog

func (wa WspvpLogArray) Len() int {
	return len(wa)
}

func (wa WspvpLogArray) Less(i, j int) bool {
	return wa[i].Time < wa[j].Time
}

func (wa WspvpLogArray) Swap(i, j int) {
	wa[i], wa[j] = wa[j], wa[i]
}

type WSPVPRobotInfo struct {
	Name     string `json:"name"`
	ServerId int    `json:"sid"`
}

type WSPVPOppSimple struct {
	Acid      string `redis:"acid"`  // account id
	Name      string `redis:"name"`  // name
	GuildName string `redis:"gname"` // 公会名字
	ServerId  int    `redis:"sid"`
	TitleId   string `redis:"title_id"`
	VipLevel  int    `redis:"vlv"`
}

type dbCmdBuffPrepare func(cb redis.CmdBuffer) error

func dbCmdBuffExec(prepare dbCmdBuffPrepare) (interface{}, int) {
	cb := redis.NewCmdBuffer()

	if err := prepare(cb); err != nil {
		logs.Error("WSPVP dbCmdBuffExec err %v", err)
		return nil, errCode.WSPVPDBError
	}
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return nil, errCode.WSPVPDBError
	}
	if reply, err := modules.DoCmdBufferWrapper(WS_PVP_DB, db, cb, true); err != nil {
		logs.Error("DoCmdBuffer error %s", err.Error())
		return nil, errCode.WSPVPDBError
	} else {
		return reply, 0
	}
}

func dbHMGETExec(prepare dbCmdBuffPrepare) (interface{}, int) {
	reply, errC := dbCmdBuffExec(prepare)
	if errC != 0 {
		return nil, errC
	}
	realReply := make([]interface{}, 0)
	replyArray, err := redis.Values(reply, nil)
	if err != nil {
		logs.Error("err %v", err)
		return nil, errCode.WSPVPDBError
	}
	for _, it := range replyArray {
		itArray, err := redis.Values(it, nil)
		if err != nil {
			logs.Error("err %v", err)
			return nil, errCode.WSPVPDBError
		}
		realReply = append(realReply, itArray...)
	}
	return realReply, 0
}

// 存阵型
func SaveFormation2DB(groupId int, acid string, formation WSPVPFormationInfo) {
	wsJson, err := json.Marshal(formation)
	if err != nil {
		logs.Error("json marshal error, %v", err)
		return
	}
	tableName := getPersonalTableName(groupId, acid)
	_, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		return cb.Send("HSET", tableName, "formation", wsJson)
	})
	if errCode != 0 {
		logs.Error("save formation err, %d", errCode)
	}
}

// 获取指定角色的排名 从1开始
func GetRanks(groupId int, acids []string) []int {
	tableName := getRankTableName(groupId)
	reply, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for _, acid := range acids {
			err := cb.Send("ZRANK", tableName, acid)
			if err != nil {
				return err
			}
		}
		return nil
	})
	logs.Debug("get ranks src %v", reply)
	if errCode != 0 {
		logs.Error("get ranks err, %d", errCode)
		return nil
	} else {
		newRank, err := IntsWithDefault(reply, -1)
		if err == nil {
			if len(newRank) != len(acids) {
				logs.Error("check ranks result err, %d", errCode)
				return nil
			}
			for i := 0; i < len(newRank); i++ {
				newRank[i]++
			}
			return newRank
		} else {
			logs.Error("parse ranks err, %v", err)
			// 如果不能解析10000名以外的，说明这个名次肯定发生了变化， 返回Nil
			return nil
		}
	}
}

func IntsWithDefault(reply interface{}, defaultIfNil int64) ([]int, error) {
	var ints []int
	values, err := redis.Values(reply, nil)
	if err != nil {
		return ints, err
	}
	for i, value := range values {
		if value == nil {
			values[i] = defaultIfNil
		}
	}
	if err := redis.ScanSlice(values, &ints); err != nil {
		return ints, err
	}
	return ints, nil
}

// 根据排名获取相关对手的简易信息
// TODO 是否需要lua， 事务？
func GetSimpleOppInfo(groupId int, rankList []int) []*WSPVPOppSimple {
	acids := getAcidsByRank(groupId, rankList)
	if len(acids) != len(rankList) {
		logs.Error("get opp info err , size not matched")
		return nil
	}
	return getSimpleInfoByAcids(groupId, acids)
}

// 内存中的排名是从1到10000， redis是从0到9999
func getAcidsByRank(groupId int, rankList []int) []string {
	rankTableName := getRankTableName(groupId)
	reply, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for _, rank := range rankList {
			err := cb.Send("ZRANGE", rankTableName, rank-1, rank-1)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if errCode != 0 {
		logs.Error("GetSimpleOppInfo rank : %d", errCode)
		return nil
	}
	logs.Debug("getAcidsByRank %v", reply)
	acids := make([]string, 0)
	resArray, err := redis.Values(reply, nil)
	if err != nil {
		logs.Error("GetSimpleOppInfo acids: %v", err)
		return nil
	}
	for _, it := range resArray {
		ranks, err := redis.Values(it, nil)
		if err != nil || len(ranks) < 1 {
			logs.Error("GetSimpleOppInfo acids: %v", err)
			return nil
		}
		rankAcid, err := redis.String(ranks[0], nil)
		if err != nil {
			logs.Error("GetSimpleOppInfo acids: %v", err)
			return nil
		}
		acids = append(acids, rankAcid)
	}

	return acids
}

func getSimpleInfoByAcids(groupId int, acids []string) []*WSPVPOppSimple {
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return nil
	}
	simpleInfo := make([]*WSPVPOppSimple, len(acids))
	for i, acid := range acids {
		if IsRobotId(acid) {
			tableName := getRobotTableName(groupId)
			robotInfo, err := redis.String(modules.DoWraper(WS_PVP_DB, db, "HGET", tableName, acid))
			if err != nil {
				logs.Error("get robot info err, %v", err)
				return nil
			} else {
				robot := &WSPVPRobotInfo{}
				err = json.Unmarshal([]byte(robotInfo), robot)
				if err != nil {
					logs.Error("parse robot info err, %v", err)
					return nil
				}
				simpleInfo[i] = &WSPVPOppSimple{
					Acid:     acid,
					Name:     robot.Name,
					ServerId: robot.ServerId,
				}
			}

		} else {
			tableName := getPersonalTableName(groupId, acid)
			reply, err := redis.Values(modules.DoWraper(WS_PVP_DB, db, "HMGET", tableName, "name", "gname", "sid", "title_id", "vlv"))
			if err != nil {
				// TODO 名次变化时可能会出现
				logs.Error("get personal info , %v", err)
				return nil
			} else {
				robotArray := []WSPVPOppSimple{}
				logs.Debug("reply %v", reply)
				err = redis.ScanSlice(reply, &robotArray, "name", "gname", "sid", "title_id", "vlv")
				if err != nil || len(robotArray) == 0 {
					logs.Error("parse personal info err, %v, %v", err, robotArray)
				}
				simpleInfo[i] = &robotArray[0]
				simpleInfo[i].Acid = acid
			}
		}
	}
	return simpleInfo
}

// 去锁定表里检查是否锁定，如果已经锁定返回false，如果已经过期或者没有锁定， 直接锁定返回true
// 保存被谁锁定 存两个key, 一个是时间， 一个是锁定人
func TryLockOpponent(groupId int, acid string, lockOwner string, lockTime int64) bool {
	//if game.Cfg.Gid == 203 {
	//	logs.Debug("203 gid")
	//	return TryLockOpponentForQQ(groupId, acid, lockOwner, lockTime)
	//}
	nowTime := time.Now().Unix()
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return false
	}
	tableName := getLockTableName(groupId)
	keyTime := fmt.Sprintf("time:%s", acid)
	keyAcid := fmt.Sprintf("acid:%s", acid)
	lua_src := `
		local tableName = KEYS[1]
		local keyTime = KEYS[2]
		local keyAcid = KEYS[3]
		local newLockTime = KEYS[4]
		local lockOwner = KEYS[5]
		local nowTime = KEYS[6]
		local setResult = redis.call("HSETNX", tableName, keyTime, newLockTime)
		if setResult == 1
		then
			redis.call("HSET", tableName, keyAcid, lockOwner)
			return 1
		else
			local expireTime = redis.call("HGET", tableName, keyTime)
			if tonumber(expireTime) < tonumber(nowTime)
			then
				redis.call("HSET", tableName, keyTime, newLockTime)
				redis.call("HSET", tableName, keyAcid, lockOwner)
				return 1
			else
				local oldLockAcid = redis.call("HGET", tableName, keyAcid)
				if oldLockAcid == lockOwner
				then
					redis.call("HSET", tableName, keyTime, newLockTime)
					redis.call("HSET", tableName, keyAcid, lockOwner)
					return 1
				else
					return 0
				end
			end
		end
		return 0
	`
	result, err := redis.Int(modules.DoWraper(WS_PVP_DB, db, "EVAL", lua_src, 6,
		tableName, keyTime, keyAcid, int(lockTime), lockOwner, nowTime))
	if err != nil {
		logs.Error("wspvp lock %d, %v", result, err)
		return false
	}
	return result == 1
}

func TryLockOpponentForQQ(groupId int, acid string, lockOwner string, lockTime int64) bool {
	nowTime := time.Now().Unix()
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return false
	}
	tableName := getLockTableName(groupId)
	keyTime := fmt.Sprintf("time:%s", acid)
	keyAcid := fmt.Sprintf("acid:%s", acid)

	ok, err := redis.Int(modules.DoWraper(WS_PVP_DB, db, "HSETNX", tableName, keyTime, int(lockTime)))
	if err != nil {
		logs.Error("wspvp lock 1 %d, %v", ok, err)
		return false
	}
	if ok == 1 {
		modules.DoWraper(WS_PVP_DB, db, "HSET", tableName, keyAcid, lockOwner)
		return true
	} else {
		expireTime, err := redis.Int64(modules.DoWraper(WS_PVP_DB, db, "HGET", tableName, keyTime))
		if err != nil {
			logs.Error("wspvp lock 2 %d, %v", ok, err)
			return false
		}
		if expireTime < nowTime {
			modules.DoWraper(WS_PVP_DB, db, "HSET", tableName, keyTime, lockTime)
			modules.DoWraper(WS_PVP_DB, db, "HSET", tableName, keyAcid, lockOwner)
			return true
		} else {
			oldLockAcid, err := redis.String(modules.DoWraper(WS_PVP_DB, db, "HGET", tableName, keyAcid))
			if err != nil {
				logs.Error("wspvp lock 3 %d, %v", ok, err)
				return false
			}
			if oldLockAcid == lockOwner {
				modules.DoWraper(WS_PVP_DB, db, "HSET", tableName, keyTime, lockTime)
				modules.DoWraper(WS_PVP_DB, db, "HSET", tableName, keyAcid, lockOwner)
				return true
			} else {
				return false
			}
		}
	}
	return false
}

// 交换排名 原子操作 返回两人的新排名,
// 只有保证aci1低于acid2的排名的时候，才会交换
// 需要保证是acid1攻打acid2
// 返回新排名和排名是否变化
func SwapRank(groupId int, acid1, acid2 string) (int, int, bool) {
	//if game.Cfg.Gid == 203 {
	//	return SwapRankForQQ(groupId, acid1, acid2)
	//}
	tableName := getRankTableName(groupId)
	lua_src := `
		local tableName = KEYS[1]
		local acid1 = KEYS[2]
		local acid2 = KEYS[3]
		local rank1 = redis.call("ZRANK", tableName, acid1)
	        local rank2 = redis.call("ZRANK", tableName, acid2)
	        if rank1 == false and rank2 == false
	        then
	        	return {-1, -1, 0}
		end
		if rank2 == false
		then
			return {rank1, -1, 0}
		end
		if rank1 == false
		then
			redis.call("ZREM", tableName, acid2)
			redis.call("ZADD", tableName, rank2, acid1)
			return {rank2, -1, 1}
		else
			if tonumber(rank2) > tonumber(rank1)
			then
				return {rank1, rank2, 0}
			else
		    		redis.call("ZADD", tableName, rank1, acid2)
				redis.call("ZADD", tableName, rank2, acid1)
				return {rank2, rank1, 1}
			end
		end
		return {rank2, rank1, 1}
	`
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return -1, -1, false
	}
	reply, err := modules.DoWraper(WS_PVP_DB, db, "EVAL", lua_src, 3, tableName, acid1, acid2)
	if err != nil {
		logs.Error("SwapRank error, %v", err)
		return -1, -1, false
	}
	ranks, err := redis.Ints(reply, nil)
	if err != nil {
		logs.Error("swapRank result error, %v", err)
		return -1, -1, false
	}
	if len(ranks) != 3 {
		logs.Error("swapRank rank error, %v", err)
		return -1, -1, false
	}
	return ranks[0] + 1, ranks[1] + 1, ranks[2] == 1
}

func SwapRankForQQ(groupId int, acid1, acid2 string) (int, int, bool) {
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return -1, -1, false
	}

	tableName := getRankTableName(groupId)
	rank1, err1 := redis.Int(modules.DoWraper(WS_PVP_DB, db, "ZRANK", tableName, acid1))
	if err1 != nil && err1 != redis.ErrNil {
		logs.Error("swap rank err 1 %v", err1)
		return -1, -1, false
	}
	rank2, err2 := redis.Int(modules.DoWraper(WS_PVP_DB, db, "ZRANK", tableName, acid2))
	if err2 != nil && err2 != redis.ErrNil {
		logs.Error("swap rank err 2 %v", err2)
		return -1, -1, false
	}
	if err1 == redis.ErrNil && err2 == redis.ErrNil {
		return -1, -1, false
	} else if err2 == redis.ErrNil {
		return rank1 + 1, 0, false
	} else if err1 == redis.ErrNil {
		modules.DoWraper(WS_PVP_DB, db, "ZREM", tableName, acid2)
		modules.DoWraper(WS_PVP_DB, db, "ZADD", tableName, rank2, acid1)
		return rank2 + 1, 0, true
	} else {
		if rank2 > rank1 {
			return rank1 + 1, rank2 + 1, false
		} else {
			modules.DoWraper(WS_PVP_DB, db, "ZADD", tableName, rank1, acid2)
			modules.DoWraper(WS_PVP_DB, db, "ZADD", tableName, rank2, acid1)
			return rank2 + 1, rank1 + 1, true
		}
	}
}

const MaxLogNum = 30

func RecordLog(groupId int, acid string, att, result bool,
	rankChg int, oppName, oppGName string, newRank int) {
	//if game.Cfg.Gid == 203 {
	//	RecordLogForQQ(groupId, acid, att, result, rankChg, oppName, oppGName, newRank)
	//	return
	//}
	logInfo := &WSPVPBattleLog{
		Attack:            att,
		Result:            result,
		RankChange:        int64(rankChg),
		OpponentName:      oppName,
		OpponentGuildName: oppGName,
		Time:              time.Now().Unix(),
		Rank:              newRank,
	}
	logJson, _ := json.Marshal(logInfo)
	tableName := getBattleLogTableName(groupId, acid)
	lua_src := `
		local tableName = KEYS[1]
		local logJson = KEYS[2]
		local maxNum = KEYS[3]
		local len = redis.call("LPUSH", tableName, logJson)
		redis.call("LTRIM", tableName, 0, maxNum-1)
		return "OK"`
	dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		err := cb.Send("EVAL", lua_src, 3, tableName, logJson, MaxLogNum)
		if err != nil {
			return err
		}
		return nil
	})
}

func RecordLogForQQ(groupId int, acid string, att, result bool,
	rankChg int, oppName, oppGName string, newRank int) {
	logInfo := &WSPVPBattleLog{
		Attack:            att,
		Result:            result,
		RankChange:        int64(rankChg),
		OpponentName:      oppName,
		OpponentGuildName: oppGName,
		Time:              time.Now().Unix(),
		Rank:              newRank,
	}
	logJson, _ := json.Marshal(logInfo)
	tableName := getBattleLogTableName(groupId, acid)

	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return
	}

	len, err := redis.Int(modules.DoWraper(WS_PVP_DB, db, "LPUSH", tableName, logJson))
	if err != nil {
		logs.Error("record log 1 err %v", err)
		return
	}
	if len > MaxLogNum {
		modules.DoWraper(WS_PVP_DB, db, "LTRIM", tableName, 0, MaxLogNum-1)
	}
}

// 取消锁定
func UnlockOpponent(groupId int, acid string, lockOwner string) {
	//if game.Cfg.Gid == 203 {
	//	UnlockOpponentForQQ(groupId, acid, lockOwner)
	//	return
	//}
	tableName := getLockTableName(groupId)
	keyTime := fmt.Sprintf("time:%s", acid)
	keyAcid := fmt.Sprintf("acid:%s", acid)
	lua_src := `
		local tableName = KEYS[1]
		local timeKey = KEYS[2]
		local nameKey = KEYS[3]
		local lockOwner = KEYS[4]
		local lockAcid = redis.call("HGET", tableName, nameKey)
		if lockAcid == lockOwner
		then
			redis.call("HDEL", tableName, nameKey)
			redis.call("HDEL", tableName, timeKey)
		end
		return "OK"`
	dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		err := cb.Send("EVAL", lua_src, 4, tableName, keyTime, keyAcid, lockOwner)
		if err != nil {
			return err
		}
		return nil
	})
}

func UnlockOpponentForQQ(groupId int, acid string, lockOwner string) {
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return
	}
	tableName := getLockTableName(groupId)
	keyTime := fmt.Sprintf("time:%s", acid)
	keyAcid := fmt.Sprintf("acid:%s", acid)

	lockAcid, err := redis.String(modules.DoWraper(WS_PVP_DB, db, "HGET", tableName, keyAcid))
	if err != nil {
		logs.Error("UnlockOpponentForQQ 1 %v", err)
		return
	}
	if lockAcid == lockOwner {
		modules.DoWraper(WS_PVP_DB, db, "HDEL", tableName, keyAcid)
		modules.DoWraper(WS_PVP_DB, db, "HDEL", tableName, keyTime)
	}
}

// 获取某个人的战斗日志
func GetWSPVPLog(groupId int, acid string) []*WSPVPBattleLog {
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return nil
	}
	tableName := getBattleLogTableName(groupId, acid)
	reply, err := modules.DoWraper(WS_PVP_DB, db, "LRANGE", tableName, 0, MaxLogNum-1)
	if err != nil {
		logs.Error("get wspvp log error, %v", err)
		return nil
	}
	logs.Debug("reply %v", reply)
	jsonLogs, err := redis.Strings(reply, nil)
	if err != nil {
		logs.Error("parse wspvp log error, %v", err)
		return nil
	}
	logs.Debug("parse reply %v", jsonLogs)
	result := make([]*WSPVPBattleLog, 0)
	for _, str := range jsonLogs {
		bLog := &WSPVPBattleLog{}
		if err := json.Unmarshal([]byte(str), bLog); err == nil {
			result = append(result, bLog)
		}

	}
	return result
}

// 获取排行榜上人的信息 自动解析
// TODO 优化 选取字段获取信息， scan slice
func GetRankPlayerAllInfo(groupId int, acid string) *WSPVPInfo {
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return nil
	}
	tableName := getPersonalTableName(groupId, acid)
	resInfo := new(WSPVPInfo)
	driver.RestoreFromHashDB(db.RawConn(), tableName, resInfo, false, false)
	return resInfo
}

// 加载top n
func loadTopN(groupId int) []*WsPvpRankPlayer {
	rankAcids := getTopRankAcids(groupId)
	return loadRankPlayer(groupId, rankAcids)
}

func loadRankPlayer(groupId int, rankAcids []string) []*WsPvpRankPlayer {
	personAcids := make([]string, 0)
	robotAcids := make([]string, 0)
	for _, acid := range rankAcids {
		if !IsRobotId(acid) {
			personAcids = append(personAcids, acid)
		} else {
			robotAcids = append(robotAcids, acid)
		}
	}

	personMap := getPersonInfo(groupId, personAcids)
	robotMap := getRobotInfo(groupId, robotAcids)
	topRank := make([]*WsPvpRankPlayer, 0)
	for i, acid := range rankAcids {
		if IsRobotId(acid) {
			topRank = append(topRank, robotMap[acid])
		} else {
			topRank = append(topRank, personMap[acid])
		}
		topRank[i].Rank = i + 1
		topRank[i].SidStr = gamedata.GetSidName(game.Cfg.EtcdRoot, uint(game.Cfg.Gid), uint(topRank[i].Sid))
	}
	return topRank
}

func getTopRankAcids(groupId int) []string {
	tableName := getRankTableName(groupId)
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return nil
	}
	reply, err := modules.DoWraper(WS_PVP_DB, db, "ZRANGE", tableName, 0, TOP_N-1)

	if err != nil {
		logs.Error("getTopRankAcids %v", err)
		return nil
	} else {
		//logs.Debug("get top rank acids, %v", reply)
	}
	rankAcids, err := redis.Strings(reply, nil)
	if err != nil {
		logs.Error("rank acids, %v", err)
		return nil
	}
	return rankAcids
}

func loadBest9TopN(groupId int) []*WsPvpRankPlayer {
	rankAcids := getBest9TopRankAcids(groupId)
	return loadRankPlayer(groupId, rankAcids)
}

func getBest9TopRankAcids(groupId int) []string {
	tableName := getBest9RankTableName(groupId)
	db := getDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return nil
	}
	reply, err := modules.DoWraper(WS_PVP_DB, db, "ZREVRANGE", tableName, 0, TOP_N-1)

	if err != nil {
		logs.Error("getBest9TopRankAcids %v", err)
		return nil
	} else {
		//logs.Debug("get top rank acids, %v", reply)
	}
	rankAcids, err := redis.Strings(reply, nil)
	if err != nil {
		logs.Error("rank acids, %v", err)
		return nil
	}
	return rankAcids
}

func getPersonInfo(groupId int, personAcids []string) map[string]*WsPvpRankPlayer {
	rankMap := make(map[string]*WsPvpRankPlayer)
	if personAcids == nil || len(personAcids) == 0 {
		return rankMap
	}
	reply, errCode := dbHMGETExec(func(cb redis.CmdBuffer) error {
		for _, acid := range personAcids {
			personTableName := getPersonalTableName(groupId, acid)
			err := cb.Send("HMGET", personTableName, "acid", "name", "sid", "gname", "clv", "allgs")
			if err != nil {
				return err
			}
		}
		return nil

	})
	logs.Debug("get person info reply %v", reply)
	if errCode != 0 {
		return nil
	}
	rankInfos, err := redis.Values(reply, nil)
	if err != nil {
		logs.Error("get person info err %v", err)
		return nil
	}
	logs.Debug("get person info values %v", rankInfos)
	rankResult := []WsPvpRankPlayer{}
	err = redis.ScanSlice(rankInfos, &rankResult, "acid", "name", "sid", "gname", "clv", "allgs")
	if err != nil {
		logs.Error("get person info err %v", err)
		return nil
	}
	logs.Debug("get person info result, %v", rankResult)
	for i, p := range rankResult {
		rankMap[p.Acid] = &rankResult[i]
	}
	return rankMap
}

func getRobotInfo(groupId int, robotIds []string) map[string]*WsPvpRankPlayer {
	rankMap := make(map[string]*WsPvpRankPlayer)
	if robotIds == nil || len(robotIds) == 0 {
		return rankMap
	}
	robotTableName := getRobotTableName(groupId)
	reply, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for _, acid := range robotIds {
			err := cb.Send("HGET", robotTableName, acid)
			if err != nil {
				return err
			}
		}
		return nil

	})
	if errCode != 0 {
		return nil
	}
	logs.Debug("get robot info %v", reply)
	rankInfos, err := redis.Strings(reply, nil)
	if err != nil {
		logs.Error("get robot info, %v", err)
		return nil
	}
	for i, p := range rankInfos {
		pRobot := &WSPVPRobotInfo{}
		err := json.Unmarshal([]byte(p), pRobot)
		if err == nil {
			rankMap[robotIds[i]] = &WsPvpRankPlayer{
				Name: pRobot.Name,
				Sid:  pRobot.ServerId,
				Acid: robotIds[i],
			}
		}
	}
	return rankMap
}

// TODO 已经掉到10000以外的呢  这种情况造成的影响是数据库会多一些冗余数据
func BatchSavePlayerInfo(groupId int, players []*WSPVPInfo) {
	best9RankTable := getBest9RankTableName(groupId)
	_, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for _, player := range players {
			personTableName := getPersonalTableName(groupId, player.Acid)
			err := driver.DumpToHashDBCmcBuffer(cb, personTableName, player)
			if err != nil {
				return err
			}
			cb.Send("ZADD", best9RankTable, player.AllGs, player.Acid)
		}
		return nil
	})
	if errCode != 0 {
		logs.Error("err code %d", errCode)
	}
}

func SavePlayerInfo(groupId int, player *WSPVPInfo) {
	best9RankTable := getBest9RankTableName(groupId)
	_, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		personTableName := getPersonalTableName(groupId, player.Acid)
		err := driver.DumpToHashDBCmcBuffer(cb, personTableName, player)
		if err != nil {
			return err
		}
		cb.Send("ZADD", best9RankTable, player.AllGs, player.Acid)
		return nil
	})
	if errCode != 0 {
		logs.Error("err code %d", errCode)
	}
}

func initRobotInfo(sid, groupId int, robotId, names []string) {
	tableName := getRankTableName(groupId)
	_, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for i, acid := range robotId {
			err := cb.Send("ZADD", tableName, i, acid)
			if err != nil {
				return err
			}
		}
		return nil

	})
	if errCode != 0 {
		logs.Error("err code %d", errCode)
	}
	robotTableName := getRobotTableName(groupId)
	_, errCode = dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for i, acid := range robotId {
			robotInfo := WSPVPRobotInfo{
				Name:     names[i],
				ServerId: sid,
			}
			robotJson, err := json.Marshal(robotInfo)
			if err != nil {
				return err
			}
			err = cb.Send("HSET", robotTableName, acid, robotJson)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if errCode != 0 {
		logs.Error("err code %d", errCode)
	}
}

func initNameRedis(sid uint32, names, robotIds []string) {
	db := driver.GetDBConn()
	defer db.Close()
	cbNames := redis.NewCmdBuffer()
	for i := 0; i < WS_PVP_RANK_MAX; i++ {
		// 用hsetnx防止重名覆盖已有玩家信息
		cbNames.Send("HSETNX", driver.TableChangeName(uint(sid)), names[i], robotIds[i])
	}
	if _, err := modules.DoCmdBufferWrapper(WS_PVP_DB, db, cbNames, true); err != nil {
		panic(fmt.Errorf("wspvp initAllRobot names DoCmdBufferWrapper err %s", err.Error()))
	}
}

func getRankSize(groupId int) int {
	tableName := getRankTableName(groupId)
	reply, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		err := cb.Send("ZCARD", tableName)
		if err != nil {
			return err
		}
		return nil

	})
	if errCode != 0 {
		logs.Error("err code %d", errCode)
		return -1
	}
	count, err := redis.Ints(reply, nil)
	if err != nil {
		logs.Error("err code %v", err)
		return -1
	}
	return count[0]
}

func DelPlayerFromInfo(groupId int, acid string) {
	tableName := getPersonalTableName(groupId, acid)
	dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		err := cb.Send("HDEL", tableName, acid)
		if err != nil {
			return err
		}
		return nil
	})
}

func DeubgSetRank(groupId int, acid string, newRank int) int {
	swapAcid := getAcidsByRank(groupId, []int{newRank})[0]
	rank1, _, ok := SwapRank(groupId, acid, swapAcid)
	if !ok {
		rank1, _, ok = SwapRank(groupId, swapAcid, acid)
		if !ok {
			return -1
		}
	}
	return rank1
}

func GetBest9RankByOne(groupId int, acid string) int {
	ranks := GetBest9Rank(groupId, []string{acid})
	if len(ranks) == 0 {
		return 0
	} else {
		return ranks[0]
	}
}

func GetBest9Rank(groupId int, acids []string) []int {
	tableName := getBest9RankTableName(groupId)
	reply, errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for _, acid := range acids {
			err := cb.Send("ZREVRANK", tableName, acid)
			if err != nil {
				return err
			}
		}
		return nil
	})
	logs.Debug("get ranks src %v", reply)
	if errCode != 0 {
		logs.Error("get ranks err, %d", errCode)
		return nil
	} else {
		newRank, err := IntsWithDefault(reply, -1)
		if err == nil {
			if len(newRank) != len(acids) {
				logs.Error("check ranks result err, %d", errCode)
				return nil
			}
			for i := 0; i < len(newRank); i++ {
				newRank[i]++
			}
			return newRank
		} else {
			logs.Error("parse ranks err, %v", err)
			// 如果不能解析10000名以外的，说明这个名次肯定发生了变化， 返回Nil
			return nil
		}
	}
}

func DebugCopyWspvpLog(groupId int, acid string, count int) {
	tableName := getBattleLogTableName(groupId, acid)
	db := getDBConn()
	headLog, err := redis.String(db.Do("LINDEX", tableName, 0))
	if err != nil {
		logs.Error("<DebugCopyWspvpLog> LINDEX err ", err)
		return
	}
	insertLogs := make([]interface{}, 0)
	insertLogs = append(insertLogs, tableName)
	for i := 0; i < count; i++ {
		insertLogs = append(insertLogs, headLog)
	}
	db.Do("LPUSH", insertLogs...)
}

func IsInitRobot(groupId int) bool {
	db := getDBConn()
	defer db.Close()
	key := getInitKeyTableName(groupId)
	isExist, err := redis.Int(db.Do("SETNX", key, game.Cfg.ShardId[0]))
	if err != nil {
		logs.Error("SETNX Key value error", err)
		return false
	}
	return isExist == 1
}
