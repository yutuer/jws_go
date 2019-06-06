package bossfight

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Boss struct {
	BossTyp      string              `json:"typ"` // boss 对应的id
	Degree       uint32              `json:"d"`   // 难度
	Point        uint32              `json:"p"`   // 积分
	TerrainID    string              `json:"te"`  // 地形
	Nature       uint32              `json:"n"`   // 天地人属性
	MaxHp        int64               `json:"mhp"` // Hp MAx
	GS           uint32              `json:"gs"`  // 推荐战力
	LevelLimit   uint32              `json:"lv"`  // 等级限制
	RewardIDs    []string            `json:"rewards"`
	RewardCounts []uint32            `json:"counts"`
	Rewards      gamedata.PriceDatas `json:"reward"`
}

func (b *Boss) IsNil() bool {
	return b.BossTyp == ""
}

func (b *Boss) FromData(boss *ProtobufGen.BOSSFIGHT) {
	b.BossTyp = boss.GetBOSSID()

	acdata, acok := gamedata.GetAcData(b.BossTyp)
	if !acok || acdata == nil {
		logs.Error("Boss AcData Err By %s", b.BossTyp)
		return
	}

	b.GS = boss.GetBossGS()
	b.Point = boss.GetBossCurrency()
	b.MaxHp = int64(acdata.GetHitPoint())
	b.LevelLimit = boss.GetLevelLimit()
	b.TerrainID = boss.GetTerrainID()
	b.Degree = boss.GetDegre()

	b.RewardIDs = make([]string, 0, gamedata.BossMaxReward)
	b.RewardCounts = make([]uint32, 0, gamedata.BossMaxReward)
	b.Rewards = gamedata.PriceDatas{}
	if boss.GetAwardID() != "" {
		b.RewardIDs = append(b.RewardIDs, boss.GetAwardID())
		b.RewardCounts = append(b.RewardCounts, boss.GetCount())
		b.Rewards.AddItem(boss.GetAwardID(), boss.GetCount())
	}
	if boss.GetAwardID2() != "" {
		b.RewardIDs = append(b.RewardIDs, boss.GetAwardID2())
		b.RewardCounts = append(b.RewardCounts, boss.GetCount2())
		b.Rewards.AddItem(boss.GetAwardID2(), boss.GetCount2())
	}
	if boss.GetAwardID3() != "" {
		b.RewardIDs = append(b.RewardIDs, boss.GetAwardID3())
		b.RewardCounts = append(b.RewardCounts, boss.GetCount3())
		b.Rewards.AddItem(boss.GetAwardID3(), boss.GetCount3())
	}
	if boss.GetAwardID4() != "" {
		b.RewardIDs = append(b.RewardIDs, boss.GetAwardID4())
		b.RewardCounts = append(b.RewardCounts, boss.GetCount4())
		b.Rewards.AddItem(boss.GetAwardID4(), boss.GetCount4())
	}
}
