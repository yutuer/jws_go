package gvg_proto

import (
	"fmt"

	"vcs.taiyouxi.net/jws/multiplayer/helper"
)

type GVGPlayerData struct {
	Data []*GVGHeroData
	Pos  int
}

type GVGHeroData struct {
	HP int
	ID int32
}

type GVGGameDatas struct {
	PlayerDatas map[string]*GVGPlayerData
}

func (g *GVGGameDatas) AddPlayer(data *helper.GVGStartFightData) {
	g.PlayerDatas = make(map[string]*GVGPlayerData, 2)
	avatarData := make([]*GVGHeroData, 0)

	for _, item := range data.Avatar1 {
		avatarData = append(avatarData, &GVGHeroData{
			ID: int32(item),
		})
	}
	g.PlayerDatas[data.Acid1] = &GVGPlayerData{
		Data: avatarData,
		Pos:  0,
	}

	for _, item := range data.Avatar2 {
		avatarData = append(avatarData, &GVGHeroData{
			ID: int32(item),
		})
	}
	g.PlayerDatas[data.Acid2] = &GVGPlayerData{
		Data: avatarData,
		Pos:  1,
	}
}
func (g *GVGGameDatas) AddPlayerHP(acid string, avatarID int32, hp int) error {
	p, ok := g.PlayerDatas[acid]
	if !ok {
		return fmt.Errorf("no player data for acid: %v", acid)
	}
	exist := false
	for i, item := range p.Data {
		if item.ID == avatarID {
			p.Data[i].HP = hp
			exist = true
		}
	}
	if !exist {
		return fmt.Errorf("no player hero data for acid: %v, hero id: %v", acid, avatarID)
	}
	return nil
}
