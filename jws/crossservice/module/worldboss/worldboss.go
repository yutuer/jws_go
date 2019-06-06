package worldboss

import (
	"fmt"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//..
const (
	ModuleID = "worldboss"

	MethodGetInfoID          = "getinfo"
	MethodJoinID             = "join"
	MethodAttackID           = "attack"
	MethodLeaveID            = "leave"
	MethodGetRankID          = "getrank"
	MethodGetFormationRankID = "formationrank"
	MethodPlayerDetailID     = "playerdetail"

	CallbackDamageRankID = "damagerank"
	CallbackMarqueeID    = "marquee"
)

func init() {
	module.RegModule(&Generator{})
}

//Generator ..
type Generator struct {
}

//ModuleID ..
func (g *Generator) ModuleID() string {
	return ModuleID
}

//NewModule ..
func (g *Generator) NewModule(group uint32) module.Module {
	moduleWorldBoss := &WorldBoss{
		BaseModule: module.BaseModule{
			GroupID: group,
			Module:  ModuleID,
			Methods: map[string]module.Method{},
			Static:  false,
		},
	}
	moduleWorldBoss.Methods[MethodGetInfoID] = newMethodGetInfo(moduleWorldBoss)
	moduleWorldBoss.Methods[MethodJoinID] = newMethodJoin(moduleWorldBoss)
	moduleWorldBoss.Methods[MethodAttackID] = newMethodAttack(moduleWorldBoss)
	moduleWorldBoss.Methods[MethodLeaveID] = newMethodLeave(moduleWorldBoss)
	moduleWorldBoss.Methods[MethodGetRankID] = newMethodGetRank(moduleWorldBoss)
	moduleWorldBoss.Methods[MethodGetFormationRankID] = newMethodGetFormationRank(moduleWorldBoss)
	moduleWorldBoss.Methods[MethodPlayerDetailID] = newMethodPlayerDetail(moduleWorldBoss)

	moduleWorldBoss.Methods[CallbackDamageRankID] = newCallbackDamageRank(moduleWorldBoss)
	moduleWorldBoss.Methods[CallbackMarqueeID] = newCallbackMarquee(moduleWorldBoss)

	logs.Info("[WorldBoss] NewModule for Group [%d]", group)

	return moduleWorldBoss
}

//WorldBoss ..
type WorldBoss struct {
	module.BaseModule

	res *resources
}

//HashMask ..
func (s *WorldBoss) HashMask() uint32 {
	return 32
}

//Start ..
func (s *WorldBoss) Start() {
	s.res = newResources(s.GetGroupID(), s)
	logs.Info("[WorldBoss] Group [%d] Start", s.GroupID)
}

//AfterStart ..
func (s *WorldBoss) AfterStart() {
	logs.Info("[WorldBoss] Group [%d] AfterStart", s.GroupID)

	if err := s.res.ticker.loadRoundFromDB(); nil != err {
		logs.Error(fmt.Sprintf("[WorldBoss] Group [%d] AfterStart, loadRoundFromDB failed, %v", s.GroupID, err))
	}
	if err := s.res.BossMod.loadBossFromDB(); nil != err {
		logs.Error(fmt.Sprintf("[WorldBoss] Group [%d] AfterStart, loadBossFromDB failed, %v", s.GroupID, err))
	}
	if err := s.res.RankDamageMod.loadRankFromDB(); nil != err {
		logs.Error(fmt.Sprintf("[WorldBoss] Group [%d] AfterStart, loadRankFromDB failed, %v", s.GroupID, err))
	}
	if err := s.res.FormationRankMod.loadRankFromDB(); nil != err {
		logs.Error(fmt.Sprintf("[WorldBoss] Group [%d] AfterStart, loadRankFromDB failed, %v", s.GroupID, err))
	}

	s.res.ticker.start()
}

//BeforeStop ..
func (s *WorldBoss) BeforeStop() {
	logs.Info("[WorldBoss] Group [%d] BeforeStop, begin...", s.GroupID)
	s.res.RankDamageMod.saveAllRank()
	s.res.FormationRankMod.saveAllRank()
	s.res.BossMod.saveBossToDB()
	s.res.ticker.saveRoundToDB()
	logs.Info("[WorldBoss] Group [%d] BeforeStop, end...", s.GroupID)
}

//Stop ..
func (s *WorldBoss) Stop() {
	s.res.ticker.stop()
	logs.Info("[WorldBoss] Group [%d] Stop", s.GroupID)
}
