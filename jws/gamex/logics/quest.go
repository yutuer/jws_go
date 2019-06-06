package logics

import (
	"time"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestReceiveQuest struct {
	Req
	CIdx int `codec:"cidx"`
}

type ResponseReceiveQuest struct {
	SyncResp
}

func (p *Account) ReceiveQuestReq(r servers.Request) *servers.Response {
	const (
		_           = iota
		CODE_ID_Err // 失败:ID不存在
	)

	req := &RequestReceiveQuest{}
	resp := &ResponseReceiveQuest{}

	initReqRsp(
		"PlayerAttr/ReceiveQuestResponse",
		r.RawBytes,
		req, resp, p)

	logs.Trace("[%s]ReceiveQuest %d", p.AccountID, req.CIdx)

	_, code := p.ReceiveQuest(req.CIdx)

	if code != 0 {
		return rpcError(resp, uint32(code))
	} else {
		resp.OnChangeQuestAll()
		resp.mkInfo(p)
	}

	return rpcSuccess(resp)
}

type RequestFinishQuest struct {
	Req
	QIdx int `codec:"qidx"`
}

type ResponseFinishQuest struct {
	SyncRespWithRewards
}

func (p *Account) FinishQuestReq(r servers.Request) *servers.Response {
	req := &RequestFinishQuest{}
	resp := &ResponseFinishQuest{}
	initReqRsp(
		"PlayerAttr/FinishQuestResponse",
		r.RawBytes,
		req, resp, p)

	param := []int{req.QIdx}
	_, code := p.FinishQuest(param, resp, false)

	if code != 0 {
		return rpcWarn(resp, uint32(code))
	} else {
		// 处理教学ID提交 TODO By FanYang 以后所有请求都会由此逻辑 到时候要加在通用的地方
		// 同时保证只有包处理成功才会增加这些
		tids := req.GetTIDs()
		if tids != "" {
			logs.Trace("req tids %s", tids)
			err := p.Profile.AppendNewHand(tids)
			if err != nil {
				logs.SentryLogicCritical(p.AccountID.String(), err.Error())
			}
			resp.OnChangeNewHand()
		}

		resp.OnChangeQuestAll()
		resp.mkInfo(p)
	}

	return rpcSuccess(resp)
}

type RequestFinishManyQuests struct {
	Req
	QIdxs []int `codec:"qidxs"`
}

type ResponseFinishManyQuests struct {
	SyncRespWithRewards
}

func (p *Account) FinishManyQuestsReq(r servers.Request) *servers.Response {
	req := &RequestFinishManyQuests{}
	resp := &ResponseFinishManyQuests{}
	initReqRsp(
		"PlayerAttr/FinishManyQuestsResponse",
		r.RawBytes,
		req, resp, p)

	_, code := p.FinishQuest(req.QIdxs, resp, false)

	if code != 0 {
		return rpcWarn(resp, uint32(code))
	} else {
		// 处理教学ID提交 TODO By FanYang 以后所有请求都会由此逻辑 到时候要加在通用的地方
		// 同时保证只有包处理成功才会增加这些
		tids := req.GetTIDs()
		if tids != "" {
			logs.Trace("req tids %s", tids)
			err := p.Profile.AppendNewHand(tids)
			if err != nil {
				logs.SentryLogicCritical(p.AccountID.String(), err.Error())
			}
			resp.OnChangeNewHand()
		}
		resp.OnChangeQuestAll()
		resp.mkInfo(p)
	}

	return rpcSuccess(resp)
}

type questCanReceive2Client struct {
	QuestId uint32 `codec:"qid"`
	Idx     int    `codec:"idx"`
}

type questReceived2Client struct {
	QuestId  uint32   `codec:"qid"`
	Idx      int      `codec:"idx"`
	Progress int      `codec:"p"`
	All      int      `codec:"a"`
	ItemId   []string `codec:"item"`
	Count    []uint32 `codec:"count"`
}

type questBossClosed2Client struct {
	QuestId uint32   `codec:"qid"`
	ItemId  []string `codec:"item"`
	Count   []uint32 `codec:"count"`
}

func (p *Account) ReceiveQuest(can_received_idx int) (qidx, code int) {
	qidx = 0
	code = 0

	const (
		_            = iota
		CODE_IDX_Err // 失败:IDX对应的数据不存在
	)

	player_quest := p.Profile.GetQuest()
	player_quest.DailyTaskReset(p.Account)
	qid, crt, ok := player_quest.GetCanReceive(can_received_idx)
	if !ok {
		code = CODE_IDX_Err
		return
	}

	qidx, code = p.Profile.ReceiveQuestByQid(qid, crt)
	if code != 0 {
		code += 10 // 区分开错误码
	}
	return
}

func (p *Account) FinishQuest(qidxs []int, sync interfaces.ISyncRspWithRewards, is_force bool) (next_idx []int, code int) {
	code = 0
	next_idx = make([]int, 0, 8)

	player_quest := p.Profile.GetQuest()
	player_quest.DailyTaskReset(p.Account)
	nowT := p.Profile.GetProfileNowTime()
	for _, qidx := range qidxs {
		q, ok := player_quest.GetReceived(qidx)
		if !ok {
			code = errCode.QuestErrIDX
			return
		}

		qid := q.Id

		quest_data := gamedata.GetQuestData(qid)

		if quest_data == nil {
			code = errCode.QuestErrIDXData
			return
		}

		if code = _checkAccount7Day(p.Profile.CreateTime,
			p.Profile.GetProfileNowTime(), quest_data); code != 0 {
			return
		}

		logs.Trace("[%s]FinishQuest %v", p.AccountID, *q)

		progress, all := account.GetConditionProgress(
			&q.Condition,
			p.Account,
			quest_data.GetFCType(),
			int64(quest_data.GetFCValueIP1()),
			int64(quest_data.GetFCValueIP2()),
			quest_data.GetFCValueSP1(),
			quest_data.GetFCValueSP2())

		if progress < all && (!is_force) {
			code = errCode.QuestErrUnFinishCond
			return
		}

		// 完成后就把对应的条件清除掉
		account.DelCondition(&q.Condition)

		ok = player_quest.Finish(qidx)
		if !ok {
			code = errCode.QuestErrFinish
			return
		}

		questPoint := int(quest_data.GetActiveValue())
		if questPoint > 0 {
			player_quest.AddQuestPoint(p.Account, questPoint, nowT, qid)
			//41.任务积分大于等于P1
			p.updateCondition(account.COND_TYP_QuestPoint,
				0, 0, "", "", sync)
		}

		questPoint7Day := int(quest_data.GetSevenDaysActiveValue())
		if questPoint7Day > 0 {
			player_quest.AddAccount7DayQuestPoint(p.Account, questPoint7Day, fmt.Sprintf("quest.%d", qid))
		}

		quest_give_data := gamedata.GetQuestGiveData(qid)
		if quest_give_data == nil {
			logs.SentryLogicCritical(p.AccountID.String(), "GetQuestGiveData nil By %d", qid)
			code = errCode.QuestErrGiveData
			return
		}
		// 检查包裹满
		if quest_give_data.GetData().HasEquip {
			// 检查装备物品数量
			if p.BagProfile.GetEquipCount() >= gamedata.GetEquipCountUpLimit() {
				logs.SentryLogicCritical(p.AccountID.String(), "FinishQuest CODE_Bag_Full_Err for equip, quest %d", qid)
				code = errCode.QuestErrBugFull
				return
			}
		}
		if quest_give_data.GetData().HasJade {
			if p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
				logs.SentryLogicCritical(p.AccountID.String(), "FinishQuest CODE_Bag_Full_Err for jade, quest %d", qid)
				code = errCode.QuestErrBugFull
				return
			}
		}

		givesData := gamedata.CostData{}
		acID := p.AccountID.String()
		pRander := p.GetRand()
		if quest_data.GetQuestType() == helper.Quest_Daily {
			givesData.SetGST(gamedata.GST_DailyTask)
		}
		for i := 0; i < len(quest_give_data.Ids); i++ {
			id := quest_give_data.Ids[i]
			c := quest_give_data.Counts[i]
			if !gamedata.IsFixedIDItemID(id) {
				for j := 0; j < int(c); j++ {
					data := gamedata.MakeItemData(acID, pRander, id)
					givesData.AddItemWithData(id, *data, 1)
				}
			} else {
				givesData.AddItem(id, c)
			}
		}

		g := account.GiveGroup{}
		g.AddCostData(&givesData)

		questRewardType := "QuestReward"
		if quest_data.GetQuestType() == helper.Quest_PVE_Boss {
			questRewardType = "PveBossQuestReward"
		}
		give_ok := g.GiveBySyncAuto(p.Account, sync, questRewardType)
		if !give_ok {
			code = errCode.QuestErrGive
			return
		}

		// logic log
		logiclog.LogQuestFinish(p.AccountID.String(), p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, qid,
			givesData.Item2Client, givesData.Count2Client,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

		// 自动接取新的任务
		next := gamedata.GetPostQuest(qid)

		if next != nil && len(next) > 0 {
			for _, next_id := range next {
				if p.Profile.GetQuest().IsHasClosed(next_id) {
					continue
				}
				if p.Profile.GetQuest().IsHasReceived(next_id) {
					continue
				}
				n_idx, code := p.Profile.ReceiveQuestByQid(next_id, time.Now().Unix())
				if code != 0 {
					code = code + 10 // 区分开不同的错误
					logs.SentryLogicCritical(p.AccountID.String(), "Receive next %d Err %d",
						next, code)
				} else {
					next_idx = append(next_idx, n_idx)
				}
			}
		}
	}
	p.Profile.AutomationQuest(p.Account)
	return
}

func _checkAccount7Day(profileCreateTime, now_time int64, qCfg *ProtobufGen.Quest) int {
	if qCfg.GetQuestType() == helper.Quest_7Day ||
		qCfg.GetQuestType() == helper.Quest_QuestPoint_7Day {
		if now_time >= gamedata.GetAccount7DayOverTime(profileCreateTime) {
			return errCode.Account7DayQuestTimeOut
		}
		if gamedata.GetCommonDayDiff(profileCreateTime, now_time)+1 <
			int64(gamedata.GetAccount7DayQuestOpenDay(qCfg.GetQuestID())) {
			return errCode.Account7DayQuestTimeNotYet
		}
	}
	return 0
}
