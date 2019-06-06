package battlearmy

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//四个战阵
type BattleArmys struct {
	BattleArmy [helper.BATTLE_ARMY_NUM_MAX]*BattleArmy	`json:"b_a"`
}

func (b *BattleArmys) GetBattleArmys() []*BattleArmy {
	for i:=range b.BattleArmy{
		if b.BattleArmy[i] == nil{
			saveData:=BattleArmy{}
			for j := range saveData.BattleArmyLocs{
				saveData.BattleArmyLocs[j].AvatarID = -1
			}
			b.BattleArmy[i] = &saveData
		}
	}
	return b.BattleArmy[:]
}

//1蜀,2魏,3吴,4群
func (b *BattleArmys) GetBattleArmy(country int) *BattleArmy {
	if country < 1 || country > helper.BATTLE_ARMY_NUM_MAX {
		logs.Error("wrong Country Index,no such country")
	}
	return b.GetBattleArmys()[country-1]
}

type BattleArmy struct {
	BattleArmyLocs [helper.BATTLE_ARMYLOC_NUM_MAX]BattleArmyLoc `json:"b_a_l"`
}

func (b *BattleArmy) GetBattleArmyLocs() []BattleArmyLoc {
	return b.BattleArmyLocs[:]
}

//每个将位
type BattleArmyLoc struct {
	//等级
	Lev int        `json:"l"`
	//镶嵌武将
	AvatarID int	`json:"a_id"`
}
