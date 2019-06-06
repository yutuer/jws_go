package gvg

import (
	"fmt"
	"testing"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

func Test_random(t *testing.T) {
	tplayerInfo := [GVG_AVATAR_COUNT]helper.AvatarState{}
	tdestinySkill := [helper.DestinyGeneralSkillMax]int{}

	play := []*FightPlayer{
		&FightPlayer{"1", "1", "1", "1", "1", 0, 0,
			0, false, 0, tplayerInfo, tdestinySkill},
		&FightPlayer{"2", "2", "2", "2", "2", 0, 0,
			0, false, 0, tplayerInfo, tdestinySkill},
		&FightPlayer{"3", "3", "3", "3", "3", 0, 0,
			0, false, 0, tplayerInfo, tdestinySkill},
		&FightPlayer{"4", "4", "4", "4", "4", 0, 0,
			0, false, 0, tplayerInfo, tdestinySkill},
		&FightPlayer{"5", "5", "5", "5", "5", 0, 0,
			0, false, 0, tplayerInfo, tdestinySkill},
		&FightPlayer{"6", "6", "6", "6", "6", 0, 0,
			0, false, 0, tplayerInfo, tdestinySkill},
	}
	Len := len(play)
	for i := 0; i < Len; i++ {
		fmt.Print(play[i].acID)
	}
	fmt.Println()
	tplay := randFightPlayers(play)

	fmt.Println(Len)

	for i := 0; i < Len; i++ {
		fmt.Print(tplay[i].acID)
	}
}
