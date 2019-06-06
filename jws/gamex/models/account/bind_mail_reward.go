package account

type BindMailRewardInfo struct {
	BindMailRewardGet bool `redis:"bind_m_rw_g"`
	BindEGRewardGet   bool `redis:"bind_eg_rw_g"`
}
