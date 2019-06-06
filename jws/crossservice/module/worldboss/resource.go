package worldboss

type resources struct {
	group  uint32
	module *WorldBoss

	BossMod          *BossMod
	RankDamageMod    *RankDamageMod
	FormationRankMod *FormationRankMod
	PlayerMod        *PlayerMod
	ticker           *tickerHolder

	BossDB   *BossDB
	RankDB   *RankDB
	PlayerDB *PlayerDB

	callback *callbackHolder
}

func newResources(group uint32, m *WorldBoss) *resources {
	res := &resources{
		group:  group,
		module: m,
	}

	res.ticker = newTickerHolder(res)

	res.BossMod = newBossMod(res)
	res.RankDamageMod = newRankDamageMod(res)
	res.FormationRankMod = newFormationRankMod(res)
	res.PlayerMod = newPlayerMod(res)

	res.BossDB = newBossDB(res)
	res.RankDB = newRankDB(res)
	res.PlayerDB = newPlayerDB(res)

	res.callback = newCallbackHolder(res)

	return res
}

type callbackHolder struct {
	res *resources
}

func newCallbackHolder(res *resources) *callbackHolder {
	return &callbackHolder{
		res: res,
	}
}
