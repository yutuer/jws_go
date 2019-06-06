package logics

import (
	"strconv"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	ma "vcs.taiyouxi.net/jws/gamex/models/market_activity"
	"vcs.taiyouxi.net/jws/gamex/modules/market_activity"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GetOpRank : 获取运营活动排行榜
//
func (p *Account) GetOpRankHandler(req *reqMsgGetOpRank, resp *rspMsgGetOpRank) uint32 {
	//活动是否开启
	if code, _ := GetActivityValid(p.AccountID.ShardId, req.ActivityType); code != 0 {
		return code
	}
	activityCfg := gamedata.GetHotDatas().Activity
	act := activityCfg.GetActivitySimpleInfo(uint32(req.ActivityType))
	if len(act) < 1 {
		return errCode.ActivityTimeOut
	}
	info := p.GetSimpleInfo()
	res := market_activity.GetModule(p.AccountID.ShardId).GetRank(uint32(req.ActivityType), p.AccountID.String(), &info)
	if res == nil {
		return 0
	}
	minRank := getMinRankShow(act[0].ActivityId)
	if res.Pos > minRank {
		resp.SelfRank = 0
	} else {
		resp.SelfRank = int64(res.Pos)
	}
	resp.SelfScore = res.Score
	resp.RankInfo = make([][]byte, 0)
	for i, item := range res.TopN {
		if i+1 > minRank {
			break
		}
		op := OpRankInfo{
			Acid:  item.Info.AccountID,
			Name:  item.Name,
			Score: item.Score,
			Rank:  int64(i + 1),
		}
		if p.AccountID.String() == item.Info.AccountID {
			resp.SelfRank = op.Rank
			resp.SelfScore = op.Score
		}
		resp.RankInfo = append(resp.RankInfo, encode(op))
	}
	logs.Debug("market activity rank info: %v", res)
	return 0
}

func getMinRankShow(activityID uint32) int {
	subCfg := gamedata.GetHotDatas().Activity.GetMarketActivitySubConfig(activityID)
	minRank := 0
	for _, cfg := range subCfg {
		cR := int(cfg.GetFCValue2())
		if cR > minRank {
			minRank = cR
		}
	}
	return minRank
}

// GetOpRankRewardInfo : 获取运营活动排行榜可获得的奖励
//
func (p *Account) GetOpRankRewardInfoHandler(req *reqMsgGetOpRankRewardInfo, resp *rspMsgGetOpRankRewardInfo) uint32 {

	activityCfg := gamedata.GetHotDatas().Activity
	code, act := GetActivityValid(p.AccountID.ShardId, req.ActivityType)
	if code != 0 {
		return code
	}

	subCfg := activityCfg.GetMarketActivitySubConfig(act.ActivityId)
	resp.OpRankReward = make([][]byte, 0)
	for _, cfg := range subCfg {
		rw := OpRankRewardInfo{}
		rw.RankMin = int64(cfg.GetFCValue1())
		rw.RankMax = int64(cfg.GetFCValue2())
		minCond, err := strconv.ParseInt(cfg.GetSFCValue1(), 10, 0)
		if err != nil {
			logs.Error("condition value error for cfg: %v", cfg)
		}
		rw.MinCond = minCond
		rw.RewardID = make([]string, 0)
		rw.RewardCount = make([]int64, 0)
		for _, item := range cfg.GetItem_Table() {
			rw.RewardID = append(rw.RewardID, item.GetItemID())
			rw.RewardCount = append(rw.RewardCount, int64(item.GetItemCount()))
		}
		resp.OpRankReward = append(resp.OpRankReward, encode(rw))
	}
	return 0
}

func GetActivityValid(sid uint, activityType int64) (uint32, *gamedata.HotActivityInfo) {
	activityCfg := gamedata.GetHotDatas().Activity
	act := activityCfg.GetActivitySimpleInfo(uint32(activityType))
	if len(act) < 1 {
		return errCode.ActivityTimeOut, nil
	}
	if len(act) > 1 {
		logs.Warn("multi market activity for activity type: %d", activityType)
	}
	pAct := activityCfg.GetActivitySimpleInfoById(uint32(act[0].ActivityParentID))
	if pAct == nil {
		return errCode.ActivityTimeOut, nil
	}
	nowT := time.Now().Unix()
	if nowT < pAct.StartTime || nowT >= pAct.EndTime {
		return errCode.ActivityTimeOut, nil
	}
	typ := ma.GetHotTypeByActType(pAct.ActivityType)
	if !game.Cfg.GetHotActValidData(sid, typ) {
		return errCode.ActivityTimeOut, nil
	}
	return 0, act[0]
}
