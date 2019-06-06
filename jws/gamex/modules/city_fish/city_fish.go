package city_fish

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	db_counter_key = "CityFish_DB"
)

func genFishModule(sid uint) *CityFish {
	return &CityFish{
		shardId: sid,
	}
}

type CityFish struct {
	shardId uint
	w       worker
	gfs     *FishRewardInfo
}

func (f *CityFish) AfterStart(g *gin.Engine) {
}

func (f *CityFish) BeforeStop() {
}

func (f *CityFish) Start() {
	f.w.start(f)
	f.loadFishRewardInfo()
}

func (f *CityFish) Stop() {
	f.w.stop()
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (f *CityFish) CommandExec(cmd FishCmd) *FishRet {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	res_chan := make(chan FishRet, 1)
	cmd.resChan = res_chan
	errRet := &FishRet{}
	chann := f.w.cmd_chan
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("CityFish CommandExec chann full, cmd put timeout")
		return errRet
	}

	select {
	case res := <-res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("CityFish CommandExec apply <-res_chan timeout")
		return errRet
	}
}

func (fm *CityFish) loadFishRewardInfo() {
	if fm.gfs == nil {
		fm.gfs = &FishRewardInfo{}
		fm.gfs.dbLoad(fm.shardId)
	}
}

type FishLog struct {
	Time  int64  `json:""`
	Name  string `json:"n"`
	Item  string `json:"im"`
	Count uint32 `json:"c"`
}

type FishRewardInfo struct {
	NextRefTime     int64     `json:"nft"`
	RewardLeftCount []uint32  `json:"rlc"`
	RewardLeftSum   uint32    `json:"rls"`
	Logs            []FishLog `json:"logs"`
}

func (gf *FishRewardInfo) update(shard uint) bool {
	now_time := game.GetNowTimeByOpenServer(shard)
	if now_time < gf.NextRefTime {
		return false
	}
	// 刷新钓鱼奖励
	gf.NextRefTime = util.GetNextDailyTime(
		gamedata.GetCommonDayBeginSec(now_time), now_time)
	gf.RewardLeftCount, gf.RewardLeftSum = gamedata.FishRewardCount()
	gf.Logs = make([]FishLog, 0, 16)
	return true
}

func (f *FishRewardInfo) dbLoad(shardId uint) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), tableFishReward(shardId), f, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return err
}

func (f *FishRewardInfo) dbSave(shardId uint) error {
	cb := redis.NewCmdBuffer()

	if err := driver.DumpToHashDBCmcBuffer(cb, tableFishReward(shardId), f); err != nil {
		return fmt.Errorf("DumpToHashDBCmcBuffer err %v", err)
	}

	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		return fmt.Errorf("cant get redis conn")
	}

	if _, err := modules.DoCmdBufferWrapper(db_counter_key, db, cb, true); err != nil {
		return fmt.Errorf("DoCmdBuffer error %s", err.Error())
	}
	return nil
}
