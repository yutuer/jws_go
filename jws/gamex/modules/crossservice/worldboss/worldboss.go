package worldboss

import (
	"fmt"

	cs_worldboss "vcs.taiyouxi.net/jws/crossservice/module/worldboss"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice"
)

func init() {
	crossservice.RegGroupHandle(GetGroupIDbyShardID)
}

//BossStatus ..
type BossStatus cs_worldboss.BossStatus

//DamageRankElem ..
type DamageRankElem cs_worldboss.DamageRankElemInfo

//FormationRankElem ..
type FormationRankElem cs_worldboss.FormationRankElemInfo

//Status ..
type Status struct {
	TotalDamage uint64
	Boss        BossStatus
	Rank        []DamageRankElem
	MyPos       uint32
	MyDamage    uint64
	DamageRound uint64
}

//DamageRank ..
type DamageRank struct {
	Top    []DamageRankElem
	MyRank DamageRankElem
}

//FormationRank ..
type FormationRank struct {
	Top    []FormationRankElem
	MyRank FormationRankElem
}

//PlayerInfo ..
type PlayerInfo cs_worldboss.PlayerInfo

//AttackInfo ..
type AttackInfo cs_worldboss.AttackInfo

//TeamInfoDetail ..
type TeamInfoDetail struct {
	DamageInLife uint64
	EquipAttr    []int64
	DestinyAttr  []int64
	JadeAttr     []int64
	Team         []HeroInfoDetail
	BuffLevel    uint32
}

//HeroInfoDetail ..
type HeroInfoDetail struct {
	Idx       int   `json:"idx"`        // id
	StarLevel int   `json:"star_level"` // 星级
	Level     int   `json:"level"`
	BaseGs    int64 `json:"base_gs"`
	ExtraGs   int64 `json:"extra_gs"`
}

//PlayerDetailInfo ..
type PlayerDetailInfo struct {
	PlayerInfo     PlayerInfo
	TeamInfoDetail TeamInfoDetail
}

//GetInfo ..
func GetInfo(sid uint, acid string) (*Status, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	param := &cs_worldboss.ParamGetInfo{
		Sid:  uint32(sid),
		Acid: acid,
	}
	source := acid
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, cs_worldboss.ModuleID, cs_worldboss.MethodGetInfoID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss GetInfo CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("WorldBoss GetInfo CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*cs_worldboss.RetGetInfo)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss GetInfo CrossService CallSync Return Un-match")
	}

	status := &Status{
		Boss:     BossStatus(ret.Boss),
		MyPos:    ret.MyPos,
		MyDamage: ret.MyDamage,
		Rank:     []DamageRankElem{},
	}

	return status, crossservice.ErrOK, nil
}

//Join ..
func Join(sid uint, acid string, player *PlayerInfo) (*Status, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	param := &cs_worldboss.ParamJoin{
		Sid:    uint32(sid),
		Acid:   acid,
		Player: cs_worldboss.PlayerInfo(*player),
	}
	source := acid
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, cs_worldboss.ModuleID, cs_worldboss.MethodJoinID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss Join CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("WorldBoss Join CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*cs_worldboss.RetJoin)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss Join CrossService CallSync Return Un-match")
	}

	rankList := make([]DamageRankElem, 0, len(ret.Rank))
	for _, r := range ret.Rank {
		rankList = append(rankList, DamageRankElem(r))
	}

	status := &Status{
		Boss:        BossStatus(ret.Boss),
		MyPos:       ret.MyPos,
		MyDamage:    ret.MyDamage,
		TotalDamage: ret.TotalDamage,
		Rank:        rankList,
	}

	return status, crossservice.ErrOK, nil
}

//Attack ..
func Attack(sid uint, acid string, attack *AttackInfo) (*Status, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	param := &cs_worldboss.ParamAttack{
		Sid:    uint32(sid),
		Acid:   acid,
		Attack: cs_worldboss.AttackInfo(*attack),
	}
	source := acid
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, cs_worldboss.ModuleID, cs_worldboss.MethodAttackID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss Attack CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("WorldBoss Attack CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*cs_worldboss.RetAttack)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss Attack CrossService CallSync Return Un-match")
	}

	rankList := make([]DamageRankElem, 0, len(ret.Rank))
	for _, r := range ret.Rank {
		rankList = append(rankList, DamageRankElem(r))
	}

	status := &Status{
		Boss:        BossStatus(ret.Boss),
		MyPos:       ret.MyPos,
		MyDamage:    ret.MyDamage,
		DamageRound: ret.DamageRound,
		TotalDamage: ret.TotalDamage,
		Rank:        rankList,
	}

	return status, crossservice.ErrOK, nil
}

//Leave ..
func Leave(sid uint, acid string, team *TeamInfoDetail, badCheat bool) (*Status, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	heroList := []cs_worldboss.HeroInfoDetail{}
	for _, t := range team.Team {
		hero := cs_worldboss.HeroInfoDetail{
			Idx:       t.Idx,
			StarLevel: t.StarLevel,
			Level:     t.Level,
			BaseGs:    t.BaseGs,
			ExtraGs:   t.ExtraGs,
		}
		heroList = append(heroList, hero)
	}
	param := &cs_worldboss.ParamLeave{
		Sid:  uint32(sid),
		Acid: acid,
		Team: cs_worldboss.TeamInfoDetail{
			EquipAttr:   team.EquipAttr,
			DestinyAttr: team.DestinyAttr,
			JadeAttr:    team.JadeAttr,
			Team:        heroList,
			BuffLevel:   team.BuffLevel,
		},
		BadCheat: badCheat,
	}
	source := acid
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, cs_worldboss.ModuleID, cs_worldboss.MethodLeaveID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss Leave CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("WorldBoss Leave CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*cs_worldboss.RetLeave)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss Leave CrossService CallSync Return Un-match")
	}

	status := &Status{
		Boss:     BossStatus(ret.Boss),
		MyPos:    ret.MyPos,
		MyDamage: ret.MyDamage,
		Rank:     []DamageRankElem{},
	}

	return status, crossservice.ErrOK, nil
}

//GetRank ..
func GetRank(sid uint, acid string) (*DamageRank, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	param := &cs_worldboss.ParamGetRank{
		Sid:  uint32(sid),
		Acid: acid,
	}
	source := acid
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, cs_worldboss.ModuleID, cs_worldboss.MethodGetRankID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss GetRank CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("WorldBoss GetRank CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*cs_worldboss.RetGetRank)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss GetRank CrossService CallSync Return Un-match")
	}

	top := make([]DamageRankElem, 0, len(ret.Rank))
	for _, r := range ret.Rank {
		top = append(top, DamageRankElem(r))
	}

	rank := &DamageRank{
		MyRank: DamageRankElem(ret.MyRank),
		Top:    top,
	}

	return rank, crossservice.ErrOK, nil
}

//GetFormationRank ..
func GetFormationRank(sid uint, acid string) (*FormationRank, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	param := &cs_worldboss.ParamGetFormationRank{
		Sid:  uint32(sid),
		Acid: acid,
	}
	source := acid
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, cs_worldboss.ModuleID, cs_worldboss.MethodGetFormationRankID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss GetFormationRank CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("WorldBoss GetFormationRank CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*cs_worldboss.RetGetFormationRank)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss GetFormationRank CrossService CallSync Return Un-match")
	}

	top := make([]FormationRankElem, 0, len(ret.Rank))
	for _, r := range ret.Rank {
		top = append(top, FormationRankElem(r))
	}

	rank := &FormationRank{
		MyRank: FormationRankElem(ret.MyRank),
		Top:    top,
	}

	return rank, crossservice.ErrOK, nil
}

//PlayerDetail ..
func PlayerDetail(sid uint, acid string, checkAcid string) (*PlayerDetailInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	param := &cs_worldboss.ParamPlayerDetail{
		Sid:       uint32(sid),
		Acid:      acid,
		CheckAcid: checkAcid,
	}
	source := acid
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, cs_worldboss.ModuleID, cs_worldboss.MethodPlayerDetailID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss PlayerDetail CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("WorldBoss PlayerDetail CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*cs_worldboss.RetPlayerDetail)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("WorldBoss PlayerDetail CrossService CallSync Return Un-match")
	}

	info := &PlayerDetailInfo{
		PlayerInfo: PlayerInfo(ret.PlayerInfo),
		TeamInfoDetail: TeamInfoDetail{
			EquipAttr:   ret.Team.EquipAttr,
			DestinyAttr: ret.Team.DestinyAttr,
			JadeAttr:    ret.Team.JadeAttr,
		},
	}
	info.TeamInfoDetail.Team = make([]HeroInfoDetail, len(ret.Team.Team))
	for i, h := range ret.Team.Team {
		info.TeamInfoDetail.Team[i] = HeroInfoDetail(h)
	}

	return info, crossservice.ErrOK, nil
}

//GetGroupIDbyShardID ..
func GetGroupIDbyShardID(sid uint) uint32 {
	return gamedata.GetWBGroupId(uint32(sid))
}
