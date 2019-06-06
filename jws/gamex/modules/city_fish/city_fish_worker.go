package city_fish

import (
	"math/rand"

	"encoding/json"
	"fmt"

	"time"

	"strconv"
	"strings"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/city_broadcast"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	_ = iota
	CityFish_Cmd_Award
	CityFish_Cmd_Get_Info
	CityFish_Cmd_Get_Record
	CityFish_Cmd_Debug_Set_Global_Award
	CityFish_Cmd_Debug_Reset_Award
)

type FishCmd struct {
	Typ               int
	AName             string
	ARand             *rand.Rand
	AwardCount        int
	DebugRewardIdx    int
	DebugGlobalReward string
	sid               uint
	resChan           chan FishRet
}

type FishRet struct {
	AwardId           []uint32
	AwardItem         []string
	AwardCount        []uint32
	AwardLeftCount    int
	NextRefTime       int64
	GlobalRewardCount []uint32
	Logs              []FishLog
	Success           bool
}

type worker struct {
	waitter            util.WaitGroupWrapper
	cmd_chan           chan FishCmd
	cityBroadcast_chan chan broadcastMsg_wrapper
	save_chan          chan FishRewardInfo
}

func (w *worker) start(f *CityFish) {
	w.cmd_chan = make(chan FishCmd, 2048)
	w.cityBroadcast_chan = make(chan broadcastMsg_wrapper, 2048)
	w.save_chan = make(chan FishRewardInfo, 2048)

	w.waitter.Wrap(func() {
		for cmd := range w.cmd_chan {
			func() {
				//by YZH 这个让parent never dead, 应该如此吗？
				defer logs.PanicCatcherWithInfo("city_fish Worker Panic")
				cmd.sid = f.shardId
				w.processCommand(&cmd, f.gfs)
			}()
		}
		close(w.save_chan)
	})

	// 存db的goroutine，不需要关闭
	w.waitter.Wrap(func() {
		for fr := range w.save_chan {
			if err := fr.dbSave(f.shardId); err != nil {
				logs.Error("FishRewardInfo save err: %s", err.Error())
			}
		}
	})

	// 广播大奖状态，用单rountine发，由于gamex和chat之间是http，如果奖励变化太频繁的话，这块可能是瓶颈
	// 优化：gamex和chat之间socket
	go func() {
		for msg := range w.cityBroadcast_chan {
			bb, err := json.Marshal(msg.msg)
			if err != nil {
				logs.Error("Fish broadcast json %v err %v", msg.msg, err)
				continue
			}
			city_broadcast.Pool.UseRes2Send(
				city_broadcast.CBC_Typ_Fish,
				fmt.Sprintf("%d:%d", msg.gid, msg.shardId),
				string(bb),
				nil,
			)
		}
	}()
}

func (w *worker) stop() {
	close(w.cmd_chan)
	w.waitter.Wait()
}

func (w *worker) processCommand(c *FishCmd, gfs *FishRewardInfo) {
	if gfs.update(c.sid) {
		// save db
		w.send2db(gfs)
	}

	switch c.Typ {
	case CityFish_Cmd_Award:
		w.award(c, gfs)
	case CityFish_Cmd_Get_Info:
		w.getInfo(c, gfs)
	case CityFish_Cmd_Get_Record:
		w.getRecord(c, gfs)
	case CityFish_Cmd_Debug_Set_Global_Award:
		w.debugSetGlobal(c, gfs)
	case CityFish_Cmd_Debug_Reset_Award:
		w.debugResetGlobal(c, gfs)
	}
}

func (w *worker) award(c *FishCmd, fr *FishRewardInfo) {
	ret := FishRet{
		Success:    true,
		AwardId:    make([]uint32, 0, 10),
		AwardItem:  make([]string, 0, 10),
		AwardCount: make([]uint32, 0, 10),
	}
	isChg := false
	// for debug
	if c.DebugRewardIdx > 0 {
		if c.DebugRewardIdx > len(fr.RewardLeftCount) ||
			fr.RewardLeftCount[c.DebugRewardIdx-1] <= 0 {
			c.resChan <- ret
		}
		_giveReward(c.DebugRewardIdx-1, c, fr, &ret)
		isChg = true
	} else { // 正常流程
		ret.AwardLeftCount = c.AwardCount
		for i := 0; i < c.AwardCount; i++ {
			if fr.RewardLeftSum <= 0 {
				break
			}
			var sumWeight uint32
			weight := make([]uint32, 0, len(fr.RewardLeftCount))
			for _, lc := range fr.RewardLeftCount {
				sumWeight += lc
				weight = append(weight, sumWeight)
			}
			rd := c.ARand.Int31n(int32(sumWeight))
			for idx, wt := range weight {
				lc := fr.RewardLeftCount[idx]
				if lc > 0 && uint32(rd) < wt {
					_giveReward(idx, c, fr, &ret)
					break
				}
			}
			isChg = true
			ret.AwardLeftCount--
		}
	}
	ret.NextRefTime = fr.NextRefTime
	ret.GlobalRewardCount = fr.RewardLeftCount
	c.resChan <- ret

	if isChg {
		// 广播
		w.sendBroadcast(c.sid, fr)
		// save db
		w.send2db(fr)
	}
}

func _giveReward(idx int, c *FishCmd, fr *FishRewardInfo, ret *FishRet) {
	rCfg := gamedata.GetFishReward(idx)
	// 记log
	if rCfg.GetLogIDS() != "" {
		fr.Logs = append(fr.Logs, FishLog{
			Time: game.GetNowTimeByOpenServer(c.sid),
			Name: c.AName,
			Item: rCfg.GetShowItemID(),
		})
	}
	// 奖励减少
	fr.RewardLeftCount[idx] = fr.RewardLeftCount[idx] - 1
	fr.RewardLeftSum = fr.RewardLeftSum - 1
	// 跑马灯
	if rCfg.GetServerMsgID() > 0 {
		sysnotice.NewSysRollNotice(fmt.Sprintf("%d:%d", game.Cfg.Gid, c.sid),
			int32(rCfg.GetServerMsgID())).
			AddParam(sysnotice.ParamType_RollName, c.AName).
			AddParam(sysnotice.ParamType_ItemId, rCfg.GetShowItemID()).Send()
	}
	// 随机物品
	for i := 0; i < int(rCfg.GetLootDataAmount()); i++ {
		gives, err := gamedata.LootItemGroupRand(c.ARand, rCfg.GetLootDataID())
		if err != nil || !gives.IsNotEmpty() {
			logs.Error("fish loot err")
		} else {
			for i := 0; i < gives.Len(); i++ {
				ok, itemID, count, _, _ := gives.GetItem(i)
				if ok {

					ret.AwardItem = append(ret.AwardItem, itemID)
					ret.AwardCount = append(ret.AwardCount, count)
				}
			}
		}
	}
	for i := 0; i < int(rCfg.GetLootData2Amount()); i++ {
		gives, err := gamedata.LootItemGroupRand(c.ARand, rCfg.GetLootData2ID())
		if err != nil || !gives.IsNotEmpty() {
			logs.Error("fish loot2 err")
		} else {
			for i := 0; i < gives.Len(); i++ {
				ok, itemID, count, _, _ := gives.GetItem(i)
				if ok {

					ret.AwardItem = append(ret.AwardItem, itemID)
					ret.AwardCount = append(ret.AwardCount, count)
				}
			}
		}
	}
	ret.AwardId = append(ret.AwardId, rCfg.GetRewardID())
	if rCfg.GetItemID() != "" && rCfg.GetAmount() > 0 {
		ret.AwardItem = append(ret.AwardItem, rCfg.GetItemID())
		ret.AwardCount = append(ret.AwardCount, rCfg.GetAmount())
	}
}

func (w *worker) getInfo(c *FishCmd, fr *FishRewardInfo) {
	ret := FishRet{
		Success: true,
	}
	ret.NextRefTime = fr.NextRefTime
	ret.GlobalRewardCount = fr.RewardLeftCount
	c.resChan <- ret
}

func (w *worker) getRecord(c *FishCmd, fr *FishRewardInfo) {
	ret := FishRet{
		Success: true,
	}
	ret.Logs = fr.Logs
	c.resChan <- ret
}

func (w *worker) sendBroadcast(sid uint, fr *FishRewardInfo) {
	msg := broadcastMsg_wrapper{
		gid:     uint(game.Cfg.Gid),
		shardId: sid,
		msg: broadcastMsg{
			FishNextRefTime:     fr.NextRefTime,
			FishRewardLeftCount: fr.RewardLeftCount,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	select {
	case w.cityBroadcast_chan <- msg:
	case <-ctx.Done():
		logs.Error("CityFish sendBroadcast chann full")
	}
}

func (w *worker) send2db(fr *FishRewardInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	select {
	case w.save_chan <- *fr:
	case <-ctx.Done():
		logs.Error("CityFish send2db chann full")
	}
}

type broadcastMsg_wrapper struct {
	gid     uint
	shardId uint
	msg     broadcastMsg
}
type broadcastMsg struct {
	FishNextRefTime     int64    `json:"nreft"`
	FishRewardLeftCount []uint32 `json:"rlc"`
}

func (w *worker) debugSetGlobal(c *FishCmd, fr *FishRewardInfo) {
	ret := FishRet{}
	if c.DebugGlobalReward == "" {
		c.resChan <- ret
	}
	for idx, v := range strings.Split(c.DebugGlobalReward, ",") {
		c, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		fr.RewardLeftCount[idx] = uint32(c)
	}
	fr.RewardLeftSum = 0
	for _, c := range fr.RewardLeftCount {
		fr.RewardLeftSum += c
	}
	ret.Success = true
	ret.NextRefTime = fr.NextRefTime
	ret.GlobalRewardCount = fr.RewardLeftCount
	c.resChan <- ret
}

func (w *worker) debugResetGlobal(c *FishCmd, fr *FishRewardInfo) {
	ret := FishRet{
		Success: true,
	}
	fr.NextRefTime = game.GetNowTimeByOpenServer(c.sid) + 5*util.MinSec
	w.sendBroadcast(c.sid, fr)
	c.resChan <- ret
}
