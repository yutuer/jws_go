package gamedata

import (
	"github.com/golang/protobuf/proto"
	//"vcs.taiyouxi.net/jws/gamex/models/helper"
	"math/rand"

	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// BOSSFIGHTPOINTS
const MaxDegree = 18 // 增加难度等级时记得增加
const BossMaxReward = 4

type bossPool struct {
	bosss []*ProtobufGen.BOSSFIGHT
}

func (b *bossPool) FromData(data *ProtobufGen.BOSSFIGHT_ARRAY) {
	items := data.GetItems()
	gdBossFightCfgCount = len(items)
	b.bosss = make([]*ProtobufGen.BOSSFIGHT, len(items), len(items))

	for idx, i := range items {
		b.bosss[idx] = i
		logs.Trace("bossPool From Data %v", *i)
	}
}

func (b *bossPool) Get(degree int, rd *rand.Rand) *ProtobufGen.BOSSFIGHT {
	idx := degree
	if idx < 0 || idx >= len(b.bosss) {
		logs.Error("bossPool Get Err by %d, %d -> %d", degree, idx)
		return nil
	}
	//logs.Trace("bossPool Get by %d, %d -> %d %v", class, degree, idx, b.pools)
	return b.bosss[idx]
}

var (
	gdBossPool          bossPool
	gdBossFightCfgCount int
)

func GetBossFightCfgCount() int {
	return gdBossFightCfgCount
}

func GetBoss(degree int, rd *rand.Rand) *ProtobufGen.BOSSFIGHT {
	return gdBossPool.Get(degree, rd)
}

func loadBossPool(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.BOSSFIGHT_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdBossPool.FromData(ar)
	logs.Trace("gdBossPool %v", gdBossPool)
}
