package gamedata

import (
	"math/rand"
	"sort"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	normalPoolMaxCount = 4
	GachaMaxCount      = 18

	gachaPoolRanderInitSize = 32
	gachaLevelSize          = 32 // 等级段最大
)

type GachaReward struct {
	Id    string
	Count uint32
	Give  CostData
}

type GachaRewardPoolForOneAvatar struct {
	PoolId   uint32
	ToSelect []GachaReward
	Rander   util.RandIntSet
}

type GachaRewardPool struct {
	PoolId  uint32
	Rewards GachaRewardPoolForOneAvatar
}

func (g *GachaRewardPool) GetByAvatar() *GachaRewardPoolForOneAvatar {
	return &g.Rewards
}

func (g *GachaRewardPool) Init(id uint32) {
	g.PoolId = id
	g.Rewards.Init(id)
}

func (g *GachaRewardPool) loadData(data *ProtobufGen.GACHAGROUP_ARRAY) {
	items := data.GetItems()
	for _, reward := range items {
		if g.PoolId == reward.GetGroupID() {
			if reward.GetRoleLimit() > 0 { // 只用第一个角色的数据
				continue
			}
			g.Rewards.add(reward)
		}
	}
	g.Rewards.Rander.Make()
}

func (g *GachaRewardPoolForOneAvatar) Init(id uint32) {
	g.PoolId = id
	g.ToSelect = make([]GachaReward, 0, gachaPoolRanderInitSize)
	g.Rander.Init(gachaPoolRanderInitSize)
}

func (g *GachaRewardPoolForOneAvatar) add(data *ProtobufGen.GACHAGROUP) {
	idx := len(g.ToSelect)
	gd := &GachaReward{
		data.GetItemID(),
		data.GetCount(),
		CostData{},
	}

	gd.Give.AddItem(data.GetItemID(), data.GetCount())

	g.ToSelect = append(g.ToSelect, *gd)
	ok := g.Rander.Add(idx, data.GetWeight())
	if !ok {
		logs.Error("GachaRewardPoolForOneAvatar Add To Rander err  %d %v ", idx, data)
	}
}

type GachaData struct {
	GachaID uint32

	LevelMin uint32
	LevelMax uint32

	NormalPoolRander util.RandUInt32Set
	NormalPool       [normalPoolMaxCount]GachaRewardPool //普通组

	SpecInNormalNum   uint32          // 普通组中出现特殊组的Num
	SpecInNormalSpace uint32          // 普通组中出现特殊组的Space
	SpecPool          GachaRewardPool // 特殊组

	TreasureInSpaceNum   uint32          // 特殊组中出现珍品组的Num
	TreasureInSpaceSpace uint32          // 特殊组中出现珍品组的Space
	TreasurePool         GachaRewardPool // 珍品组

	RewardSerialId uint32
	RewardSerial   [AVATAR_NUM_CURR][]GachaReward // 十连抽额外奖励序列

	CostForOneCoin   CostData // 单次抽奖消耗货币
	CostForTenCoin   CostData // 十连抽消耗货币
	CostForOneTicket CostData // 单次抽奖消耗奖券
	CostForTenTicket CostData // 十次抽奖消耗奖券

	CostForOne_Typ   string
	CostForOne_Count uint32
	CostForTen_Typ   string
	CostForTen_Count uint32

	CostForOne_TTyp   string
	CostForOne_TCount uint32
	CostForTen_TTyp   string
	CostForTen_TCount uint32

	GachaCategory uint32

	GiveForOne PriceData //显示在界面上用于购买的物品，即买XX随机送某些东西
	GiveForTen PriceData //显示在界面上用于购买的物品，即买XX随机送某些东西

	FreeCoolTime         int64
	FreeCountEveryOneDay int
	FirstGive            GachaReward

	ExtraGroupRewardPool GachaRewardPool
	ExtraSpace           uint32
	ExtraStartNum        uint32

	AfricaNumber uint32 //前x-1次没抽到M时，第x次必给
	ItemID       string //M的ID
	ItemNum      uint32 //M的个数

}

func (g *GachaData) loadData(data *ProtobufGen.NORMALGACHA) {
	g.GachaID = data.GetGachaType()

	g.NormalPoolRander.Init(4)
	g.LevelMin = data.GetLevelMin()
	g.LevelMax = data.GetLevelMax()

	if data.GetItemGroupID1() != 0 {
		g.NormalPoolRander.Add(0, data.GetWeight1())
		g.NormalPool[0].Init(data.GetItemGroupID1())
	}
	if data.GetItemGroupID2() != 0 {
		g.NormalPoolRander.Add(1, data.GetWeight2())
		g.NormalPool[1].Init(data.GetItemGroupID2())
	}
	if data.GetItemGroupID3() != 0 {
		g.NormalPoolRander.Add(2, data.GetWeight3())
		g.NormalPool[2].Init(data.GetItemGroupID3())
	}
	if data.GetItemGroupID4() != 0 {
		g.NormalPoolRander.Add(3, data.GetWeight4())
		g.NormalPool[3].Init(data.GetItemGroupID4())
	}

	if !g.NormalPoolRander.Make() {
		logs.Error("NormalPoolRander Err %v", g.NormalPoolRander)
	}

	g.SpecInNormalNum = data.GetNumForSpecial()
	g.SpecInNormalSpace = data.GetSpaceForSpecial()
	g.SpecPool.Init(data.GetSpecialGroupID())

	g.TreasureInSpaceNum = 1 // N次出一个
	g.TreasureInSpaceSpace = data.GetSpaceForTreasure()
	g.TreasurePool.Init(data.GetTreasureGroupID())

	g.RewardSerialId = data.GetRewardSerialID()
}

func (g *GachaData) loadCommonData(data *ProtobufGen.GACHASETTINGS) {
	g.CostForOne_Typ = data.GetGachaCoin()
	g.CostForOne_Count = data.GetAPrice()
	g.CostForTen_Typ = data.GetGachaCoin()
	g.CostForTen_Count = data.GetTenPrice()

	g.CostForOne_TTyp = data.GetGachaTicket()
	g.CostForOne_TCount = data.GetTAPrice()
	g.CostForTen_TTyp = data.GetGachaTicket()
	g.CostForTen_TCount = data.GetTTenPrice()

	g.GachaCategory = data.GetGachaCategory()

	g.CostForOneCoin.AddItem(data.GetGachaCoin(), data.GetAPrice())
	g.CostForTenCoin.AddItem(data.GetGachaCoin(), data.GetTenPrice())
	g.CostForOneTicket.AddItem(data.GetGachaTicket(), data.GetTAPrice())
	g.CostForTenTicket.AddItem(data.GetGachaTicket(), data.GetTTenPrice())

	g.FreeCoolTime = int64(data.GetFreeTime()) * 60    // 以分钟为单位
	g.FreeCountEveryOneDay = int(data.GetDailyLimit()) // 以分钟为单位

	g.FirstGive.Id = data.GetFirstGachaItem()
	g.FirstGive.Count = data.GetFirstGachaCount()
	g.FirstGive.Give.AddItem(g.FirstGive.Id, g.FirstGive.Count)

	if data.GetGachaItem() != "" {
		g.GiveForOne.AddItem(data.GetGachaItem(), 1)
		g.GiveForTen.AddItem(data.GetGachaItem(), 10)
	}

	g.ExtraGroupRewardPool.Init(data.GetExtraGroup())
	g.ExtraSpace = data.GetExtraSpace()
	g.ExtraStartNum = data.GetExtraStartNum()
	g.AfricaNumber = data.GetAfricaNumber()
	g.ItemID = data.GetItemID()
	g.ItemNum = data.GetItemNum()
}

func (g *GachaData) loadPoolData(data *ProtobufGen.GACHAGROUP_ARRAY) {
	for i := 0; i < len(g.NormalPool); i++ {
		g.NormalPool[i].loadData(data)
	}

	g.SpecPool.loadData(data)
	g.TreasurePool.loadData(data)
}

func (g *GachaData) loadExtPoolData(data *ProtobufGen.GACHAGROUP_ARRAY) {
	g.ExtraGroupRewardPool.loadData(data)
	logs.Trace("g.ExtraGroupRewardPool %v", g.ExtraGroupRewardPool)
}

func (g *GachaData) loadRewardSerialData(data *ProtobufGen.REWARDSERIAL_ARRAY) {
	items := data.GetItems()
	// 对不同武将有不同的序列 T2206
	for i := 0; i < AVATAR_NUM_CURR; i++ {
		g.RewardSerial[i] = make([]GachaReward, 0, gachaPoolRanderInitSize)
	}
	for _, rs := range items {
		if g.RewardSerialId == rs.GetSerialID() {
			avatar_id := int(rs.GetRoleLimit())
			if avatar_id < 0 || avatar_id >= AVATAR_NUM_CURR {
				logs.Error("RewardSerial Avatar_id Err by %v", rs)
			}
			gr := &GachaReward{
				rs.GetItemID(),
				rs.GetCount(),
				CostData{},
			}
			gr.Give.AddItem(rs.GetItemID(), rs.GetCount())
			for int(rs.GetSubID()-1) >= len(g.RewardSerial[avatar_id]) {
				g.RewardSerial[avatar_id] = append(g.RewardSerial[avatar_id], GachaReward{})
			}
			g.RewardSerial[avatar_id][int(rs.GetSubID()-1)] = *gr // 表中从1开始
		}
	}
}

var (
	gdGachaData [GachaMaxCount][]GachaData
)

func loadGachaDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.NORMALGACHA_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	for _, ga := range dataList.GetItems() {
		id := int(ga.GetGachaType()) - 1 // 表中从1开始

		if id < 0 || id > len(gdGachaData) {
			logs.Error("Unknown Gacha %v", id)
			continue
		}

		if gdGachaData[id] == nil {
			gdGachaData[id] = make([]GachaData, 0, gachaLevelSize)
		}

		l := len(gdGachaData[id])
		gdGachaData[id] = append(gdGachaData[id], GachaData{})
		gdGachaData[id][l].loadData(ga)
	}
	logs.Debug("gdGachaData %v", gdGachaData)
}

func loadGachaPoolDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GACHAGROUP_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	for i := 0; i < len(gdGachaData); i++ {
		for j := 0; j < len(gdGachaData[i]); j++ {
			gdGachaData[i][j].loadPoolData(dataList)
		}
	}
}

func loadGachaExtPoolDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GACHAGROUP_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	for i := 0; i < len(gdGachaData); i++ {
		for j := 0; j < len(gdGachaData[i]); j++ {
			gdGachaData[i][j].loadExtPoolData(dataList)
		}
	}
}

func loadGachaRewardDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.REWARDSERIAL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	for i := 0; i < len(gdGachaData); i++ {
		for j := 0; j < len(gdGachaData[i]); j++ {
			gdGachaData[i][j].loadRewardSerialData(dataList)
			//logs.Trace("gdGachaData %d -> %v", i, gdGachaData[i][j])
		}
	}
}

func loadGachaCommonDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GACHASETTINGS_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	for _, ga := range dataList.GetItems() {
		id := int(ga.GetGachaType()) - 1 // 表中从1开始

		if id < 0 || id > len(gdGachaData) {
			logs.Error("Unknown Gacha %v", id)
			continue
		}

		if gdGachaData[id] == nil {
			//			logs.Error("Unknown Gacha Data ToFill %v", id)
			continue
		}

		l := len(gdGachaData[id])
		for i := 0; i < l; i++ {
			gdGachaData[id][i].loadCommonData(ga)
			//logs.Trace("gdGachaData %d -> %v", id, gdGachaData[id][i])
		}
	}

}

func GetGachaData(corp_lv uint32, typ int) *GachaData {
	if typ < 0 || typ >= len(gdGachaData) {
		logs.Error("Unknown Gacha Typ %d", typ)
		return nil
	}

	res_idx := sort.Search(len(gdGachaData[typ]),
		func(i int) bool {
			return gdGachaData[typ][i].LevelMax >= corp_lv
		})

	if res_idx >= len(gdGachaData[typ]) {
		return nil
	}

	return &gdGachaData[typ][res_idx]
}

var (
	gdGachaProbabilityToSpecPool [HC_TYPE_COUNT]float32
)

func loadGachaProbabilityToSpecPoolDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.HCINFLUENCE_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	if len(dataList.GetItems()) < 1 {
		logs.Error("loadGachaProbabilityToSpecPoolDataConfig No Data")
	}

	data := dataList.GetItems()[0]
	gdGachaProbabilityToSpecPool[HC_From_Buy] = data.GetHC_Buy()
	gdGachaProbabilityToSpecPool[HC_From_Give] = data.GetHC_Give()
	gdGachaProbabilityToSpecPool[HC_From_Compensate] = data.GetHC_Compensate()

	//logs.Trace("gdGachaProbabilityToSpecPool %v",
	//	gdGachaProbabilityToSpecPool)
}

func IsGachaToSpecPool(hc_t int, rd *rand.Rand) bool {
	if hc_t < 0 || hc_t >= HC_TYPE_COUNT {
		logs.Error("Unknown hc_t %d", hc_t)
		return false
	}
	// 0.0001 为了防止误差
	k := rd.Float32() - 0.0001
	//logs.Trace("IsGachaToSpecPool %d %d %d", k, gdGachaProbabilityToSpecPool[hc_t], hc_t)
	return k <= gdGachaProbabilityToSpecPool[hc_t]
}

func IsHeroSurplusGacha(typ int) bool {
	return typ == GachaType_Surplus3 || typ == GachaType_Surplus4 || typ == GachaType_Surplus5
}

func GetSurplusTypeById(gachaId int) int {
	switch gachaId {
	case GachaType_Surplus3:
		return helper.Hero_Surplus_3
	case GachaType_Surplus4:
		return helper.Hero_Surplus_4
	case GachaType_Surplus5:
		return helper.Hero_Surplus_5
	}
	return -1
}
