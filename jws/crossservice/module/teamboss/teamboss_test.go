package teamboss

import (
	//"fmt"
	"sync"
	"testing"

	CrossConfig "vcs.taiyouxi.net/jws/crossservice/config"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/jws/crossservice/util/http_util"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"

	"github.com/stretchr/testify/assert"
)

// TODO: 清除redis里的 tb_reward_log: gid
var (
	testGroupId1 uint32 = 10001
	testGroupId2 uint32 = 20002
	testGroupId3 uint32 = 90009
	level        uint32 = 100

	wg sync.WaitGroup
)

func init() {
	gamedata.DebugLoadLocalGamedata()
	debugSetCSConfig()
}

// debugSetConfig 设置CrossServer测试参数
func debugSetCSConfig() {
	// config
	cofgName := "src/vcs.taiyouxi.net/jws/crossservice/conf/config.toml"

	var common_cfg struct{ CommonCfg CrossConfig.CommonConfig }
	err := config.DebugLoadConfigToml(cofgName, &common_cfg)
	if err != nil {
		panic(err)
	}
	CrossConfig.Cfg = common_cfg.CommonCfg

	// etcd
	http_util.Init("127.0.0.1:2379", 100001)
	etcd.InitClient([]string{"http://127.0.0.1:2379/"})

	// redis
	rc := &csdb.RedisConfig{Server: "127.0.0.1:6379", Auth: "", DB: 15}
	rc_redis_down := &csdb.RedisConfig{Server: "127.0.0.1:26379", Auth: "", DB: 15}

	mCfg := make(map[uint32]*csdb.RedisConfig)
	for i := 1; i < 64; i++ {
		mCfg[uint32(i)] = rc
	}
	mCfg[404] = rc_redis_down

	csdb.SetupRedis(mCfg, true)
}

// 生成建房间的参数
func genCreateRoomParam(level uint32, name string, acid string) *ParamCreateRoom {
	param := new(ParamCreateRoom)

	info := helper.CreateRoomInfo{}
	info.RoomLevel = level
	info.JoinInfo.Name = name
	info.JoinInfo.AcID = acid
	info.JoinInfo.Sid = uint(testGroupId1)

	param.Info = info

	return param
}

// 生成加入房间的参数
func genJoinRoomParam(roomID string, level int, name string, acid string) *ParamJoinRoom {
	param := new(ParamJoinRoom)

	info := helper.JoinRoomInfo{}
	info.RoomID = roomID
	info.JoinInfo.Name = name
	info.JoinInfo.AcID = acid
	info.JoinInfo.Sid = uint(testGroupId2)
	info.JoinInfo.Level = level

	param.Info = info

	return param
}

func TestGenerator_NewModule(t *testing.T) {
	m := module.LoadModulesList[0].NewModule(1)

	assert.Equal(t, ModuleID, m.ModuleID())
	assert.Equal(t, uint32(1), m.GetGroupID())
}
