package info

type ActBossData2Client struct {
	BossId          string   `codec:"id_"`
	BossType        int64    `codec:"t_"`
	BossGroup       string   `codec:"group_"`
	BossHp          int64    `codec:"hp_"`
	BossTotalHp     int64    `codec:"thp_"`
	BossState       int64    `codec:"s_"`
	BossIsLock      bool     `codec:"bsl_"`
	BossPlayerState int64    `codec:"bsp_"`
	BossEndTime     int64    `codec:"bsendt_"`
	PlayerID        string   `codec:"bspid_"`
	PlayerName      string   `codec:"bspname_"`
	PlayerAvatarID  int      `codec:"bspa_"`
	RewardCount     int64    `codec:"bsrwc_"`
	RankPlayerIDs   []string `codec:"rank_pids_"`
	RankPlayerNames []string `codec:"rank_pnames_"`
	RankPlayerScore []int64  `codec:"rank_pscores_"`
}

type ActBoss2Client struct {
	BossStats         [][]byte `codec:"guildactbss_"`
	CurrPlayerNum     int64    `codec:"guildactb_pn_"`
	CurrBossLevel     int64    `codec:"guildactb_bl_"`
	DamagePlayerNames []string `codec:"damage_pnames_"`
	DamagePlayerScore []int64  `codec:"damage_pscores_"`
}
