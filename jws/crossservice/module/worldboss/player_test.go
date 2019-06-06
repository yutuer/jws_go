package worldboss

import (
	"reflect"
	"testing"

	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func DebugGetPlayerMod() *PlayerMod {
	return newResources(uint32(0), &WorldBoss{}).PlayerMod
}

func DebugGetPlayerInfo() (string, *PlayerInfo) {
	acid := uuid.NewV4().String()
	playerInfo := &PlayerInfo{
		Acid:         acid,
		Sid:          0,
		Name:         "Test",
		Vip:          1,
		Level:        66,
		Gs:           188888,
		GuildName:    "Testers",
		damageInLife: 666666,
	}
	return acid, playerInfo
}

func DebugGetPlayerModWithPlayerInfo() (string, *PlayerMod) {
	pm := DebugGetPlayerMod()
	acid, playerInfo := DebugGetPlayerInfo()

	pm.updatePlayerInfo(playerInfo)

	return acid, pm
}

func DebugGetTeamInfo() *TeamInfoDetail {
	teamInfo := &TeamInfoDetail{
		DamageInLife: 255,
		EquipAttr:    make([]int64, 8),
		DestinyAttr:  make([]int64, 8),
		JadeAttr:     make([]int64, 8),
		Team:         make([]HeroInfoDetail, 8),
		BuffLevel:    1,
	}
	return teamInfo
}

func TestNewPlayerMod(t *testing.T) {
	pm := newPlayerMod(newResources(uint32(0), &WorldBoss{}))

	if pm.res == nil || pm.players == nil || pm.teams == nil {
		t.Error("newPlayerMod is incorrenct!")
	}
}

func TestUpdatePlayerInfo(t *testing.T) {
	pm := DebugGetPlayerMod()

	acid, playerInfo := DebugGetPlayerInfo()

	pm.updatePlayerInfo(playerInfo)

	if pi, ok := pm.players[acid]; !ok || !reflect.DeepEqual(pi, playerInfo) {
		t.Error("updatePlayerInfo is incorrect!")
	}

	// 暂时略过DB部分
}

func TestGetPlayerInfoFromCache(t *testing.T) {
	pm := DebugGetPlayerMod()
	acid, playerInfo := DebugGetPlayerInfo()

	if pl := pm.getPlayerInfoFromCache(acid); pl != nil {
		t.Error("getPlayerInfoFromCache is incorrect!")
	}

	pm.updatePlayerInfo(playerInfo)

	if pl := pm.getPlayerInfoFromCache(acid); pl == nil || !reflect.DeepEqual(pl, playerInfo) {
		t.Error("getPlayerInfoFromCache is incorrect!")
	}
}

// func TestGetPlayerInfo(t *testing.T) {}

func TestUpdateTeamInfo(t *testing.T) {
	acid, pm := DebugGetPlayerModWithPlayerInfo()

	// teams为nil
	teamInfo := DebugGetTeamInfo()

	pm.updateTeamInfo(acid, teamInfo)

	if pm.teams == nil {
		t.Error("updateTeamInfo is incorrect!")
	}

	// DamageInLife小于上次
	teamInfo2 := teamInfo.Copy()
	teamInfo2.DamageInLife = 128

	pm.updateTeamInfo(acid, teamInfo2)

	if !reflect.DeepEqual(pm.teams[acid], teamInfo) {
		t.Error("updateTeamInfo is incorrect! %d, %d",
			pm.teams[acid].DamageInLife, teamInfo.DamageInLife)
	}

	// DamageInLife大于上次
	teamInfo3 := teamInfo.Copy()
	teamInfo3.DamageInLife = 512

	pm.updateTeamInfo(acid, teamInfo3)

	if !reflect.DeepEqual(pm.teams[acid], teamInfo3) {
		t.Error("updateTeamInfo is incorrect! %d, %d",
			pm.teams[acid].DamageInLife, teamInfo3.DamageInLife)

	}
}

// func TestGetTeamInfo(t *testing.T) {}

func TestGetTeamInfoFromCache(t *testing.T) {
	acid, pm := DebugGetPlayerModWithPlayerInfo()
	teamInfo := DebugGetTeamInfo()

	// 无
	if pm.getTeamInfoFromCache(acid) != nil {
		t.Error("getTeamInfoFromCache is incorrect!")
	}

	pm.updateTeamInfo(acid, teamInfo)

	// 有
	if pm.getTeamInfoFromCache(acid) == nil {
		t.Error("getTeamInfoFromCache is incorrect!")
	}

	// 新
	if newAcid := uuid.NewV4().String(); pm.getTeamInfoFromCache(newAcid) != nil {
		t.Error("getTeamInfoFromCache is incorrect!")
	}
}

func TestClearDamageInLife(t *testing.T) {
	acid, pm := DebugGetPlayerModWithPlayerInfo()

	pm.clearDamageInLife(acid)

	if pm.players[acid].damageInLife != 0 {
		t.Error("clearDamageInLife is incorrect!")
	}

	// 不panic就好
	newAcid := uuid.NewV4().String()
	pm.clearDamageInLife(newAcid)
}

func TestAddDamageInLife(t *testing.T) {
	acid, pm := DebugGetPlayerModWithPlayerInfo()

	pm.addDamageInLife(acid, 333333)

	if pm.players[acid].damageInLife != 999999 {
		t.Error("addDamageInLife is incorrect!")
	}

	// 不panic就好
	newAcid := uuid.NewV4().String()
	pm.addDamageInLife(newAcid, 666666)
}

func TestGetDamageInLife(t *testing.T) {
	acid, pm := DebugGetPlayerModWithPlayerInfo()

	if pm.getDamageInLife(acid) != 666666 {
		t.Error("getDamageInLife is incorrect!")
	}

	// 0
	newAcid := uuid.NewV4().String()
	if pm.getDamageInLife(newAcid) != 0 {
		t.Error("getDamageInLife is incorrect!")
	}
}

func TestCopy(t *testing.T) {
	/*
		注意，任何slice不创建会导致reflect.DeepEqual为false，
		不影响实质数据，只影响测试判断

		teamInfo := &TeamInfoDetail{
			DamageInLife: 255,
		}

		// 会判断为false
		if dstTeamInfo := teamInfo.Copy(); !reflect.DeepEqual(teamInfo, dstTeamInfo) {
			t.Error("TeamInfo copy function is incorrect!")
		}
	*/

	teamInfo := DebugGetTeamInfo()

	if dstTeamInfo := teamInfo.Copy(); !reflect.DeepEqual(teamInfo, dstTeamInfo) {
		t.Error("TeamInfo copy function is incorrect!")
	}
}
