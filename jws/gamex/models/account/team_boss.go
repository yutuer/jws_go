package account

import (
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	BOX_POS_TYPE_STORAGE = iota // 仓库预存
	BOX_POS_TYPE_1              // 倒计时1
	BOX_POS_TYPE_2              // 倒计时2
	BOX_POS_TYPE_3              // 倒计时3
	BOX_POS_TYPE_COUNT
)
const (
	TEAM_BOSS_DIFF_0 = iota
	TEAM_BOSS_DIFF_1 //组队boss难度1
	TEAM_BOSS_DIFF_2 //组队boss难度2
	TEAM_BOSS_DIFF_3 //组队boss难度3
	TEAM_BOSS_DIFF_4 //组队boss难度4
	TEAM_BOSS_DIFF_5 //组队boss难度5

	TEAM_MEMBER_NUM = 2 //组队boss队伍人数

	TEAM_SETTING_OPEN     = 0 //邀请加入设定公开
	TEAM_SETTING_INV_ONLY = 1 //设定仅限邀请
	TEAM_SETTING_FULL     = 2 //队伍已满

	TEAM_REDBOX_UNTICK = 0 //没勾选必掉红宝箱
	TEAM_REDBOX_TICK   = 1 //勾选必掉红宝箱

	TEAM_LEAVE_BY_SELF = 0 //自己离开队伍
	TEAM_LEAVE_BY_KICK = 1 //被踢出队伍

	TEAM_UNREADY = 0 //队伍没准备
	TEAM_READY   = 1 //队伍准备

	TEAM_HERO_POS_LEFT  = 0 //队伍中选将左边的
	TEAM_HERO_POS_RIGHT = 1 //队伍中选将右边的

)

type TeamBossStorageInfo struct {
	OpenTimes       int                                 `json:"open_t"`   // 花钻石开宝箱的次数
	BoxList         [BOX_POS_TYPE_COUNT]TeamBossBoxInfo `json:"box_list"` // 宝箱数组
	BoxCtrlTimes    uint32                              `json:"box_ctrl"` //暗控未得宝箱次数
	LastRefreshTime int64                               `json:"refresh"`
}

type TeamBossTeamInfo struct {
	NowTeamID         string            `json:"now_team_id"`
	TeamBossLeaveInfo TeamBossLeaveInfo `json:"tb_leave"`
	GlobalRoomId      string            `json:"battle_id"`
}

type TeamBossLeaveInfo struct {
	LeaveType   int    `json:"l_type"` //是被踢还是主动离开的
	LeaveTeamId string `json:"l_tid"`
	LeaveTime   int64  `json:"l_time"`
}

type TeamBossBoxInfo struct {
	TBBoxId      string `json:"tb_boxid"` //组队BOSS宝箱ID
	TBBoxEndTime int64  `json:"tb_btime"` //组队BOSS宝箱到期截止时间
}

func (t *TeamBossStorageInfo) ResetBox(index int) {
	t.BoxList[index] = TeamBossBoxInfo{}
}

func (t *TeamBossStorageInfo) ResetControlTimes() {
	t.BoxCtrlTimes = 0
}

func (t *TeamBossStorageInfo) IncreaseControlTimes() {
	t.BoxCtrlTimes++
}

func (t *TeamBossStorageInfo) GetTBossBoxCount() int {
	count := 0
	for _, box := range t.BoxList {
		if box.TBBoxId != "" {
			count++
		}
	}
	logs.Debug("<TBoss> get tboss box count : %v", count)
	return count
}

//cheat用的直接加宝箱方法
func (t *TeamBossStorageInfo) SetNewBoxForCheat(boxId string, index int, endtime int64) {
	logs.Debug("cheat set new box boxid: %v index: %v endtime: %v", boxId, index, endtime)
	t.BoxList[index] = TeamBossBoxInfo{
		TBBoxId:      boxId,
		TBBoxEndTime: endtime,
	}
	logs.Debug("cheat tbbox now boxlist %v", t.BoxList)
}

func (t *TeamBossStorageInfo) SetTBBoxNormalForCheat(boxId string) {
	t.AddNewBox(boxId)
}

func (t *TeamBossStorageInfo) AddNewBox(boxId string) {
	logs.Debug("<TBoss> add new box id: %v", boxId)
	for i, box := range t.BoxList {
		//有格子直接加进去
		if box.TBBoxId == "" && i != BOX_POS_TYPE_STORAGE {
			t.BoxList[i] = TeamBossBoxInfo{
				TBBoxId:      boxId,
				TBBoxEndTime: int64(*gamedata.GetTBBoxDataByBoxId(boxId).OpenNeedTime) + time.Now().Unix(),
			}
			logs.Debug("<TBoss> add new box list: %v", t.BoxList)
			return
		}
	}
	//没有格子替换预存位置
	t.BoxList[BOX_POS_TYPE_STORAGE] = TeamBossBoxInfo{
		TBBoxId:      boxId,
		TBBoxEndTime: int64(*gamedata.GetTBBoxDataByBoxId(boxId).OpenNeedTime),
	}
	logs.Debug("<TBoss> add new box list: %v", t.BoxList)

}

func (t *TeamBossStorageInfo) TryDailyReset(nowTime int64) {
	if !gamedata.IsSameDayCommon(nowTime, t.LastRefreshTime) {
		t.LastRefreshTime = nowTime
		t.OpenTimes = 0
	}
}

func (t *TeamBossStorageInfo) MoveStorageBox(index int, nowTime int64) {
	if t.BoxList[BOX_POS_TYPE_STORAGE].TBBoxId != "" && t.BoxList[index].TBBoxId == "" {
		t.BoxList[index], t.BoxList[BOX_POS_TYPE_STORAGE] = t.BoxList[BOX_POS_TYPE_STORAGE], t.BoxList[index]
		t.BoxList[index].TBBoxEndTime = nowTime + gamedata.GetTBBoxNeedTime(t.BoxList[index].TBBoxId)
	}
}

func (lt *TeamBossLeaveInfo) SetTeamBossLeaveInfo(leaveType int, teamId string, leaveTime int64) {
	if teamId == "" || leaveTime == 0 {
		logs.Error("set teamboss leave info err teamid: %v, leavetime: %v", teamId, leaveTime)
		return
	}
	lt.LeaveType = leaveType
	lt.LeaveTeamId = teamId
	lt.LeaveTime = leaveTime
	logs.Debug("set teamboss leave info success teamid: %v, leavetime: %v", teamId, leaveTime)
}

func (lt *TeamBossLeaveInfo) IsCanJoinTBossTeamNow(teamID string, leaveType int, nowTime int64) bool {
	time := nowTime - lt.LeaveTime
	logs.Debug("<TBoss> join room time space is %v", time)
	if teamID == lt.LeaveTeamId {
		if leaveType == TEAM_LEAVE_BY_SELF {
			if time >= int64(gamedata.BoxCfg.TeamBackTime) {
				//logs.Debug("can join tbteam because of time: %v",time)
				return true
			}
		}
		if leaveType == TEAM_LEAVE_BY_KICK {
			if time >= int64(gamedata.BoxCfg.GoOutTeamTime) {
				return true
			}
		}

		logs.Debug("cannot join tbteam because of time: %v", time)
		return false
	}
	return true
}
