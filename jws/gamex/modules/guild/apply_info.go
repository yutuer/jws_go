package guild

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MaxGuildApply  = 20
	MaxPlayerApply = 10
)

// 公会的申请信息
type GuildApplyInfo struct {
	ApplyTime     int64                    `json:"aply_t"`
	PlayerInfo    helper.AccountSimpleInfo `json:"player_i"`
	AssignID      []string                 `json:"assign_id"`
	AssignTimes   []int64                  `json:"assign_times"`
	LastLeaveTime int64                    `json:"last_leave_time""`
}

type GuildApply struct {
	Guild     guild_info.GuildSimpleInfo    `json:"guild"`
	ApplyList [MaxGuildApply]GuildApplyInfo `json:"aply"`
	ApplyNum  int                           `json:"aply_n"`
}

func (ga *GuildApply) DBName() string {
	return fmt.Sprintf("%s:%s", Table_GuildApply, ga.Guild.GuildUUID)
}

func (ga *GuildApply) DBSave(cb redis.CmdBuffer) error {
	key := ga.DBName()
	return driver.DumpToHashDBCmcBuffer(cb, key, ga)
}

func (ga *GuildApply) DBDel(cb redis.CmdBuffer) error {
	return cb.Send("DEL", ga.DBName())
}

func (ga *GuildApply) DBLoad() error {
	key := ga.DBName()

	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, ga, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return nil
}

// 以下信息需要公会改变的时候，同步过来
func (ga *GuildApply) updateGuildInfo(g guild_info.GuildSimpleInfo) {
	ga.Guild.GuildID = g.GuildID
	ga.Guild.Name = g.Name
	ga.Guild.Level = g.Level
	ga.Guild.ApplyGsLimit = g.ApplyGsLimit
	ga.Guild.ApplyAuto = g.ApplyAuto
	ga.Guild.Notice = g.Notice
	ga.Guild.MemNum = g.MemNum
	ga.Guild.MaxMemNum = g.MaxMemNum
}

func (ga *GuildApply) updateApply(now_time int64) {
	tl := (int64)(gamedata.GetCommonCfg().GetGuildApplyTimeLimit())
	for i := ga.ApplyNum - 1; i >= 0; i-- {
		apply := ga.ApplyList[i]
		if apply.PlayerInfo.AccountID != "" {
			if now_time-apply.ApplyTime >= tl {
				ga._delApply(i)
			}
		}
	}
}

func (ga *GuildApply) _delApply(index int) {
	if index == ga.ApplyNum-1 {
		ga.ApplyList[index] = GuildApplyInfo{}
	} else {
		ga.ApplyList[index] = ga.ApplyList[ga.ApplyNum-1]
		ga.ApplyList[ga.ApplyNum-1] = GuildApplyInfo{}
	}
	ga.ApplyNum--
}

// PlayerApplyInfo----------------------------------------------------------------
// 玩家的申请信息
type PlayerApplyInfo2Client struct {
	GuildUuid   string
	GuildName   string
	GuildLvl    uint32
	GuildNotice string
	ApplyTime   int64
}

type PlayerApplyInfo struct {
	GuildUuid string `json:"guid"`
	ApplyTime int64  `json:"aply_t"`
}

type PlayerApply struct {
	AccountId string                          `json:"acid"`
	ApplyList [MaxPlayerApply]PlayerApplyInfo `json:"aply"`
	ApplyNum  int                             `json:"aply_n"`
}

func (pa *PlayerApply) DBName() string {
	return fmt.Sprintf("%s:%s", Table_PlayerGuildApply, pa.AccountId)
}

func (pa *PlayerApply) DBSave(cb redis.CmdBuffer) error {
	key := pa.DBName()
	return driver.DumpToHashDBCmcBuffer(cb, key, pa)
}

func (pa *PlayerApply) DBDel(cb redis.CmdBuffer) error {
	return cb.Send("DEL", pa.DBName())
}

func (pa *PlayerApply) DBLoad() error {
	key := pa.DBName()

	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, pa, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return nil
}

// 检查并更新玩家申请列表
func (pa *PlayerApply) checkPlayerGuildApply(guildUuid string, now_time int64) int {
	// 是否满
	if pa.ApplyNum >= (int)(gamedata.GetCommonCfg().GetGuildApplyNumLimit()) {
		return Err_Player_Apply_Max
	}
	// 是否已经申请过了
	for i := pa.ApplyNum - 1; i >= 0; i-- {
		aply := pa.ApplyList[i]
		if aply.GuildUuid == guildUuid {
			return Err_Player_Apply_Repeat
		}
	}
	return 0
}

func (pa *PlayerApply) hasPlayerGuildApply(guildUuid string, now_time int64) bool {
	// 是否已经申请过了
	for i := pa.ApplyNum - 1; i >= 0; i-- {
		aply := pa.ApplyList[i]
		if aply.GuildUuid == guildUuid {
			return true
		}
	}
	return false
}

func (pa *PlayerApply) updateApply(now_time int64) {
	tl := (int64)(gamedata.GetCommonCfg().GetGuildApplyTimeLimit())
	for i := pa.ApplyNum - 1; i >= 0; i-- {
		aply := pa.ApplyList[i]
		if now_time-aply.ApplyTime >= tl {
			pa._delApply(i)
		}
	}
}

func (pa *PlayerApply) _delApply(index int) {
	if index == pa.ApplyNum-1 {
		pa.ApplyList[index] = PlayerApplyInfo{}
	} else {
		pa.ApplyList[index] = pa.ApplyList[pa.ApplyNum-1]
		pa.ApplyList[pa.ApplyNum-1] = PlayerApplyInfo{}
	}
	pa.ApplyNum--
}

// ---------------------------------------------------------------------

func _updateApply(pa *PlayerApply, ga *GuildApply, now_time int64) {
	pa.updateApply(now_time)
	ga.updateApply(now_time)
}

func _addApply(pa *PlayerApply, ga *GuildApply, time int64,
	guid string, playerInfo helper.AccountSimpleInfo, assignID []string, assignTimes []int64,
	lastLeaveTime int64) {
	pa.ApplyList[pa.ApplyNum] = PlayerApplyInfo{
		GuildUuid: guid,
		ApplyTime: time,
	}
	pa.ApplyNum++

	ga.ApplyList[ga.ApplyNum] = GuildApplyInfo{
		ApplyTime:     time,
		PlayerInfo:    playerInfo,
		AssignID:      assignID,
		AssignTimes:   assignTimes,
		LastLeaveTime: lastLeaveTime,
	}
	logs.Debug("add applyList: %v", ga.ApplyList[ga.ApplyNum])
	ga.ApplyNum++
}

func _delApply(pa *PlayerApply, ga *GuildApply, guid, acid string) *helper.AccountSimpleInfo {
	for i := pa.ApplyNum - 1; i >= 0; i-- {
		aply := pa.ApplyList[i]
		if aply.GuildUuid == guid {
			pa._delApply(i)
		}
	}
	for i := ga.ApplyNum - 1; i >= 0; i-- {
		apply := ga.ApplyList[i]
		if apply.PlayerInfo.AccountID == acid {
			playerInfo := apply.PlayerInfo
			ga._delApply(i)
			return &playerInfo
		}
	}
	return nil
}

func (aw *ApplyWorker) _delPlayerAllApply(pa *PlayerApply) (
	*helper.AccountSimpleInfo, []*GuildApply, []string, []int64, int64) {

	var applyTime int64
	var res *helper.AccountSimpleInfo
	var lastLeaveTime int64
	assignID := []string{}
	assignTimes := []int64{}
	chgGuild := make([]*GuildApply, 0, pa.ApplyNum)
	for i := pa.ApplyNum - 1; i >= 0; i-- {
		aply := pa.ApplyList[i]
		if ga, ok := aw.guildApply[aply.GuildUuid]; ok {
			for j := ga.ApplyNum - 1; j >= 0; j-- {
				apply := ga.ApplyList[j]
				if apply.PlayerInfo.AccountID == pa.AccountId {
					ga._delApply(j)
					chgGuild = append(chgGuild, ga)
					if res == nil {
						res = &apply.PlayerInfo
						if apply.AssignID != nil {
							assignID = apply.AssignID
						}
						if apply.AssignTimes != nil {
							assignTimes = apply.AssignTimes
						}
						lastLeaveTime = apply.LastLeaveTime
						applyTime = apply.ApplyTime
					} else {
						if apply.ApplyTime > applyTime {
							res = &apply.PlayerInfo
						}
					}
					// 红点通知
					aw.m.noticeHasApply(aply.GuildUuid, ga.ApplyNum > 0)
					break
				}
			}
		}
	}
	delete(aw.playerApply, pa.AccountId)
	return res, chgGuild, assignID, assignTimes, lastLeaveTime
}

func (aw *ApplyWorker) _updatePlayerInfo(pa *PlayerApply,
	info *helper.AccountSimpleInfo,
	now_time int64) (chgGuild []*GuildApply) {

	tl := (int64)(gamedata.GetCommonCfg().GetGuildApplyTimeLimit())
	chgGuild = make([]*GuildApply, 0, pa.ApplyNum)
	for i := pa.ApplyNum - 1; i >= 0; i-- {
		aply := pa.ApplyList[i]
		if now_time-aply.ApplyTime >= tl {
			continue
		}
		if ga, ok := aw.guildApply[aply.GuildUuid]; ok {
			for j := ga.ApplyNum - 1; j >= 0; j-- {
				apply := &ga.ApplyList[j]
				if apply.PlayerInfo.AccountID == pa.AccountId {
					apply.PlayerInfo = *info
					chgGuild = append(chgGuild, ga)
					break
				}
			}
		}
	}
	return chgGuild
}
