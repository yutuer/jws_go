package notify

const (
	RedPointTyp_CondActGift = iota
	RedPointTyp_Guild
	RedPointTyp_Quest
	RedPointTyp_ActGift
	RedPointTyp_Gank
	RedPointTyp_Count
)

type NotifySyncMsg struct {
	notifyAddress          string
	SyncPlayerGuild        bool  `codec:"guild_p_n_"`
	SyncPlayerGuildApply   bool  `codec:"guild_p_a_n_"`
	SyncGuildInfoNeed      bool  `codec:"guild_i_n_"`
	SyncGuildMemsNeed      bool  `codec:"guild_mm_n_"`
	SyncApplyGuildMemsNeed bool  `codec:"guild_a_mm_n_"`
	SyncGuildScienceNeed   bool  `codec:"guild_sc_n_"`
	SyncGuildInventory     bool  `codec:"guild_invty_"`
	SyncClientTagNeed      bool  `codec:"tag_n_"`
	SyncRedPoint           []int `codec:"redp_"`
	SyncTeamPvpNeed        bool  `codec:"tpvp_b_"`
	SyncTitleNeed          bool  `codec:"title_n_"`
	SyncGuildRedPacketNeed bool  `codec:"grp_need"`
	SyncGuildWorshipNeed   bool  `codec:"gwship_n_"`
	SyncPlayerHc           bool  `codec:"player_hc_n_"`

	SyncRoom      []byte `codec:"room_"`
	SyncRoomEvent []byte `codec:"roomev_"`

	needSyncGuildActBoss            bool
	gates_enemy_data_need_sync      bool
	gates_enemy_push_data_need_sync bool
}

func (s *NotifySyncMsg) SetAddr(addr string) {
	s.notifyAddress = addr
}

func (s *NotifySyncMsg) GetAddr() string {
	return s.notifyAddress
}

func (s *NotifySyncMsg) OnChangeRedPoint(redPointTyp int) {
	if s.SyncRedPoint == nil {
		s.SyncRedPoint = make([]int, RedPointTyp_Count)
	}
	s.SyncRedPoint[redPointTyp] = 1
}

func (s *NotifySyncMsg) OnChangePlayerHc() {
	s.SyncPlayerHc = true
}

func (s *NotifySyncMsg) OnChangeGatesEnemyData() {
	s.gates_enemy_data_need_sync = true
	s.gates_enemy_push_data_need_sync = true
}

func (s *NotifySyncMsg) OnChangeGatesEnemyPushData() {
	s.gates_enemy_push_data_need_sync = true
}

func (s *NotifySyncMsg) OnChangePlayerGuild() {
	s.SyncPlayerGuild = true
}

func (s *NotifySyncMsg) OnChangePlayerGuildApply() {
	s.SyncPlayerGuildApply = true
}

func (s *NotifySyncMsg) OnChangeGuildInfo() {
	s.SyncGuildInfoNeed = true
}

func (s *NotifySyncMsg) OnChangeGuildMemsInfo() {
	s.SyncGuildMemsNeed = true
}

func (s *NotifySyncMsg) OnChangeGuildApplyMemsInfo() {
	s.SyncApplyGuildMemsNeed = true
}

func (s *NotifySyncMsg) OnChangeGuildScience() {
	s.SyncGuildScienceNeed = true
}

func (s *NotifySyncMsg) OnChangeGuildInventory() {
	s.SyncGuildInventory = true
}

func (s *NotifySyncMsg) OnChangeGuildRedPacket() {
	s.SyncGuildRedPacketNeed = true
}

func (s *NotifySyncMsg) OnChangeClientTagInfo() {
	s.SyncClientTagNeed = true
}

func (s *NotifySyncMsg) IsChangeGatesEnemyData() bool {
	return s.gates_enemy_data_need_sync
}

func (s *NotifySyncMsg) IsChangeGatesEnemyPushData() bool {
	return s.gates_enemy_push_data_need_sync
}

func (s *NotifySyncMsg) OnChangeTeamPvp() {
	s.SyncTeamPvpNeed = true
}

func (s *NotifySyncMsg) OnChangeTitle() {
	s.SyncTitleNeed = true
}

func (s *NotifySyncMsg) SetNeedSyncGuildActBoss() {
	s.needSyncGuildActBoss = true
}

func (s *NotifySyncMsg) IsNeedSyncGuildActBoss() bool {
	return s.needSyncGuildActBoss
}

func (s *NotifySyncMsg) OnChangeGuildWorshipInfo() {
	s.SyncGuildWorshipNeed = true
}
