package gve_proto

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/gve_notify/post_data"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

// TODO by ljz tmp data
type GVEGameDatas struct {
	BossAcDatas []*ProtobufGen.GVEENEMY
	BossModel   []*ProtobufGen.GVEMODEL
	PlayerDatas []post_data.StartGVEPostResData
}

func (g *GVEGameDatas) AppendAccount(acID string, avatarData *post_data.StartGVEPostResData) *helper.Avatar2ClientByJson {
	g.PlayerDatas = append(g.PlayerDatas, *avatarData)
	return &g.PlayerDatas[len(g.PlayerDatas)-1].Data
}

func (g *GVEGameDatas) AppendBoss(boss *ProtobufGen.GVEENEMY, model *ProtobufGen.GVEMODEL) {
	g.BossAcDatas = append(g.BossAcDatas, boss)
	g.BossModel = append(g.BossModel, model)
}
