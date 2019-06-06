package gvg

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/errorcode"
)

const GVG_AVATAR_COUNT = helper.GVG_AVATAR_COUNT
const GVG_SCORE_BASE = 10000000000
const GVG_CITY_COUNT = 15

// 长安城,写死。。。
const CHANGAN_ID = 1

const WORLD_PLAYER_INFO_COUNT = 100

const BILOG_CITY_GUILDINFO_COUNT = 3

const MATCH_POLL_TIME = 8

const (
	_ = iota
	Cmd_Typ_EnterCity
	Cmd_Typ_LeaveCity
	Cmd_Typ_PrepareFight
	Cmd_Typ_CancelMatch
	Cmd_Typ_EndFight
	Cmd_Typ_MutiplayEnd

	Cmd_Typ_Get_SelfGuildInfo
	Cmd_Typ_Get_GuildRank
	Cmd_Typ_Get_PlayerInfo
	Cmd_Typ_Get_SelfGuildAllInfo
	Cmd_Typ_Get_GuildWorldRank
	Cmd_Typ_Get_PlayerWorldInfo

	Cmd_Typ_Get_CityLeader

	Cmd_Typ_Remove_Player
	Cmd_Typ_Remove_Guild

	Cmd_Typ_Rename_Guild
	Cmd_Type_Rename_Player
)

var ( // TODO by zz 都改为warn
	err_city = errorcode.New("City Arg Error", 21)
)

const (
	_ = iota
	player_state_idle
	player_state_prepare
	player_state_fight
)

const gvg_stop_url = "/gamex/gvg/stop"
