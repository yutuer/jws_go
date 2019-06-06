package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/team_pvp"
	"vcs.taiyouxi.net/platform/planx/util"
)

type PlayerTeamPvp struct {
	Rank             int                `json:"r" codec:"r"`
	FightAvatars     []int              `json:"fas" codec:"fas"`
	NextEnemyRefTime int64              `json:"nt" codec:"nt"`
	Enemies          []team_pvp.TPEnemy `json:"es" codec:"es"`
	// 排名被动变化标记
	RankChgPassive bool     `json:"rchgp" codec:"rchgp"`
	FightEnemyID   string   `json:"f_e"`
	PvpCountToday  int      `json:"f_c_t"`
	OpenedChestIDs []uint32 `json:"open_c_id"`
	ResetChestTime int64    `json:"rs_c_t`
}

func (tpvp *PlayerTeamPvp) OnAfterLogin() {
	if tpvp.FightAvatars == nil {
		tpvp.FightAvatars = []int{0, 1, 2}
	}
}

func (tpvp *PlayerTeamPvp) SyncRank(p *Account) {
	// 取自己的排名
	ret := team_pvp.GetModule(p.AccountID.ShardId).CommandExec(team_pvp.TeamPvpCmd{
		Typ:  team_pvp.TeamPvp_Cmd_MyRank,
		Acid: p.AccountID.String(),
	})
	tpvp.SetRank(ret.MyNewRank, false)
	p.Profile.GetFirstPassRank().OnRank(gamedata.FirstPassRankTypTeamPvp, ret.MyNewRank)
}

func (tpvp *PlayerTeamPvp) SetRank(rank int, beFight bool) {
	tpvp.Rank = rank
	if beFight {
		tpvp.RankChgPassive = true
	}
}

func (p *PlayerTeamPvp) CanOpenChest(id uint32) bool {
	for _, chestID := range p.OpenedChestIDs {
		if chestID == id {
			return false
		}
	}
	return true
}

func (p *PlayerTeamPvp) SetChestOpen(id uint32) {
	p.OpenedChestIDs = append(p.OpenedChestIDs, id)
}

func (p *PlayerTeamPvp) UpdateChestInfo(now_t int64) {
	if now_t < p.ResetChestTime {
		return
	}
	p.ResetChestTime = util.DailyBeginUnixByStartTime(now_t,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	p.ResetChestTime += util.DaySec
	p.OpenedChestIDs = make([]uint32, 0, 10)
	p.PvpCountToday = 0

}
