package ws_pvp

const (
	WS_PVP_DB = "wspvp"
	TOP_N     = 100
)

const (
	WSPVP_PLAYER_GET_TOPN = iota
	WSPVP_UPDATE_TOPN
	WSPVP_PLAYER_GET_BEST9_TOPN // 最强9人战力排行榜
	WSPVP_UPDATE_BEST9_TOPN
)

const WS_PVP_RANK_MAX = 10000
const WS_PVP_ROBOT_ID_PREFIX = "wspvp"

const Wspvp_Time_Interval = 60

const Wspvp_Redis_Pool_Size = 5
