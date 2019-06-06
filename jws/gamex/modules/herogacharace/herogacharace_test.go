package herogacharace

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"

	"github.com/stretchr/testify/assert"
)

const (
	testGroupID = 1
)

var (
	//testConn    redis.Conn
	testEtcdCfg map[string]RedisDBSetting
	testHGA     HGRActivity
	testHGA2    HGRActivity //测试相同分数不同时间的情况
	testHGR     *HeroGachaRace
	testHGR2    *HeroGachaRace
)

func TestMain(m *testing.M) {
	gamedata.DebugLoadLocalGamedata()
	MySetup()
	exitc := m.Run()

	testHGR.debugClear()
	testHGR2.debugClear()
	os.Exit(exitc)
}

func MySetup() {
	testHGR = NewHeroGachaRace(1000)
	testHGR2 = NewHeroGachaRace(1000)
	testEtcdCfg = make(map[string]RedisDBSetting)
	testEtcdCfg[fmt.Sprintf("%d", testGroupID)] = RedisDBSetting{
		AddrPort: "127.0.0.1:6379",
		Auth:     "",
		DB:       15,
	}
	game.Cfg.EtcdRoot = "/a6k"
	game.Cfg.Gid = 200
	etcd.InitClient([]string{"http://127.0.0.1:2379/"})

	key := GetEtcdInfoKey()
	jsonvalue, err := json.Marshal(testEtcdCfg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	etcd.Set(key, string(jsonvalue), 0)

	testHGA.GroupID = testGroupID
	testHGA.ActivityId = 1001
	testHGA.StartTime = time.Now().Unix() - 3600
	testHGA.EndTime = time.Now().Unix() + 3600

	testHGA2.GroupID = testGroupID
	testHGA2.ActivityId = 1002
	testHGA2.StartTime = time.Now().Unix() - 3600
	testHGA2.EndTime = time.Now().Unix() + 3600

	conn, err2 := redis.Dial("tcp", "127.0.0.1:6379", redis.DialDatabase(8), redis.DialPassword(""))
	if err2 != nil {
		fmt.Println(err2.Error())
		os.Exit(1)
	}
	//conn.Do("DEL", testHGA.GetRedisKey())
	//conn.Do("DEL", testHGA2.GetRedisKey())
	conn.Close()
	//rkey := testHGA.GetRedisKey()
	//for i:=1; i < 200; i
	//testConn.Do("ZADD", rkey, )
	//fmt.Println("Setup Done")

}

func Test_UpdateScore(t *testing.T) {
	for i := 0; i < 200; i++ {
		_, e := testHGR.UpdateScore(testHGA, uint64(i), HGRankMember{
			AccountID:  fmt.Sprintf("200:1000:%d", i),
			PlayerName: fmt.Sprintf("Name:%d", i),
		})

		if e != nil {
			if e.Code() > 0 {
				t.Error("发生意外错误0 %s", e.Error())
			}
		}

		//fmt.Println("T", r, e)
	}
	testHGR.ForcePullAllScores()
	list, num, err := testHGR.GetAllScores()
	//fmt.Println(list)
	if err != nil {
		if err.Code() > 0 {
			t.Error("发生意外错误2 %s", err.Error())
		}
	}

	if num <= 0 {
		t.Error("发生意外错误3")
	}

	if list[0].Score != 199 {
		t.Error("发生意外错误3", list[0].Score)
	}
}

func Test_UpdateScoreSame(t *testing.T) {
	for i := 0; i < 5; i++ {
		_, e := testHGR2.UpdateScore(testHGA2, 3000, HGRankMember{
			AccountID:  fmt.Sprintf("200:1000:%d", i),
			PlayerName: fmt.Sprintf("Name:%d", i),
		})
		if e != nil {
			if e.Code() > 0 {
				t.Error("发生意外错误0 %s", e.Error())
			}
		}
		time.Sleep(time.Second)
		//fmt.Println("T", r, e)
	}
	testHGR2.ForcePullAllScores()
	list, num, err := testHGR2.GetAllScores()
	if err != nil {
		if err.Code() > 0 {
			t.Error("发生意外错误2 %s", err.Error())
		}
	}

	if num <= 0 {
		t.Error("发生意外错误3")
	}

	if !strings.HasSuffix(list[0].PlayerName, "Name:0") {
		t.Error("发生意外错误4")
	}

	//fmt.Println(list)
}

func Test_Rename(t *testing.T) {
	//fmt.Println("rename test, %d", testHGR.sid)
	testHGR.OnPlayerRename("200:1000:120", "Name:120", "Name:re:120", uint64(120))
	testHGR.ForcePullAllScores()
	list, num, err := testHGR.GetAllScores()

	assert.True(t, err == nil || err.Code() <= 0)
	assert.NotEqual(t, 0, num)
	assert.Equal(t, "Name:re:120", list[79].PlayerName)
}
