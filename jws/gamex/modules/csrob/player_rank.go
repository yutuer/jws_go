package csrob

import (
	"time"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type rankPlayer struct {
	// close chan struct{}
	// sync.WaitGroup

	res *resources
}

func newRankPlayer(res *resources) *rankPlayer {
	return &rankPlayer{
		// close: make(chan struct{}, 1),
		res: res,
	}
}

func (r *rankPlayer) Add(acid string, nat uint32, team []HeroInfoForRank) {
	if 0 == len(team) {
		return
	}

	rt := &RankTeam{
		Heros: team,
		Acid:  acid,
	}

	if err := r.res.PlayerRankDB.pushFormationAndRank(nat, acid, rt, time.Now().Unix()); nil != err {
		logs.Error(fmt.Sprintf("[CSRob] rankPlayer Add, pushFormationAndRank failed, %v", err))
		return
	}

	return
}

func (r *rankPlayer) GetRank(nat uint32) []*RankTeam {
	list, err := r.res.PlayerRankDB.rangeFormationByRank(nat, 100)
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] rankPlayer GetRank, rangeFormationByRank failed, %v", err))
		return []*RankTeam{}
	}

	for id, team := range list {
		list[id].Name = r.res.poolName.GetPlayerCSName(team.Acid)
	}

	return list
}

func (r *rankPlayer) GetPos(nat uint32, acid string) (*RankTeam, uint32) {
	team, pos, err := r.res.PlayerRankDB.getFormationAndPos(nat, acid)
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] rankPlayer GetPos, getFormationAndPos failed, %v", err))
		return nil, 0
	}

	if nil == team {
		team = &RankTeam{
			Acid:  acid,
			Heros: []HeroInfoForRank{},
		}
	}

	team.Name = r.res.poolName.GetPlayerCSName(acid)

	return team, pos
}

func (r *rankPlayer) RemoveRank(nat uint32, acid string) {
	err := r.res.PlayerRankDB.removeFromFormationAndRank(nat, acid)
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] rankPlayer RemoveRank, removeFromFormationAndRank failed, %v", err))
	}

	return
}

func (a *rankPlayer) RemovePlayerRank(acid string) {
	natList := gamedata.CSRobNatList()
	for _, nat := range natList {
		a.RemoveRank(nat, acid)
	}
}

// func (r *rankPlayer) Start() {
// 	go func() {
// 		r.Add(1)
// 		r.doRank()
// 		r.Done()
// 	}()
// 	logs.Info("[CSRob] rankPlayer Start")
// }

// func (r *rankPlayer) Stop() {
// 	r.close <- struct{}{}
// 	r.Wait()
// 	close(r.close)
// 	logs.Info("[CSRob] rankPlayer Stop")
// }

// func (r *rankPlayer) doRank() {
// 	after := time.After(delayRankDo)
// 	weekEnd := time.After(1)
// 	bClose := false
// 	for !bClose {
// 		select {
// 		case <-r.close:
// 			bClose = true
// 		case trig := <-r.trigQueue:
// 			logs.Debug("[CSRob] receive trig")
// 			r.preMap[trig] = r.preMap[trig] + 1
// 		case <-after:
// 			r.refreshTriggers()
// 			after = time.After(delayRankDo)
// 		case <-weekEnd:
// 			logs.Info("[CSRob] doRank weekEnd")
// 			weekEnd = time.After(r.getNextRankRewardTimeAfter())
// 			r.res.CommandMod.notifyRewardGuildWeek(r.getLastRankRewardTime())
// 		case <-r.clearSignal:
// 			r.clearRankData()
// 		}
// 	}

// 	logs.Info("[CSRob] rankPlayer doRank close")
// }
