package guild

import (
	"vcs.taiyouxi.net/jws/gamex/models/clienttag"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	_ "vcs.taiyouxi.net/jws/gamex/modules/guild/update"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Guild struct {
	dbkey        db.ProfileDBKey
	dirtiesCheck map[string]interface{}
	Ver          int64 `redis:"version"`
	CreateTime   int64 `redis:"createtime"`

	GuildPosition      int             `redis:"pos"`
	GuildUUID          string          `redis:"guid"`
	GuildName          string          `redis:"gname"`
	NextEnterGuildTime int64           `redis:"leavetime"`
	GuildAssignInfo    GuildAssignInfo `redis:"guild_assign_info"`

	Post         string `redis:"post"`
	PostOverTime int64  `redis:"postot"`

	GuildSignTimesToday         int   `redis:"signc"`
	GuildSignTimesTodayLastTime int64 `redis:"signt"`

	// 最近一次离开公会的时间, 由于有玩家离线被T的情况, 所以还写了个离线存库的方式 @ account_leave_guild
	// 玩家如果是被T的, 会同时更新在线内存和redis, 风险是内存更新失败导致redis数据被覆盖
	LastLeaveGuildTime int64 `redis:"lastleavetime"`

	HasApplyCanApprove bool
	RedPacketInfo      GuildRedPacketInfo `redis:"grpi"`
	WorshipInfo        GuildWorshipInfo   `redis:"worship_info"`
}

type GuildAssignInfo struct {
	AssignID    []string `json:"assign_id"`
	AssignTimes []int64  `json:"assign_times"`
}

const (
	_ = iota
	Err_Apply_Max
	Err_Apply_Repeat
)

func (g *Guild) OnAfterLogin(ac db.Account, tag *clienttag.ClientTag) {
	acid := ac.String()
	guuid := guild.GetPlayerGuild(acid)

	if guuid == "" {
		g.ClearGuild()
		return
	}
	if g.GuildUUID != guuid {
		g.GuildUUID = guuid
		tag.SetTag(clienttag.Tag_GuildIn, 1)
	}
	guildInfo, errRet := guild.GetModule(ac.ShardId).GetGuildInfo(g.GuildUUID)
	if errRet.HasError() {
		return
	}
	isInGuild := false
	for i := 0; i < int(guildInfo.Base.MemNum); i++ {
		if guildInfo.Members[i].AccountID == acid {
			g.GuildPosition = guildInfo.Members[i].GuildPosition
			g.GuildName = guildInfo.Base.Name
			isInGuild = true
			break
		}
	}
	if isInGuild && gamedata.CheckApprovePosition(g.GuildPosition) {
		aplys := guild.GetModule(ac.ShardId).GetGuildApplyInfo(g.GuildUUID, acid)
		if len(aplys) > 0 {
			g.HasApplyCanApprove = true
		} else {
			g.HasApplyCanApprove = false
		}
	}
}

func (g *Guild) GetCurrPosition() int {
	return g.GuildPosition
}

func (g *Guild) GetCurrGuildUUID() string {
	return g.GuildUUID
}

// 返回的是可用次数
func (g *Guild) GetGuildSignCount(vipLv int, nowT int64) int {
	vd := gamedata.GetVIPCfg(vipLv)
	if vd == nil {
		return 0
	}

	if !gamedata.IsSameDayCommon(g.GuildSignTimesTodayLastTime, nowT) {
		g.GuildSignTimesTodayLastTime = nowT
		g.GuildSignTimesToday = 0
	}

	return vd.GuildSignMaxCount - g.GuildSignTimesToday
}

func (g *Guild) UseGuildSignCount(vipLv int, nowT int64) bool {
	c := g.GetGuildSignCount(vipLv, nowT)
	if c <= 0 {
		return false
	}
	g.GuildSignTimesToday += 1
	return true
}

func (g *Guild) SyncGuildInfo(acid string, guuid, gname string, guildPosition int, leaveTime, nextJoinTime int64, tag *clienttag.ClientTag,
	lootID []string, times []int64) {
	if guuid == "" {
		g.ClearGuild()
		if leaveTime != 0 {
			g.LastLeaveGuildTime = leaveTime
		}
		if nextJoinTime != 0 {
			g.NextEnterGuildTime = nextJoinTime
		}
		if lootID != nil && times != nil {
			g.GuildAssignInfo = GuildAssignInfo{
				AssignID:    lootID,
				AssignTimes: times,
			}
		}
		return
	}

	if g.GuildUUID != guuid {
		tag.SetTag(clienttag.Tag_GuildIn, 1)
	}

	g.GuildUUID = guuid
	g.GuildName = gname
	g.GuildPosition = guildPosition

	if !gamedata.CheckApprovePosition(guildPosition) {
		g.HasApplyCanApprove = false
	}
}

func (g *Guild) InGuild() bool {
	return g.GuildUUID != ""
}

// 获取当前官职
func (g *Guild) GetPost() (string, int64) {
	return g.Post, g.PostOverTime
}

func (g *Guild) SetPost(post string, overTime int64) {
	logs.Trace("SetPost %s %v", post, overTime)
	g.Post = post
	g.PostOverTime = overTime
}

func (g *Guild) ClearGuild() {
	g.GuildUUID = ""
	g.GuildName = ""
	g.GuildPosition = 0
	g.HasApplyCanApprove = false
}

func (g *Guild) IsLimitedByRejoinGuild(nowSt int64) bool {
	return gamedata.IsSameDayCommon(g.LastLeaveGuildTime, nowSt)
}

func (g *Guild) UpdateName(newName string) {
	g.GuildName = newName
}

func (g *Guild) IsTodayChangeGuild(now int64) bool {
	if !g.InGuild() {
		return false
	}
	return g.IsLimitedByRejoinGuild(now)
}
