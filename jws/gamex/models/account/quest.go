package account

import (
	"time"

	"fmt"

	"sort"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// Boss任务和每日任务需要更新
// Quest_Main, Quest_Branch, Quest_PVE_Boss, Quest_Story_No_Use_Now, Quest_Daily, Quest_QuestPoint
var is_quest_typ_daily = []bool{false, false, true, false, true, true}

func isQuestTypNeedDailyGet(typ int) bool {
	if typ < 0 || typ >= len(is_quest_typ_daily) {
		return false
	}
	return is_quest_typ_daily[typ]
}

type quest struct {
	Id               uint32             `json:"id"`
	Typ              uint32             `json:"typ"`
	Condition        gamedata.Condition `json:"c"`
	Can_receive_time int64              `json:"crt"`
	Receive_time     int64              `json:"rt"`
}

type quest_cr struct {
	Id               uint32 `json:"id"`
	Can_receive_time int64  `json:"crt"`
}

func (q *quest) IsVailed() bool {
	return q.Id == 0
}

func (q *quest) SetVailed() {
	q.Id = 0
}

func (q *quest) GetProgress(p *Account) (int, int) {
	data := gamedata.GetQuestData(q.Id)
	if data == nil {
		logs.Error("Quest %d GetProgress Err No Data", q.Id)
		return 0, 0
	}

	return GetConditionProgress(&q.Condition, p,
		data.GetFCType(),
		int64(data.GetFCValueIP1()),
		int64(data.GetFCValueIP2()),
		data.GetFCValueSP1(),
		data.GetFCValueSP2())
}

type hasClosedList map[uint32]uint32

type PlayerQuest struct {
	helper.SyncObj

	Has_closed          hasClosedList
	Daily_has_closed    []uint32
	Account7_has_closed []uint32

	Received     []quest
	Received_len int

	Can_receive []quest_cr

	LastRefreshN [helper.Quest_Typ_count]int64

	// 任务积分
	QuestPoint      int
	QuestUpdateTime int64

	// 新账号7天任务
	ClearAccount7DayQuest bool
	Account7DayQuestPoint int
}

type playerQuestInDB struct {
	HasClosed        []uint32                      `json:"c"`
	DailyHasClose    []uint32                      `json:"bc"`
	Account7HasClose []uint32                      `json:"a7c"`
	HasReceived      []quest                       `json:"r"`
	LastRefreshN     [helper.Quest_Typ_count]int64 `json:"lr"`
	// 任务积分
	QuestPoint      int   `json:"qp"`
	QuestUpdateTime int64 `json:"qut"`
	// 新账号7天任务
	ClearAccount7DayQuest bool `json:"clra7d"`
	Account7DayQuestPoint int  `json:"a7dp"`
}

func (p *PlayerQuest) ToDB() playerQuestInDB {
	db_info := playerQuestInDB{
		HasClosed:        make([]uint32, 0, len(p.Has_closed)),
		DailyHasClose:    p.Daily_has_closed,
		Account7HasClose: p.Account7_has_closed,
		HasReceived:      make([]quest, 0, p.Received_len),
	}
	for qid, _ := range p.Has_closed {
		db_info.HasClosed = append(db_info.HasClosed, qid)
	}
	sort.Sort(uutil.UInt32Slice(db_info.HasClosed))
	for i := 0; i < len(p.Received); i++ {
		if !p.Received[i].IsVailed() {
			db_info.HasReceived = append(db_info.HasReceived, p.Received[i])
		}
	}
	db_info.LastRefreshN = p.LastRefreshN
	db_info.QuestPoint = p.QuestPoint
	db_info.QuestUpdateTime = p.QuestUpdateTime
	db_info.ClearAccount7DayQuest = p.ClearAccount7DayQuest
	db_info.Account7DayQuestPoint = p.Account7DayQuestPoint

	return db_info
}

func (p *PlayerQuest) FromDB(data *playerQuestInDB) error {
	p.Received = data.HasReceived
	p.Received_len = len(data.HasReceived)
	p.Has_closed = make(hasClosedList, len(data.HasClosed))
	for _, qid := range data.HasClosed {
		p.Has_closed[qid] = qid
	}
	p.Daily_has_closed = data.DailyHasClose
	p.Account7_has_closed = data.Account7HasClose
	p.LastRefreshN = data.LastRefreshN
	p.QuestPoint = data.QuestPoint
	p.QuestUpdateTime = data.QuestUpdateTime
	p.ClearAccount7DayQuest = data.ClearAccount7DayQuest
	p.Account7DayQuestPoint = data.Account7DayQuestPoint
	return nil
}

func (p *PlayerQuest) RegCondition(c *gamedata.PlayerCondition) {
	for i := 0; i < len(p.Received); i++ {
		if !p.Received[i].IsVailed() {
			c.RegCondition(&p.Received[i].Condition)
		}
	}
}

func (p *PlayerQuest) RegOneCondition(qidx int, c *gamedata.PlayerCondition) {
	c.RegCondition(&p.Received[qidx].Condition)
}

func (p *PlayerQuest) GetQuestPoint(nowT int64) int {
	if !gamedata.IsSameDayCommon(p.QuestUpdateTime, nowT) {
		p.QuestUpdateTime = nowT
		p.QuestPoint = 0
	}
	return p.QuestPoint
}

func (p *PlayerQuest) AddQuestPoint(a *Account, padd int, nowT int64, qid uint32) {
	if !gamedata.IsSameDayCommon(p.QuestUpdateTime, nowT) {
		p.QuestUpdateTime = nowT
		p.QuestPoint = 0
	}
	old := p.QuestPoint
	p.QuestPoint += padd
	// log
	logiclog.LogPoint(a.AccountID.String(), a.Profile.CurrAvatar, a.Profile.GetCorp().GetLvlInfo(),
		a.Profile.ChannelId, true, padd, old, p.QuestPoint, fmt.Sprintf("quest.%d", qid),
		func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
}

func (p *PlayerQuest) reallocReceived() {
	if p.Received_len == 0 {
		p.Received = p.Received[0:0]
		return
	}

	if len(p.Received) == p.Received_len {
		return
	}
	/*
		- 11, 2, 0, 3, 0, 0, 44, 5, 0, 0, 4
		- To
		- 11, 2, 4, 3, 5, 44
	*/
	for i := 0; i < len(p.Received); i++ {
		if p.Received[i].IsVailed() {
			j := len(p.Received) - 1
			for ; j >= 0 && p.Received[j].IsVailed(); j-- {
				// 去除末尾的空值
				p.Received = p.Received[:len(p.Received)-1]
			}

			// 说明扫描完毕
			if j < 0 || i > j {
				return
			}

			// 交换
			if i != j {
				p.Received[i] = p.Received[j]
				p.Received[j].SetVailed()

				p.Received = p.Received[:len(p.Received)-1]
			}
		}
	}
}

func (p *PlayerQuest) GetReceivedQuest() []quest {
	return p.Received[:]
}

func (p *PlayerQuest) GetCanReceivedQuest() []quest_cr {
	return p.Can_receive[:]
}

// For Sync To Client
func (p *PlayerQuest) GetDailyTaskClosed() []uint32 {
	return p.Daily_has_closed[:]
}

func (p *PlayerQuest) GetAccount7Closed() []uint32 {
	return p.Account7_has_closed[:]
}

func (p *PlayerQuest) AddReceived(qid uint32, can_rec_time int64, typ uint32, cond *gamedata.Condition) (qidx int) {
	// TODO 内存优化
	p.Received = append(p.Received, quest{qid, typ, *cond, can_rec_time, time.Now().Unix()})
	p.Received_len += 1
	qidx = len(p.Received) - 1
	return
}

func (p *PlayerQuest) GetReceived(qidx int) (*quest, bool) {
	if qidx >= len(p.Received) {
		return nil, false
	}

	if p.Received[qidx].IsVailed() {
		return nil, false
	}

	return &p.Received[qidx], true
}

func (p *PlayerQuest) IsHasReceived(qid uint32) bool {
	// TBD 当已接任务列表会多于20个时， 使用sort的search优化查询
	for i := 0; i < len(p.Received); i++ {
		if !p.Received[i].IsVailed() {
			if qid == p.Received[i].Id {
				return true
			}
		}
	}
	return false
}

func (p *PlayerQuest) GetCanReceive(can_idx int) (uint32, int64, bool) {
	if can_idx >= len(p.Can_receive) {
		return 0, 0, false
	}
	qcr := p.Can_receive[can_idx]
	return qcr.Id, qcr.Can_receive_time, true
}

func (p *PlayerQuest) Finish(qidx int) bool {
	if qidx >= len(p.Received) {
		return false
	}

	if p.Received[qidx].IsVailed() {
		return false
	}

	qid := p.Received[qidx].Id
	p.setHasClosed(qid)
	p.Received[qidx].Id = 0
	p.Received_len -= 1
	quest_data := gamedata.GetQuestData(qid)

	// 这个只是PVE Boss用
	if int(quest_data.GetQuestType()) == helper.Quest_Daily ||
		int(quest_data.GetQuestType()) == helper.Quest_PVE_Boss ||
		int(quest_data.GetQuestType()) == helper.Quest_QuestPoint {
		p.Daily_has_closed = append(p.Daily_has_closed, qid)
	}
	if int(quest_data.GetQuestType()) == helper.Quest_QuestPoint_7Day ||
		int(quest_data.GetQuestType()) == helper.Quest_7Day {
		p.Account7_has_closed = append(p.Account7_has_closed, qid)
	}
	return true
}

func (p *PlayerQuest) setHasClosed(qid uint32) {
	if p.Has_closed == nil {
		p.Has_closed = make(map[uint32]uint32, 256)
	}
	p.Has_closed[qid] = qid
}

func (p *PlayerQuest) IsHasClosed(qid uint32) bool {
	_, found := p.Has_closed[qid]
	return found
}

func (p *PlayerQuest) DebugSetCanReceiveForce(qid uint32) {
	// 先从Has_closed里清除
	delete(p.Has_closed, qid)

	// 强制加上
	p.Can_receive = append(p.Can_receive, quest_cr{qid, time.Now().Unix()})
}

func (p *PlayerQuest) CheckQuestReceiveConds(account *Account, data *ProtobufGen.Quest) bool {
	ac_conditions := data.GetAccCon_Table()

	//logs.Warn("UpdateCanReceiveList %d %v", data.GetQuestID(), data)
	for j := 0; j < len(ac_conditions); j++ {
		cond := ac_conditions[j]
		//logs.Warn("UpdateCanReceiveList CheckCondition %d %v", data.GetQuestID(), cond)
		if !CheckCondition(
			account,
			cond.GetACType(),
			int64(cond.GetACValueIP1()),
			int64(cond.GetACValueIP2()),
			cond.GetACValueSP(),
			"") {
			//logs.Warn("UpdateCanReceiveList CheckCondition false")
			return false
		}
	}
	return true
}

func (p *PlayerQuest) UpdateCanReceiveList(account *Account) (isUpdated bool) {
	if p.Can_receive == nil {
		p.Can_receive = make([]quest_cr, 0, 32)
	} else {
		p.Can_receive = p.Can_receive[0:0]
	}

	quest_to_check := gamedata.GetQuestNeedCheck()
	for i := 0; i < len(quest_to_check); i++ {
		q := quest_to_check[i]

		if p.IsHasClosed(q.GetQuestID()) {
			continue
		}

		if p.IsHasReceived(q.GetQuestID()) {
			continue
		}

		if p.CheckQuestReceiveConds(account, q) {
			isUpdated = true
			qtyp := int(q.GetQuestType())
			//logs.Trace("Add New Quest %d -- %d", q.GetQuestID(), qtyp)
			if isQuestTypNeedDailyGet(qtyp) { // 每日任务自动接收
				logs.Trace("Add New Daily Quest %d", q.GetQuestID())
				account.Profile.ReceiveQuestByQid(q.GetQuestID(), time.Now().Unix())
			} else { // 其他任务进入可接列表
				p.Can_receive = append(p.Can_receive, quest_cr{q.GetQuestID(), time.Now().Unix()})
			}
		}
	}
	return
}

// 重置每日任务呢，最好在UpdateCanReceiveList之前调用一次
func (p *PlayerQuest) DailyTaskReset(account *Account) (isUpdated bool) {
	for qtyp := 0; qtyp < helper.Quest_Typ_count; qtyp++ {
		if !isQuestTypNeedDailyGet(qtyp) {
			continue
		}

		nowT := account.Profile.GetProfileNowTime()
		isSameDay := gamedata.IsSameDayCommon(
			p.LastRefreshN[qtyp],
			nowT)

		//logs.Warn("LastRefreshN %d %d %s", p.LastRefreshN[qtyp], offsetDay, fresh_time_typ_daily[qtyp])
		if !isSameDay {
			isUpdated = true
			p.LastRefreshN[qtyp] = nowT
			// remove from has_close and reveived
			for _, qid := range gamedata.GetDailyTaskQuestId(qtyp) {
				//logs.Warn("%s del %d", account.Profile.DBName(), qid)
				delete(p.Has_closed, qid)
				for i := 0; i < len(p.Received); i++ {
					if qid == p.Received[i].Id {
						p.Received[i].Id = 0
						p.Received_len -= 1
						break
					}
				}

				if p.CheckQuestReceiveConds(account, gamedata.GetQuestData(qid)) {
					account.Profile.ReceiveQuestByQid(qid, time.Now().Unix())
				}
			}
			p.Daily_has_closed = []uint32{}
			//logs.Warn("%s boss daily task reset", account.Profile.DBName())
		}
	}
	return
}

func (p *PlayerQuest) UpdateAccount7DayQuest(profileCreateTime, now_time int64) {
	if p.ClearAccount7DayQuest { // 为了此方法只执行一次
		return
	}
	if now_time < gamedata.GetAccount7DayOverTime(profileCreateTime) {
		return
	}
	// 删除所有7天任务
	for qidx := 0; qidx < len(p.Received); qidx++ {
		q := p.Received[qidx]
		if !q.IsVailed() {
			quest_data := gamedata.GetQuestData(q.Id)
			if quest_data.GetQuestType() == helper.Quest_7Day ||
				quest_data.GetQuestType() == helper.Quest_QuestPoint_7Day {
				DelCondition(&q.Condition)
				p.Finish(qidx)
			}
		}
	}
	p.ClearAccount7DayQuest = true
}

func (p *PlayerQuest) GetAccount7DayQuestPoint() int {
	return p.Account7DayQuestPoint
}

func (p *PlayerQuest) AddAccount7DayQuestPoint(a *Account, padd int, reason string) {
	old := p.Account7DayQuestPoint
	p.Account7DayQuestPoint += padd
	// log
	logiclog.LogPoint(a.AccountID.String(), a.Profile.CurrAvatar, a.Profile.GetCorp().GetLvlInfo(),
		a.Profile.ChannelId, false, padd, old, p.Account7DayQuestPoint, reason,
		func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
}

func (p *Profile) InitQuest() {
	init_quests := gamedata.GetQuestInit()
	for i := 0; i < len(init_quests); i++ {
		if !p.GetQuest().IsHasReceived(init_quests[i]) {
			p.ReceiveQuestByQid(init_quests[i], time.Now().Unix())
		}
	}
}

func (p *Profile) ReceiveQuestByQid(qid uint32, can_rec_time int64) (qidx, code int) {
	qidx = -1
	code = 0

	const (
		_                = iota
		CODE_ID_Data_Err // 失败:ID对应的数据不存在
	)

	quest_data := gamedata.GetQuestData(qid)

	if quest_data == nil {
		code = CODE_ID_Data_Err
		return
	}

	player_quest := p.GetQuest()

	nc := NewCondition(
		quest_data.GetFCType(),
		int64(quest_data.GetFCValueIP1()),
		int64(quest_data.GetFCValueIP2()),
		quest_data.GetFCValueSP1(),
		quest_data.GetFCValueSP2())

	qidx = player_quest.AddReceived(qid, can_rec_time, quest_data.GetQuestType(), nc)
	player_quest.RegOneCondition(qidx, p.GetCondition())

	//logs.Trace("Quest %v", player_quest)
	return
}

//登入时自动领取任务
func (p *Profile) AutomationQuest(account *Account) {
	player_cond := p.GetQuest().GetReceivedQuest()
	auto_accep := gamedata.GetAutoAccep()
	idx := gamedata.GetAutoAccepIdx()

	for i := 0; i < len(idx); i++ {

		id := idx[i]
		if p.GetQuest().IsHasClosed(id) {
			continue
		}
		if p.GetQuest().IsHasReceived(id) {
			continue
		}
		if !p.GetQuest().CheckQuestReceiveConds(account, gamedata.GetQuestNeedCheckById(id)) {
			continue
		}
		p.ReceiveQuestByQid(id, time.Now().Unix())

	}
	for i := 0; i < len(auto_accep); i++ {
		for m := 0; m < len(player_cond); m++ {
			c := &player_cond[m].Condition
			if c.Ctyp == auto_accep[i] {
				UpdateCondition(c, 0, 0, "", "")

			}

		}

	}
}
