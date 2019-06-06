package worldboss

import (
	"sync"
	"time"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ..
const (
	defaultName = "N/A"
)

//PlayerInfo ..
type PlayerInfo struct {
	Acid      string
	Sid       uint32
	Name      string
	Vip       uint32
	Level     uint32
	Gs        int64
	GuildName string

	damageInLife uint64
}

//HeroInfoDetail ..
type HeroInfoDetail struct {
	Idx       int   `json:"idx"`        // id
	StarLevel int   `json:"star_level"` // 星级
	Level     int   `json:"level"`
	BaseGs    int64 `json:"base_gs"`
	ExtraGs   int64 `json:"extra_gs"`
}

//TeamInfoDetail ..
type TeamInfoDetail struct {
	DamageInLife uint64
	EquipAttr    []int64
	DestinyAttr  []int64
	JadeAttr     []int64
	Team         []HeroInfoDetail
	BuffLevel    uint32
}

//PlayerMod ..
type PlayerMod struct {
	res *resources

	lockPlayer sync.RWMutex
	players    map[string]*PlayerInfo

	lockTeams sync.RWMutex
	teams     map[string]*TeamInfoDetail
}

func newPlayerMod(res *resources) *PlayerMod {
	pm := &PlayerMod{}
	pm.res = res
	pm.players = make(map[string]*PlayerInfo)
	pm.teams = make(map[string]*TeamInfoDetail)
	return pm
}

func (pm *PlayerMod) updatePlayerInfo(info *PlayerInfo) {
	pm.updatePlayerInfoToCache(info)
	pm.updatePlayerInfoToDB(info)
}

func (pm *PlayerMod) updatePlayerInfoToCache(info *PlayerInfo) {
	pm.lockPlayer.Lock()
	defer pm.lockPlayer.Unlock()
	pm.players[info.Acid] = info
}

func (pm *PlayerMod) updatePlayerInfoToDB(info *PlayerInfo) {
	if err := pm.res.PlayerDB.setPlayerInfo(info.Acid, *info, pm.res.ticker.roundStatus.BatchTag); nil != err {
		logs.Error(fmt.Sprintf("[WorldBoss] PlayerMod updatePlayerInfoToDB failed, %v, ..info %+v", err, info))
	}
}

func (pm *PlayerMod) getPlayerInfo(acid string) *PlayerInfo {
	info := pm.getPlayerInfoFromCache(acid)
	if nil != info {
		return info
	}
	info = pm.getPlayerInfoFromDB(acid)
	if nil != info {
		pm.updatePlayerInfoToCache(info)
	}
	return info
}

func (pm *PlayerMod) getPlayerInfoFromCache(acid string) *PlayerInfo {
	pm.lockPlayer.RLock()
	defer pm.lockPlayer.RUnlock()
	return pm.players[acid]
}

func (pm *PlayerMod) getPlayerInfoFromDB(acid string) *PlayerInfo {
	info, err := pm.res.PlayerDB.getPlayerInfo(acid, pm.res.ticker.roundStatus.BatchTag)
	if nil != err {
		logs.Error(fmt.Sprintf("[WorldBoss] PlayerMod getPlayerInfoFromDB failed, %v, ..acid %s", err, acid))
		return nil
	}
	return info
}

func (pm *PlayerMod) updateTeamInfo(acid string, team *TeamInfoDetail) {
	old := pm.getTeamInfoFromCache(acid)
	if nil != old && old.DamageInLife >= team.DamageInLife {
		return
	}
	pm.updateTeamInfoToCache(acid, team)
	pm.updateTeamInfoToDB(acid, team)
}

func (pm *PlayerMod) updateTeamInfoToCache(acid string, team *TeamInfoDetail) {
	nt := team.Copy()
	pm.lockTeams.Lock()
	defer pm.lockTeams.Unlock()
	pm.teams[acid] = nt
}

func (pm *PlayerMod) updateTeamInfoToDB(acid string, team *TeamInfoDetail) {
	if err := pm.res.PlayerDB.setPlayerTeam(acid, *team, pm.res.ticker.roundStatus.BatchTag); nil != err {
		logs.Error(fmt.Sprintf("[WorldBoss] PlayerMod updateTeamInfoToDB failed, %v, ..team %+v", err, team))
	}
}

func (pm *PlayerMod) getTeamInfo(acid string) *TeamInfoDetail {
	team := pm.getTeamInfoFromCache(acid)
	if nil != team {
		return team
	}
	team = pm.getTeamInfoFromDB(acid)
	if nil != team {
		pm.updateTeamInfoToCache(acid, team)
	}
	return team
}

func (pm *PlayerMod) getTeamInfoFromCache(acid string) *TeamInfoDetail {
	pm.lockTeams.RLock()
	defer pm.lockTeams.RUnlock()
	ot := pm.teams[acid]
	if nil == ot {
		return nil
	}
	nt := ot.Copy()
	return nt
}

func (pm *PlayerMod) getTeamInfoFromDB(acid string) *TeamInfoDetail {
	info, err := pm.res.PlayerDB.getPlayerTeam(acid, pm.res.ticker.roundStatus.BatchTag)
	if nil != err {
		logs.Error(fmt.Sprintf("[WorldBoss] PlayerMod getTeamInfoFromDB failed, %v, ..acid %s", err, acid))
		return nil
	}
	return info
}

func (pm *PlayerMod) clearDamageInLife(acid string) {
	info := pm.getPlayerInfoFromCache(acid)
	if nil == info {
		return
	}
	info.damageInLife = 0
}

func (pm *PlayerMod) addDamageInLife(acid string, damage uint64) {
	info := pm.getPlayerInfoFromCache(acid)
	if nil == info {
		return
	}
	info.damageInLife += damage
}

func (pm *PlayerMod) getDamageInLife(acid string) uint64 {
	info := pm.getPlayerInfoFromCache(acid)
	if nil == info {
		return 0
	}
	return info.damageInLife
}

//resetNewRound ..
func (pm *PlayerMod) resetNewRound(now time.Time) {
	logs.Trace("[WorldBoss] RankDamageMod resetNewRound")

	pm.lockPlayer.Lock()
	pm.players = make(map[string]*PlayerInfo)
	pm.lockPlayer.Unlock()

	pm.lockTeams.Lock()
	pm.teams = make(map[string]*TeamInfoDetail)
	pm.lockTeams.Unlock()
}

//Copy ..
func (src *TeamInfoDetail) Copy() *TeamInfoDetail {
	dst := &TeamInfoDetail{
		DamageInLife: src.DamageInLife,
		BuffLevel:    src.BuffLevel,
	}
	dst.DestinyAttr = make([]int64, len(src.DestinyAttr))
	copy(dst.DestinyAttr, src.DestinyAttr)
	dst.EquipAttr = make([]int64, len(src.EquipAttr))
	copy(dst.EquipAttr, src.EquipAttr)
	dst.JadeAttr = make([]int64, len(src.JadeAttr))
	copy(dst.JadeAttr, src.JadeAttr)
	dst.Team = make([]HeroInfoDetail, len(src.Team))
	copy(dst.Team, src.Team)

	return dst
}
