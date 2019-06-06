package gamedata

import (
	"errors"
	"math/rand"

	"fmt"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type lootIGItem struct {
	ItemID   string
	Weight   int32
	CountMin int32
	CountMax int32
}

func (l lootIGItem) GetWeight() int32 {
	return l.Weight
}

func (l lootIGItem) GetID() string {
	return l.ItemID
}

type lootItemGroup struct {
	Infos []lootIGItem
}

type lootItemGroupCfg struct {
	Infos map[string]lootItemGroup
}

func (l *lootItemGroupCfg) Load(id string, info []*ProtobufGen.ITEMGROUP_ItemRule) {
	var new_ig_info lootItemGroup
	new_ig_info.Infos = make([]lootIGItem, 0, len(info))

	for _, r := range info {
		i := lootIGItem{
			r.GetItemID(),
			r.GetWeight(),
			r.GetCountMin(),
			r.GetCountMax(),
		}
		new_ig_info.Infos = append(new_ig_info.Infos, i)
	}
	l.Infos[id] = new_ig_info
	return
}

type lootTItem struct {
	ItemGroupID string
	Weight      int32
}

func (l lootTItem) GetWeight() int32 {
	return l.Weight
}

func (l lootTItem) GetID() string {
	return l.ItemGroupID
}

type lootTemplate struct {
	Infos []lootTItem
}

type lootTemplateCfg struct {
	Infos map[string]lootTemplate
}

func (l *lootTemplateCfg) Load(id string, info []*ProtobufGen.TEMPLATE_GroupWeight) {
	var new_info lootTemplate
	new_info.Infos = make([]lootTItem, 0, len(info))

	for _, r := range info {
		//logs.Trace("new t item %s", r)

		i := lootTItem{
			r.GetItemGroupID(),
			r.GetWeight(),
		}
		new_info.Infos = append(new_info.Infos, i)
	}
	l.Infos[id] = new_info
	return
}

func loadLootItemGroup(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	// 读取ItemGroup表
	buffer, err := loadBin(filepath)
	errcheck(err)

	items_group_info := &ProtobufGen.ITEMGROUP_ARRAY{}
	err = proto.Unmarshal(buffer, items_group_info)
	errcheck(err)

	groups := items_group_info.GetItems()

	lootItemGroupConfig := lootItemGroupCfg{make(map[string]lootItemGroup)}

	for _, ig := range groups {
		if ig.GetRoleLimit() == 0 {
			//logs.Trace("Loot Item Group ID:%s, len:%d", *ig.ID, len(ig.Rules))
			lootItemGroupConfig.Load(ig.GetID(), ig.GetRules())
		}
	}

	//logs.Trace("GDLootItemGroupConfig Info %s", lootItemGroupConfig)
	loadIGRandArrayImp(lootItemGroupConfig.Infos)
}

func loadLootTemplate(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	// 读取Template表
	buffer, err := loadBin(filepath)
	errcheck(err)

	loot_template_info := &ProtobufGen.TEMPLATE_ARRAY{}
	err = proto.Unmarshal(buffer, loot_template_info)
	errcheck(err)

	templates := loot_template_info.GetItems()

	lootTemplateConfig := lootTemplateCfg{make(map[string]lootTemplate)}
	for _, t := range templates {
		//logs.Trace("Loot Template ID:%s, len:%d", *t.ID, len(t.Rules))
		lootTemplateConfig.Load(t.GetID(), t.GetRules())
	}

	//logs.Trace("GDLootTemplateConfig Info %s", lootTemplateConfig)
	loadTRandArrayImp(lootTemplateConfig.Infos)

}

// 随机数生成器 固定seed便于重放 注意这里是所有玩家公用的，所以不走玩家自己的随机数生成器
var rander = rand.New(rand.NewSource(99))

func randInt31(min, max int32) int32 {
	return rander.Int31n(max-min) + min
}

// 权重总量 定为100
const max_weight = 100

var (
	gdLootTemplateRandArray map[string][]lootTItem
	// 根据不同的角色分别建立不同的表
	gdLootItemGroupRandArray map[string]*util.RandSet
	gdLootItemGroupCfg       map[string]map[string]*lootIGItem
)

// 根据template信息加载随机数组
func loadTRandArrayImp(ts map[string]lootTemplate) {
	gdLootTemplateRandArray = make(map[string][]lootTItem)
	ts_len := len(ts)
	all_rand_array := make([]lootTItem, ts_len*max_weight, ts_len*max_weight)

	n := 0
	for i, v := range ts {
		mem := all_rand_array[n*max_weight : n*max_weight : (n+1)*max_weight]
		n++
		gdLootTemplateRandArray[i] = makeTRandArray(v.Infos, mem)
	}

}

// 根据item group信息加载随机数组
func loadIGRandArrayImp(igs map[string]lootItemGroup) {
	gdLootItemGroupRandArray = make(map[string]*util.RandSet, len(igs))
	gdLootItemGroupCfg = make(map[string]map[string]*lootIGItem, len(igs))

	for i, v := range igs {
		rs := &util.RandSet{}
		cfg := make(map[string]*lootIGItem, len(v.Infos))
		rs.Init(len(v.Infos))
		for _, _i := range v.Infos {
			rs.Add(_i.ItemID, uint32(_i.Weight))
			cfg[_i.ItemID] = &lootIGItem{
				ItemID:   _i.ItemID,
				Weight:   _i.Weight,
				CountMin: _i.CountMin,
				CountMax: _i.CountMax,
			}
		}
		if !rs.Make() {
			panic(fmt.Errorf("loadIGRandArrayImp rs.Make err %s", i))
		}
		gdLootItemGroupCfg[i] = cfg
		gdLootItemGroupRandArray[i] = rs
	}

}

// 展开template随机数组
func makeTRandArray(info []lootTItem, mem []lootTItem) []lootTItem {
	weight_has_use := 0
	for _, r := range info {
		weight := r.GetWeight()
		var i int32
		for i = 0; i < weight; i++ {
			if weight_has_use >= max_weight {
				logs.Warn("makeTRandArray MakeRandArray weight_has_use > %d !", max_weight)
				break
			}
			mem = append(mem, r)
			weight_has_use++
		}
	}
	return mem
}

// 展开item group随机数组
func makeIGRandArray(info []lootIGItem, mem []lootIGItem) []lootIGItem {
	weight_has_use := 0

	for _, r := range info {
		weight := r.GetWeight()
		var i int32
		for i = 0; i < weight; i++ {
			if weight_has_use >= max_weight {
				logs.Warn("makeIGRandArray MakeRandArray weight_has_use > %d !", max_weight)
				logs.Warn("makeIGRandArray  %v !", info)
				break
			}
			mem = append(mem, r)
			weight_has_use++
		}
	}
	return mem
}

var (
	GetlootResNoId = errors.New("No Loot Info By Id")
)

func LootTemplateRandSelect(player_rander *rand.Rand, id string) (*lootTItem, error) {
	var r int
	if player_rander != nil {
		r = player_rander.Intn(max_weight)
	} else {
		r = rand.Intn(max_weight)
	}
	rand_array, ok := gdLootTemplateRandArray[id]
	if !ok {
		logs.Error("No tid in Info %s", id)
		return nil, GetlootResNoId
	}

	if len(rand_array) > r {
		return &rand_array[r], nil
	} else {
		return nil, nil
	}
}

func LootTemplateRand(player_rander *rand.Rand, template_id string) (PriceDatas, error) {
	item, err := LootTemplateRandSelect(player_rander, template_id)

	if err != nil {
		return PriceDatas{}, err
	}

	if item == nil {
		return PriceDatas{}, nil
	}
	return LootItemGroupRand(player_rander, item.ItemGroupID)
}

func LootItemGroupRandSelect(player_rander *rand.Rand, id string) (*lootIGItem, error) {
	rand_s, ok := gdLootItemGroupRandArray[id]
	if !ok {
		logs.Error("No ig id in Info %s", id)
		return nil, GetlootResNoId
	}

	item := rand_s.Rand(player_rander)
	if item != "" {
		cfg := gdLootItemGroupCfg[id]
		return cfg[item], nil

	}
	return nil, nil
}

func LootItemGroupRand(player_rander *rand.Rand, item_group_id string) (PriceDatas, error) {
	item, err := LootItemGroupRandSelect(player_rander, item_group_id)
	p := PriceDatas{}

	if err != nil {
		return p, err
	}

	if item == nil {
		return p, nil
	}

	p.AddItem(
		item.ItemID,
		RandInt31(
			item.CountMin,
			item.CountMax,
			player_rander))

	return p, nil
}

func RandInt31(min, max int32, r *rand.Rand) uint32 {
	if max > min {
		if r == nil {
			return uint32(rand.Int31n(max-min) + min)
		}
		return uint32(r.Int31n(max-min) + min)
	} else if max == min {
		return uint32(max)
	} else {
		logs.Error("randInt31 min > max!")
		return RandInt31(max, min, r)
	}
}

// 随机掉落计算接口 通过掉落Template随机
func GetGivesByTemplate(template_id, accountID string, pRand *rand.Rand) (PriceDatas, error) {
	item, err := LootTemplateRandSelect(pRand, template_id)

	if err != nil {
		return PriceDatas{}, err
	}

	if item == nil {
		return PriceDatas{}, nil
	}
	return GetGivesByItemGroup(item.ItemGroupID, accountID, pRand)
}

// 随机掉落计算接口 通过掉落ItemGroup随机
func GetGivesByItemGroup(item_group_id, accountID string, pRand *rand.Rand) (PriceDatas, error) {
	item, err := LootItemGroupRandSelect(pRand, item_group_id)

	res := PriceDatas{}

	if err != nil {
		return res, err
	}

	if item == nil {
		return res, nil
	}

	if !IsFixedIDItemID(item.ItemID) {
		c := RandInt31(
			item.CountMin,
			item.CountMax,
			pRand)
		for i := 0; i < int(c); i++ {
			data := MakeItemData(accountID, pRand, item.ItemID)
			res.AddItemWithData(item.ItemID, *data, 1)
		}
	} else {
		res.AddItem(
			item.ItemID,
			RandInt31(
				item.CountMin,
				item.CountMax,
				pRand))
	}

	logs.Trace("[%s][RandRes]getLootByItemGroup %v", accountID, res)
	return res, nil
}
