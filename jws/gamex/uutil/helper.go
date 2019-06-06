package uutil

import (
	"fmt"

	"time"

	"os"

	"math/rand"

	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
)

const (
	Lang_HANS = "zh-Hans"
	Lang_HMT  = "zh-HMT"
	Lang_EN   = "en"
	Lang_VN   = "vi"
	Lang_KO   = "ko"
	Lang_JA   = "ja"
	Lang_TH   = "th"
)

const CHEAT_INT_MAX = 10000

var (
	TimerSec = timingwheel.NewTimingWheel(time.Second, 120)
	TimerMS  = timingwheel.NewTimingWheel(10*time.Millisecond, 100*60)
)

func init() {
	rand.Seed(time.Now().Unix())
}

func MetricRedisByAccount(id db.Account, typ, value string) {
	name := fmt.Sprintf("redis.%d.%d.%s.%s", id.GameId, id.ShardId, typ, "time")
	metrics.SimpleSend(name, value)
}

func MetricRedis(sid uint, typ, value string) {
	name := fmt.Sprintf("redis.%d.%d.%s.%s", game.Cfg.Gid, sid, typ, "time")
	metrics.SimpleSend(name, value)
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func IsOverseaVer() bool {
	return game.Cfg.Lang != Lang_HANS
}

func IsHMTVer() bool {
	return game.Cfg.Lang == Lang_HMT
}

func IsVNVer() bool {
	return game.Cfg.Lang == Lang_VN
}

func IsKOVer() bool {
	return game.Cfg.Lang == Lang_KO
}

func IsThVer() bool {
	return game.Cfg.Lang == Lang_TH
}

func IsJAVer() bool {
	return game.Cfg.Lang == Lang_JA
}

type UInt32Slice []uint32

func (p UInt32Slice) Len() int           { return len(p) }
func (p UInt32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p UInt32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
