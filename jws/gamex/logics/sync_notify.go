package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logics/notify"
	"vcs.taiyouxi.net/jws/gamex/logics/sync"
)

type SyncRespNotify struct {
	Resp
	notify.NotifySyncMsg

	syncData.SyncGuildActBoss

	// guild
	SyncPlayerGuildInfo []byte `codec:"guild_p_i_"`

	SyncPlayerApplyGuild_UUID   []string `codec:"guild_p_a_uid"`
	SyncPlayerApplyGuild_Name   []string `codec:"guild_p_a_nm"`
	SyncPlayerApplyGuild_Lvl    []uint32 `codec:"guild_p_a_lvl"`
	SyncPlayerApplyGuild_Notice []string `codec:"guild_p_a_ntc"`
	SyncPlayerApplyGuild_Time   []int64  `codec:"guild_p_a_t"`

	SyncGuildInfo []byte `codec:"guild_i_"`
	SyncPost      string `codec:"post_"`

	SyncGuildMems [][]byte `codec:"guild_mm_"`

	SyncGuildApplyList [][]byte `codec:"guild_a_mm_"`

	SyncClientTag    []int  `codec:"tag_"`
	SyncGuildRanking string `codec:"ranking_"`

	// guild inventory
	SyncGuildInventoryNextRefTime    int64    `codec:"guild_invy_nrt_"`
	SyncGuildInventoryNextResetTime  int64    `codec:"guild_invy_nrstt_"`
	SyncGuildInventoryBossCoin       int64    `codec:"guild_invy_bc_"`
	SyncGuildInventoryLoots          [][]byte `codec:"guild_invy_lts_"` // []GuildInventoryLoot
	SyncGuildInventoryLootMemCount   []int    `codec:"guild_invy_ltms_c_"`
	SyncGuildInventoryLootMemAssign  [][]byte `codec:"guild_invy_ltms_assign"`
	SyncGuildInventorySelfApplyCount []int    `codec:"guild_invy_slf_c_"`
	SyncGuildInventoryPreLoots       [][]byte `codec:"guild_invy_prelts_"` // []GuildInventoryLoot

	SyncGuildLostInventoryLoots          [][]byte `codec:"guild_lost_invy_lts_"` // []GuildInventoryLoot
	SyncGuildLostInventoryLootMemCount   []int    `codec:"guild_lost_invy_ltms_c_"`
	SyncGuildLostInventoryLootMemAssign  [][]byte `codec:"guild_lost_invy_ltms_assign"`
	SyncGuildLostInventorySelfApplyCount []int    `codec:"guild_lost_invy_slf_c_"`
	SyncGuildLostInventoryActive         bool     `codec:"guild_lost_invy_act"`

	SyncGuildInventoryInfo []byte `codec:"guild_"`

	// guild science
	SyncGuildScience [][]byte `codec:"guild_sc_"`

	// GatesEnemy
	// 需要推送的
	SyncGatesEnemyEnemyInfo     [][]byte `codec:"gees_"`
	SyncGatesEnemyState         int      `codec:"gestat_"`
	SyncGatesEnemyStateOverTime int64    `codec:"getime_"`
	SyncGatesEnemyKillPoint     int      `codec:"gekp_"`
	SyncGatesEnemyBossMax       int      `codec:"gebm_"`
	SyncGatesEnemyBuffMemName   []string `codec:"gebfnm_"`
	SyncGatesEnemyBuffCurLv     uint32   `codec:"gebflv_"`

	// 不需推送跟玩家请求走
	SyncGatesEnemyBossState          []byte   `codec:"gebstat_"`
	SyncGatesEnemyRankNames          []string `codec:"gernames_"`
	SyncGatesEnemyRankPoints         []int    `codec:"gerpoints_"`
	SyncGatesEnemyRankFashion        []string `codec:"gerfashion_"`
	SyncGatesEnemyRankWeaponStartLvl []uint32 `codec:"gerwstar_"`
	SyncGatesEnemyRankEqStartLvl     []uint32 `codec:"gerestar_"`
	SyncGatesEnemyRankStates         []int    `codec:"gerstats_"`
	SyncGatesEnemyRankAvatarIDs      []int    `codec:"geraIDs_"`
	SyncGatesEnemyRankTitleOn        []string `codec:"gertitle_"`
	SyncGatesEnemyPointAll           int      `codec:"gepall_"`
	SyncGatesEnemyNeed               bool     `codec:"geneed_"`
	SyncGatesEnemySwing              []int    `codec:"geswing"`
	SyncGatesEnemyMagicPet           []uint32 `codec:"geMagicpet"`

	// commoninfo 生效就带的
	SyncGateEnemyStartTime int64 `codec:"ge_s_t_"`
	SyncGateEnemyEndTime   int64 `codec:"ge_e_t_"`

	// team pvp
	SyncTeamPvpRank         int      `codec:"tpvp_r_"`
	SyncTeamPvpAvatars      []int    `codec:"tpvp_as_"`
	SyncTeamPvpCountToday   int      `codec:"tpvp_c_t"`
	SyncTeamPvpOpenedChests []uint32 `codec:"tpvp_chest"`

	SyncFirstPassReward          bool  `codec:"fpr_is_"`
	SyncSimplePvpFirstPassReward []int `codec:"spvp_fpr_"`
	SyncTeamPvFirstPassReward    []int `codec:"tpvp_fpr_"`
	SyncSimplePvpMaxRanks        int   `codec:"spvp_max_rank_"`
	SyncTeamPvpMaxRank           int   `codec:"tpvp_maxr_"`

	// 称号
	SyncTitleCanActivate []string `codec:"title_act"`
	SyncTitles           []string `codec:"title_"`
	SyncTitleOn          string   `codec:"title_on_"`
	SyncTitleNextRefTime int64    `codec:"title_nx_t_"`
	SyncTitleHint        []string `codec:"title_ht_"`

	// 军团红包
	SyncGuildRedPacket     []byte `codec:"grp_info"`
	SyncRedPacketIpaStatus int    `codec:"grp_status"`

	// 军团膜拜
	PersionTakeNum int64    `codec:"per_t_n"`
	GuildTakeNum   int64    `codec:"gui_t_n"`
	TakeId         int64    `codec:"t_id"`
	TakeSign       int64    `codec:"t_sign"`
	Reward         []string `codec:"guild_r"`
	OneTime        []int64  `codec:"guild_n_1"`
	DoubleTime     []int64  `codec:"guild_n_2"`
	HasReward      []int64  `codec:"has_reward"`
	IsOpen         bool     `codec:"is_o"`
}

func (s *SyncRespNotify) FromSyncMsg(msg notify.NotifySyncMsg) {
	s.NotifySyncMsg = msg
}

func (s *SyncRespNotify) MkNotifyInfo(p *Account) {
	// guild
	s.mkGuildInfo(p)
	// team pvp
	s.mkTeamPvpInfo(p)
	s.mkFirstPassRewardInfo(p)
	// title
	s.mkTitleInfo(p)
}
