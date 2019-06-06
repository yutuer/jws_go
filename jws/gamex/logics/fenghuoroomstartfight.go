package logics

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FenghuoRoomStartFight : 房间开始战斗
// 房间开始战斗, 会返回muitlplay房间号

// reqMsgFenghuoRoomStartFight 房间开始战斗请求消息定义
type reqMsgFenghuoRoomStartFight struct {
	Req
	RoomType    int64 `codec:"_p1_"` // Room类型
	SubLevelIdx int64 `codec:"_p2_"` // 进入战斗的小关卡的数字ID 0-7
}

// rspMsgFenghuoRoomStartFight 房间开始战斗回复消息定义
type rspMsgFenghuoRoomStartFight struct {
	SyncResp
	RoomID      string `codec:"_p1_"` // Muitlplay房间号(不用了,改在ReadyForLevel处理了)
	LevelInfoID string `codec:"_p2_"` // 服务器随机决定,当前关卡应该是哪一组出兵方案
	Rewards     []byte `codec:"_p3_"` // 提前生成的奖励，用于显示
}

type Reward struct {
	ItemID string `codec:"id"`
	Count  uint32 `codec:"c"`
	Data   string `codec:"d"`
}

type Rewards struct {
	Reward [][]byte `codec:"rs"`
}

func (r *Rewards) AppendReward(rw Reward) {
	r.Reward = append(r.Reward, encode(rw))
}

func (r *Rewards) ToData() []byte {
	return encode(*r)
}

// FenghuoRoomStartFight 房间开始战斗: 房间开始战斗, 会返回muitlplay房间号
func (p *Account) FenghuoRoomStartFight(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomStartFight)
	rsp := new(rspMsgFenghuoRoomStartFight)

	initReqRsp(
		"Attr/FenghuoRoomStartFightRsp",
		r.RawBytes,
		req, rsp, p)

	acID := p.AccountID.String()

	// logic imp begin
	if req.SubLevelIdx < 0 || req.SubLevelIdx > gamedata.FenghuoStageMaxNum {
		return rpcError(rsp, room.ROOM_ERR_UNKNOWN)
	}

	if p.Tmp.CurrRoomNum > 0 {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second)
		defer cancel()

		gameModeCounter := p.Profile.GetCounts()
		HasExtraReward := gameModeCounter.Has(gamedata.CounterTypeFengHuoFreeExtraReward, p.Account)

		resCode, leveldata, rewardShow := room.Get(p.AccountID.ShardId).StartFight(
			ctx,
			p.GetRand(),
			acID,
			p.Tmp.CurrRoomNum,
			int(req.SubLevelIdx),
			HasExtraReward,
		)
		if resCode != 0 {
			return rpcError(rsp, uint32(resCode))
		}

		rsp.LevelInfoID = leveldata.LevelInfoID

		var rw Rewards
		for i, item := range rewardShow.CostData2Client.Item2Client {
			if item != "" {
				count := rewardShow.CostData2Client.Count2Client[i]
				data := rewardShow.CostData2Client.Data2Client[i]
				rw.AppendReward(Reward{
					ItemID: item,
					Count:  count,
					Data:   data,
				})
			}
		}
		rsp.Rewards = rw.ToData()
	}

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

//////////////////////////////////////////////////////////////////
//trunk\Assets\Scripts\Network\NetLayer\Gen\gen\fenghuoroommkfightrewards.cs
// FenghuoRoomMkFightRewards : 获取房间战斗奖励
// 客户端在战斗胜利后调用，获得拿到的奖励

// reqMsgFenghuoRoomMkFightRewards 获取房间战斗奖励请求消息定义
type reqMsgFenghuoRoomMkFightRewards struct {
	Req
	RoomType    int64 `codec:"_p1_"` // Room类型
	SubLevelIdx int64 `codec:"_p2_"` // 进入战斗的小关卡的数字ID 0-7
}

// rspMsgFenghuoRoomMkFightRewards 获取房间战斗奖励回复消息定义
type rspMsgFenghuoRoomMkFightRewards struct {
	SyncRespWithRewards
}

// FenghuoRoomMkFightRewards 获取房间战斗奖励: 客户端在战斗胜利后调用，获得拿到的奖励
func (p *Account) FenghuoRoomMkFightRewards(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomMkFightRewards)
	rsp := new(rspMsgFenghuoRoomMkFightRewards)

	initReqRsp(
		"Attr/FenghuoRoomMkFightRewardsRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	if req.SubLevelIdx < 0 || req.SubLevelIdx >= gamedata.FenghuoStageMaxNum {
		return rpcError(rsp, room.ROOM_ERR_UNKNOWN)
	}

	if p.Tmp.CurrRoomNum > 0 {
		acID := p.AccountID.String()

		ctx, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second)
		defer cancel()

		resCode, reward, useExtra,
			rewardPower, battleHard, masterAcID :=
			room.Get(p.AccountID.ShardId).
				MakeReward(
					ctx,
					acID,
					p.Tmp.CurrRoomNum,
					int(req.SubLevelIdx),
				)
		if resCode != 0 {
			return rpcError(rsp, uint32(resCode))
		}

		if rewardPower <= 0 {
			rewardPower = 1
		}

		sc, hc := gamedata.GetFenghuoSCHC(uint32(battleHard))
		//消耗SC
		cg := &account.CostGroup{}
		hassc := cg.AddSc(p.Account, helper.SC_Money, int64(sc*uint32(rewardPower)))

		//消耗HC
		if masterAcID != acID {
			//只有主机消耗硬通货
			hc = 0
		} else {
			hc = uint32(rewardPower-1) * hc
		}

		hashc := cg.AddHc(p.Account, int64(hc))

		if hassc && hashc {
			success := cg.CostBySync(p.Account, rsp, "FenghuoRewardSubLevel")
			if !success {
				//扣钱不成功,理论上不应该出现
				logs.Error("FenghuoRoomMkFightRewards no enough money!")
				return rpcError(rsp, room.ROOM_ERR_MKREWARD_NO_ENOUGH_MONEY)
			}
		}
		// FIXME by YZH 双方都战败后,需要离开处理,包括直接退出游戏的玩家
		// FIXME by YZH 胜利后,有机会获得免费复活次数
		// FIXME by YZH 在战斗中复活协议, 消耗钻石或者免费复活次数

		if reward != nil {
			if !account.GiveBySync(p.Account, reward.Gives(), rsp, "RoomMkRewards") {
				logs.SentryLogicCritical(acID, "FenghuoRoomMkFightRewards give Err")
				return rpcError(rsp, room.ROOM_ERR_GIVESYNC)
			} else {
				if useExtra {
					//消耗玩家前16场战斗的额外奖励
					gameModeCounter := p.Profile.GetCounts()
					if !gameModeCounter.Use(gamedata.CounterTypeFengHuoFreeExtraReward, p.Account) {
						logs.SentryLogicCritical(acID, "FenghuoRoomMkFightRewards game mode use Err")
						return rpcError(rsp, room.ROOM_ERR_GIVESYNC_EXTRAREWARD_COUNTER_CONSUME)
					}
				}
			}
		}
	}

	// logic imp end

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeFengHuoFreeExtraReward,
		1,
		p.Profile.GetProfileNowTime())

	rsp.mkInfo(p)

	// 条件更新
	p.updateCondition(account.COND_TYP_FengHuoSubLevelCount, 1, 0, "", "", rsp)

	return rpcSuccess(rsp)
}
