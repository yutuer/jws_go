/*
重构的ws_pvp测试用例，使用了SubTest
请使用GO1.8以上版本进行测试
并确保本地Redis服务正常
*/
package ws_pvp

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"

	"github.com/stretchr/testify/assert"
)

var (
	testGroupId     uint32 = 10001
	testRankGroupId uint32 = 65536
	testSid         int    = 255
)

func debugInitRank() {
	waitter := util.WaitGroupWrapper{}
	ws := new(WSPVPModule)
	ws.groupId = int(testGroupId)
	ws.sid = uint(testSid)
	ir := new(InitRobotTest)
	ir.groupId = testGroupId
	waitter.Wrap(func() {
		ws.initRank(ir)
	})
	waitter.Wait()
}

func clearDb(groupId int) {
	db := getDBConn()
	defer db.Close()
	initKey := getInitKeyTableName(groupId)
	best9Key := getBest9RankTableName(groupId)
	lockKey := getLockTableName(groupId)
	rankKey := getRankTableName(groupId)
	robotKey := getRobotTableName(groupId)
	db.Do("DEL", initKey, best9Key, lockKey, rankKey, robotKey)
}

func clearLog(groupId int, acid string) {
	db := getDBConn()
	defer db.Close()
	logkey := getBattleLogTableName(groupId, acid)
	db.Do("DEL", logkey)
}

func clearInfo(groupId int, acid string) {
	db := getDBConn()
	defer db.Close()

	logkey := getPersonalTableName(groupId, acid)
	db.Do("DEL", logkey)
}

type InitRobotTest struct {
	groupId uint32
}

func (ir *InitRobotTest) InitRobt(sid uint32) {
	//time.Sleep(time.Second * 5)

	tableName := getRankTableName(int(ir.groupId))
	robotIds := make([]string, WS_PVP_RANK_MAX)
	for i := 0; i < WS_PVP_RANK_MAX; i++ {
		robotIds[i] = GenRobotId(ir.groupId, sid, i)
	}
	dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for i, acid := range robotIds {
			err := cb.Send("ZADD", tableName, i, acid)
			if err != nil {
				return err
			}
		}
		return nil
	})

	names := gamedata.RandRobotNames(WS_PVP_RANK_MAX)
	robotTableName := getRobotTableName(int(ir.groupId))
	dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for i, acid := range robotIds {
			robotInfo := WSPVPRobotInfo{
				Name:     names[i],
				ServerId: int(sid),
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
}

func (ir *InitRobotTest) InitTopN(module *WSPVPModule) {
	logs.Info("init topn for test do nothing")
}

func debugSavePlayerInfo() {
	player := new(WSPVPInfo)
	player.Acid = "profile:10001:1234567890"
	player.Name = "test"
	player.AvatarId = 1
	player.CorpLv = 2
	player.GuildName = "guild"
	player.ServerId = testSid
	player.VipLevel = 10
	player.AllGs = 12345678
	SavePlayerInfo(int(testGroupId), player)

	robotAcid := fmt.Sprintf("wspvp:%d:%d:9999", testGroupId, testSid)

	dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		rankTable := getRankTableName(int(testGroupId))
		cb.Send("ZREM", rankTable, 9999, robotAcid)
		cb.Send("ZADD", rankTable, 9999, player.Acid)
		return nil
	})

	player = new(WSPVPInfo)
	player.Acid = "profile:10001:111111111111"
	player.CorpLv = 20
	player.AllGs = 88888888
	SavePlayerInfo(int(testGroupId), player)
}

/*
	上面都是构件
	用例由此开始
*/

func TestInitRank1(t *testing.T) {
	groupId := testRankGroupId
	defer clearDb(int(groupId))

	// 全部新服的情况 没有锁，没有rank
	waitter := util.WaitGroupWrapper{}
	for i := 0; i < 10; i++ {
		ws := new(WSPVPModule)
		ws.groupId = int(groupId)
		ws.sid = uint(i + 1)
		ir := new(InitRobotTest)
		ir.groupId = groupId
		waitter.Wrap(func() {
			ws.initRank(ir)
		})
	}
	waitter.Wait()

	count := getRankSize(int(groupId))
	assert.Equal(t, 10000, count)
}

func TestInitRank2(t *testing.T) {
	groupId := testRankGroupId
	defer clearDb(int(groupId))

	// 已存在分组的情况 没有锁， 有rank
	ir := new(InitRobotTest)
	ir.InitRobt(testGroupId)
	defer clearDb(0)		// 69行没有赋值……所以……

	waitter := util.WaitGroupWrapper{}
	for i := 0; i < 10; i++ {
		ws := new(WSPVPModule)
		ws.groupId = int(groupId)
		ws.sid = uint(i + 1)
		ir := new(InitRobotTest)
		ir.groupId = groupId
		waitter.Wrap(func() {
			ws.initRank(ir)
		})
	}
	waitter.Wait()

	count := getRankSize(int(groupId))
	assert.Equal(t, 10000, count)
}

func TestGetFuncs(t *testing.T) {
	debugInitRank()
	defer clearDb(int(testGroupId))

	t.Run("LoadTopN", func(t *testing.T) {
		topN := loadTopN(int(testGroupId))
		assert.Equal(t, 100, len(topN))
	})

	t.Run("GetAcidsByRank", func(t *testing.T) {
		acids := getAcidsByRank(int(testGroupId), []int{1, 100, 1000, 10000})
		assert.Equal(t, 4, len(acids), fmt.Sprintf("%v", acids))
	})

	t.Run("GetSimpleOppInfo", func(t *testing.T) {
		simpleInfos := GetSimpleOppInfo(int(testGroupId), []int{1, 2, 1000, 10000})
		assert.Equal(t, 4, len(simpleInfos))

		for _, simp := range simpleInfos {
			t.Log(*simp)
			assert.Equal(t, testSid, (*simp).ServerId)
			assert.NotEmpty(t, (*simp).ServerId)
		}

	})

	t.Run("GetPersonInfo", func(t *testing.T) {
		debugSavePlayerInfo()

		acid0 := "profile:10001:1234567890"
		acid1 := "profile:10001:111111111111"
		acid2 := "what?"

		defer clearInfo(int(testGroupId), acid0)
		defer clearInfo(int(testGroupId), acid1)

		info := getPersonInfo(int(testGroupId), []string{acid0, acid1, acid2})

		assert.Equal(t, 3, len(info))
		assert.Equal(t, 2, info[acid0].CorpLevel)
		assert.Equal(t, 20, info[acid1].CorpLevel)
		assert.Nil(t, info[acid2])
	})

	t.Run("GetRanks", func(t *testing.T) {
		bot10000Acid := fmt.Sprintf("wspvp:%d:%d:10000", testGroupId, testSid)

		// Setup
		dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			rankTable := getRankTableName(int(testGroupId))
			cb.Send("ZADD", rankTable, 10000, bot10000Acid)
			return nil
		})

		t.Run("正常", func(t *testing.T) {
			ranks := GetRanks(int(testGroupId), []string{
				fmt.Sprintf("wspvp:%d:%d:0", testGroupId, testSid),
				fmt.Sprintf("wspvp:%d:%d:1", testGroupId, testSid),
				fmt.Sprintf("wspvp:%d:%d:100", testGroupId, testSid),
				fmt.Sprintf("wspvp:%d:%d:10000", testGroupId, testSid)})

			assert.Equal(t, 4, len(ranks))
			assert.Equal(t, []int{1, 2, 101, 10001}, ranks)
			t.Log(ranks)
		})

		t.Run("Acid不存在", func(t *testing.T) {
			ranks := GetRanks(int(testGroupId), []string{fmt.Sprintf("wspvp:%d:%d:44444", testGroupId, testSid)})
			assert.NotNil(t, ranks)
			assert.NotEmpty(t, ranks)
			assert.Equal(t, 0, ranks[0])
		})

		// TearDown
		dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			rankTable := getRankTableName(int(testGroupId))
			cb.Send("ZREM", rankTable, 10000, bot10000Acid)
			return nil
		})
	})

	t.Run("GetRankPlayerInfo", func(t *testing.T) {
		acid1 := fmt.Sprintf("%d:%d:%s", testGroupId, testSid, uuid.NewV4().String())
		acid2 := fmt.Sprintf("%d:%d:%s", testGroupId, testSid, uuid.NewV4().String())

		// 新建一个并存储
		player := new(WSPVPInfo)
		player.Acid = acid1
		player.Name = "TestGetRankPlayerInfo"
		player.CorpLv = 99
		SavePlayerInfo(int(testGroupId), player)
		defer clearInfo(int(testGroupId), acid1)

		t.Run("正常", func(t *testing.T) {
			info := GetRankPlayerAllInfo(int(testGroupId), acid1)
			assert.Equal(t, acid1, info.Acid)
			assert.Equal(t, "TestGetRankPlayerInfo", info.Name)
			assert.Equal(t, uint32(99), info.CorpLv)
		})

		t.Run("不存在", func(t *testing.T) {
			info := GetRankPlayerAllInfo(int(testGroupId), acid2)
			assert.Equal(t, "", info.Name)
			assert.Equal(t, uint32(0), info.CorpLv)
		})
	})

	t.Run("GetBest9RankByOne", func(t *testing.T) {
		acid0 := "00001"
		acid1 := "profile:1982:111111111111"

		// 新建一个并存储
		player := new(WSPVPInfo)
		player.Acid = acid1
		player.Name = "GetBest9RankByOne"
		player.CorpLv = 110
		player.AllGs = 65536
		SavePlayerInfo(int(testGroupId), player)
		defer clearInfo(int(testGroupId), acid1)

		rank := GetBest9RankByOne(int(testGroupId), acid0)
		assert.Equal(t, 0, rank)

		rank1 := GetBest9RankByOne(int(testGroupId), acid1)
		assert.NotEqual(t, 0, rank1)
	})
}

func TestTryLockOpponent(t *testing.T) {
	debugInitRank()
	defer clearDb(int(testGroupId))

	bot1Acid := fmt.Sprintf("wspvp:%d:%d:1", testGroupId, testSid)
	bot2Acid := fmt.Sprintf("wspvp:%d:%d:2", testGroupId, testSid)
	bot3Acid := fmt.Sprintf("wspvp:%d:%d:3", testGroupId, testSid)

	t.Run("锁定没有被锁过的角色", func(t *testing.T) {
		nowTime := time.Now().Unix()
		result := TryLockOpponent(int(testGroupId), bot1Acid, bot2Acid, nowTime+10)
		assert.True(t, result)

		UnlockOpponent(int(testGroupId), bot1Acid, bot2Acid)
	})

	t.Run("单个账号连续锁定2次", func(t *testing.T) {
		nowTime := time.Now().Unix()
		r1 := TryLockOpponent(int(testGroupId), bot1Acid, bot2Acid, nowTime+10)
		assert.True(t, r1)
		r2 := TryLockOpponent(int(testGroupId), bot1Acid, bot2Acid, nowTime+10)
		assert.True(t, r2)

		UnlockOpponent(int(testGroupId), bot1Acid, bot2Acid)
	})

	t.Run("锁定已经过期的角色", func(t *testing.T) {
		nowTime := time.Now().Unix()
		r1 := TryLockOpponent(int(testGroupId), bot1Acid, bot2Acid, nowTime-10)
		assert.True(t, r1)
		r2 := TryLockOpponent(int(testGroupId), bot1Acid, bot3Acid, nowTime+9)
		assert.True(t, r2)

		UnlockOpponent(int(testGroupId), bot1Acid, bot3Acid)
	})

	t.Run("锁定已经被别人锁定而且未过期的", func(t *testing.T) {
		nowTime := time.Now().Unix()
		r1 := TryLockOpponent(int(testGroupId), bot1Acid, bot2Acid, nowTime+10)
		assert.True(t, r1)
		r2 := TryLockOpponent(int(testGroupId), bot1Acid, bot3Acid, nowTime+10)
		assert.False(t, r2)

		UnlockOpponent(int(testGroupId), bot1Acid, bot2Acid)

		r3 := TryLockOpponent(int(testGroupId), bot1Acid, bot2Acid, nowTime+10)
		assert.True(t, r3)
		r4 := TryLockOpponent(int(testGroupId), bot1Acid, bot3Acid, nowTime+20)
		assert.False(t, r4)

		UnlockOpponent(int(testGroupId), bot1Acid, bot2Acid)
	})

	t.Run("目标和Owner一样", func(t *testing.T) {
		nowTime := time.Now().Unix()
		result := TryLockOpponent(int(testGroupId), bot1Acid, bot1Acid, nowTime+10)
		assert.True(t, result)

		UnlockOpponent(int(testGroupId), bot1Acid, bot1Acid)
	})

	t.Run("db连接错误", func(t *testing.T) {
		poolSrc := pool
		pool = nil
		SetupRedis(":46379", 15, "", true)

		nowTime := time.Now().Unix()
		result := TryLockOpponent(int(testGroupId), bot1Acid, bot2Acid, nowTime+10)
		assert.False(t, result)

		pool = poolSrc
	})
}

func TestSwapRank(t *testing.T) {
	debugInitRank()
	defer clearDb(int(testGroupId))

	bot9998Acid := fmt.Sprintf("wspvp:%d:%d:9998", testGroupId, testSid)
	bot10007Acid := fmt.Sprintf("wspvp:%d:%d:10007", testGroupId, testSid)

	t.Run("case 1 被挑战的人不在排行榜里", func(t *testing.T) {
		rankArray := GetRanks(int(testGroupId), []string{bot9998Acid})
		oldRank := rankArray[0]
		rank1, rank2, _ := SwapRank(int(testGroupId), bot9998Acid, bot10007Acid)
		assert.Equal(t, oldRank, rank1)
		assert.Equal(t, 0, rank2)
	})

	t.Run("case 2 较高的名次不会和较低的名次互换", func(t *testing.T) {
		acids := getAcidsByRank(int(testGroupId), []int{1, 2})
		rank1, rank2, _ := SwapRank(int(testGroupId), acids[0], acids[1])
		assert.Equal(t, 1, rank1)
		assert.Equal(t, 2, rank2)
	})

	t.Run("case 3 10000名以外的人进入排行榜内", func(t *testing.T) {
		acids := getAcidsByRank(int(testGroupId), []int{9991})
		bot10009Acid := fmt.Sprintf("wspvp:%d:%d:10009", testGroupId, testSid)
		dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			rankTable := getRankTableName(int(testGroupId))
			cb.Send("ZADD", rankTable, 10009, bot10009Acid)
			return nil
		})

		rank1, rank2, _ := SwapRank(int(testGroupId), bot10009Acid, acids[0])
		assert.Equal(t, 9991, rank1)
		assert.Equal(t, 10001, rank2)

		dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			rankTable := getRankTableName(int(testGroupId))
			cb.Send("ZREM", rankTable, 10009, acids[0])
			return nil
		})
	})

	t.Run("case 4 正常交换", func(t *testing.T) {
		acids := getAcidsByRank(int(testGroupId), []int{1, 2})
		rank1, rank2, _ := SwapRank(int(testGroupId), acids[1], acids[0])
		assert.Equal(t, 1, rank1)
		assert.Equal(t, 2, rank2)
	})
}

func TestGetWSPVPLog(t *testing.T) {
	acid1 := fmt.Sprintf("%d:%d:%s", testGroupId, testSid, uuid.NewV4().String())
	acid2 := fmt.Sprintf("%d:%d:%s", testGroupId, testSid, uuid.NewV4().String())

	RecordLog(int(testGroupId), acid1, true, true,
		1, "lbb001", "lbb002", 121)
	defer clearLog(int(testGroupId), acid1)

	t.Run("正常", func(t *testing.T) {
		logRes := GetWSPVPLog(int(testGroupId), acid1)
		assert.NotEmpty(t, logRes)
		for _, logR := range logRes {
			assert.Equal(t, int64(1), (*logR).RankChange)
			assert.Equal(t, 121, (*logR).Rank)
			assert.Equal(t, "lbb001", (*logR).OpponentName)
			assert.Equal(t, "lbb002", (*logR).OpponentGuildName)
		}
	})

	t.Run("acid不存在", func(t *testing.T) {
		logRes := GetWSPVPLog(int(testGroupId), acid2)
		assert.Empty(t, logRes)
	})

	t.Run("Redis无法连接", func(t *testing.T) {
		savedPool := pool
		pool = nil
		SetupRedis(":404", 15, "", true)

		logRes := GetWSPVPLog(int(testGroupId), acid1)
		assert.Nil(t, logRes)

		pool = savedPool
	})

	t.Run("Table不存在", func(t *testing.T) {
		logRes := GetWSPVPLog(32768, acid1)
		assert.Empty(t, logRes)
	})
}

func TestRecordLog(t *testing.T) {
	acid := fmt.Sprintf("wspvp:%d:%d:11111", testGroupId, testSid)
	for i := 0; i < 40; i++ {
		RecordLog(int(testGroupId), acid, true, true,
			1, "lbb001", "lbb002", 1)
	}
	defer clearLog(int(testGroupId), acid)

	tableName := getBattleLogTableName(int(testGroupId), acid)
	conn := getDBConn()
	listLen, err := redis.Int(conn.Do("LLEN", tableName))
	if err != nil || listLen != 30 {
		t.Error("err or list len %v, %d", err, listLen)
		t.FailNow()
	}
}

func TestDebugCopyWspvpLog(t *testing.T) {
	//TestRecordLog(t)
	acid := fmt.Sprintf("wspvp:%d:%d:11111", testGroupId, testSid)
	for i := 0; i < 40; i++ {
		RecordLog(int(testGroupId), acid, true, true,
			1, "lbb001", "lbb002", 1)
	}
	defer clearLog(int(testGroupId), acid)

	DebugCopyWspvpLog(int(testGroupId), acid, 10)

	conn := getDBConn()
	tableName := getBattleLogTableName(int(testGroupId), acid)
	listLen, err := redis.Int(conn.Do("LLEN", tableName))
	if err != nil || listLen != 40 {
		t.Error("err or list len %v, %d", err, listLen)
		t.FailNow()
	}
}

func BenchmarkSavePersonInfo(b *testing.B) {
	player := new(WSPVPInfo)
	player.Acid = "profile:1:1234567890"
	player.Name = "test"
	player.AvatarId = 1
	player.CorpLv = 2
	player.GuildName = "guild"
	player.ServerId = 1
	player.VipLevel = 10
	player.AllGs = 12345678

	defer clearInfo(int(testGroupId), player.Acid)

	for i := 0; i < b.N; i++ {
		SavePlayerInfo(int(testGroupId), player)
	}
}