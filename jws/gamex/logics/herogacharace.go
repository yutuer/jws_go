package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/herogacharace"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// HeroGachaRaceChest : 限时名将获取宝箱奖励及信息
// 限时名将中，玩家达到一定分数时，可以领取宝箱，请求分为两种类型，type=1的情况是领取奖励，type=2的情况是更新当前宝箱领取状况信息, rsp返回此时宝箱领取状态，0为未领取，1为已领取

// reqMsgHeroGachaRaceChest 限时名将获取宝箱奖励及信息请求消息定义
type reqMsgHeroGachaRaceChest struct {
	Req
	ChestIndex int64 `codec:"_p2_"` // 请求的宝箱索引ID
}

// rspMsgHeroGachaRaceChest 限时名将获取宝箱奖励及信息回复消息定义
type rspMsgHeroGachaRaceChest struct {
	SyncRespWithRewards
}

// HeroGachaRaceChest 限时名将获取宝箱奖励及信息: 限时名将中，玩家达到一定分数时，可以领取宝箱，请求分为两种类型，type=1的情况是领取奖励，type=2的情况是更新当前宝箱领取状况信息, rsp返回此时宝箱领取状态，0为未领取，1为已领取
func (p *Account) HeroGachaRaceChest(r servers.Request) *servers.Response {
	req := new(reqMsgHeroGachaRaceChest)
	rsp := new(rspMsgHeroGachaRaceChest)

	initReqRsp(
		"Attr/HeroGachaRaceChestRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_No_Act
		Err_ChestIndexLv
		Err_Give
		Err_Cheat
	)

	heroGachaRaceInfo := p.Profile.GetHeroGachaRaceInfo()
	actId := gamedata.GetHGRCurrValidActivityId()
	if actId <= 0 {
		return rpcErrorWithMsg(rsp, Err_No_Act, "Err_No_Act")
	}
	heroGachaRaceInfo.CheckActivity(p.AccountID.String(), int64(actId))

	chestInfo := heroGachaRaceInfo.GetChestInfo()
	if req.ChestIndex < 0 || req.ChestIndex >= int64(len(chestInfo)) {
		return rpcErrorWithMsg(rsp, Err_ChestIndexLv, "Err_ChestIndexLv")
	}
	needScore, had := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHeroGachaRaceChestInfo(heroGachaRaceInfo.ActivityID, uint32(req.ChestIndex))
	if !had {
		return rpcErrorWithMsg(rsp, Err_No_Act, "Err_No_Act")
	}
	if !chestInfo[req.ChestIndex] &&
		int64(needScore) <= heroGachaRaceInfo.GetCurScore() {
		heroGachaRaceInfo.SetChestInfo(int(req.ChestIndex), true)
		// 给予奖励
		data := &gamedata.CostData{}
		reward := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHeroGachaRaceChestReward(heroGachaRaceInfo.ActivityID, uint32(req.ChestIndex))
		for _, r := range reward {
			data.AddItem(r.GetItemID(), r.GetItemNum())
		}
		give := &account.GiveGroup{}
		give.AddCostData(data)
		if !give.GiveBySyncAuto(p.Account, rsp, "HeroGachaRaceChest") {
			return rpcErrorWithMsg(rsp, Err_Give, "Err_Give")
		}
	} else {
		return rpcErrorWithMsg(rsp, Err_Cheat, "Err_Cheat")
	}
	rsp.OnChangeHeroGachaRace()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// HeroGachaRaceGet : 打开限时名将ui，获取排行等信息
// 打开限时名将ui，获取排行等信息
// reqMsgHeroGachaRaceGet 打开限时名将ui，获取排行等信息请求消息定义
type reqMsgHeroGachaRaceGet struct {
	Req
}

// rspMsgHeroGachaRaceGet 打开限时名将ui，获取排行等信息回复消息定义
type rspMsgHeroGachaRaceGet struct {
	SyncResp
	Names      []string `codec:"names"` // 玩家昵称
	ShardNames []string `codec:"sns"`   // 区服名
	Scores     []int64  `codec:"scs"`   // 玩家积分
}

// HeroGachaRaceGet 打开限时名将ui，获取排行等信息: 打开限时名将ui，获取排行等信息
func (p *Account) HeroGachaRaceGet(r servers.Request) *servers.Response {
	req := new(reqMsgHeroGachaRaceGet)
	rsp := new(rspMsgHeroGachaRaceGet)

	initReqRsp(
		"Attr/HeroGachaRaceGetRsp",
		r.RawBytes,
		req, rsp, p)

	ranks, num, err := herogacharace.Get(p.AccountID.ShardId).GetAllScores()
	if err != nil && err != herogacharace.WARN_ACTIVITY_NOT_READY {
		logs.Error("HeroGachaRaceGet GetAllScores err %v", err)
		return rpcErrorWithMsg(rsp, 1, fmt.Sprintf("HeroGachaRaceGet GetAllScores err %v", err))
	}
	rsp.Names = make([]string, num)
	rsp.ShardNames = make([]string, num)
	rsp.Scores = make([]int64, num)

	var inRank bool
	for i := 0; i < num; i++ {
		r := ranks[i]
		rsp.Names[i] = r.Member.PlayerName
		rsp.ShardNames[i] = r.ShardDisplayName
		rsp.Scores[i] = int64(r.Score)
		if r.Member.AccountID == p.AccountID.String() {
			p.Profile.GetHeroGachaRaceInfo().Rank = int64(r.Rank)
			inRank = true
		}
	}
	if !inRank {
		p.Profile.GetHeroGachaRaceInfo().Rank = 0
	}
	rsp.OnChangeHeroGachaRace()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) OnGachaRace(actId uint32, times int64) bool {
	acid := p.AccountID.String()
	hgr := p.Profile.GetHeroGachaRaceInfo()
	hgr.CheckActivity(acid, int64(actId))
	hgr.AddCurScore(int64(gamedata.GetHotDatas().HotLimitHeroGachaData.GetHGRConfig().GetGachaRacePoint()) *
		times)
	cfg := gamedata.GetHotDatas().Activity.GetActivitySimpleInfoById(actId)
	rank, err := herogacharace.Get(p.AccountID.ShardId).UpdateScore(
		herogacharace.HGRActivity{
			GroupID:    gamedata.GetHotDatas().Activity.GetShardGroup(uint32(p.AccountID.ShardId)),
			ActivityId: actId,
			StartTime:  cfg.StartTime,
			EndTime:    cfg.EndTime,
		}, uint64(hgr.GetCurScore()),
		herogacharace.HGRankMember{
			AccountID:  acid,
			PlayerName: p.Profile.Name,
		})
	if err != nil {
		logs.Warn("herogacharace UpdateScore act %d warn %v",
			actId, err)
		return false
	}
	hgr.Rank = int64(rank)
	return true
}
