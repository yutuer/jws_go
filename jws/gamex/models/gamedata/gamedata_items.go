package gamedata

import (
	"errors"
	"fmt"

	"strconv"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type ItemIdx_t uint32

const ItemIdx_nil = ItemIdx_t(0)

type GDItem struct {
	BagId  uint32
	ItemId ItemIdx_t
}

var (
	gdItems                []*ProtobufGen.Item
	gdItemsFixedIDs        map[string]GDItem
	gdItemsMap             map[string]*ProtobufGen.Item
	gdItemsAttrAddonMap    map[string]*avatarAttrAddon
	gdFashionMap           map[string]*ProtobufGen.Item
	gdJadeMap              map[string]*ProtobufGen.Item
	gdStarSoulMap          map[string]*ProtobufGen.Item
	gdStarSoulMapWithParam map[string]map[uint32]map[uint32]*ProtobufGen.Item
	gdWholeCharId          [AVATAR_NUM_MAX]string
	gdJadeLvl2Xp           []uint32
	gdJadeSlot2Lvl2Item    map[int]map[int32]string
	JadeMaxLvl             uint32
	JadeMaxExp             uint32
	gdPackage              map[string]*ProtobufGen.Item
	// 表中存储的Part是一个类似于Weapon的字符串，
	// 而实际用到的是一个int索引，这里单独生成一个map
	//gdItemsPart map[string]int
)

const (
	PartID_Weapon   = helper.PartID_Weapon
	PartID_Chest    = helper.PartID_Chest
	PartID_Necklace = helper.PartID_Necklace
	PartID_Belt     = helper.PartID_Belt
	PartID_Ring     = helper.PartID_Ring
	PartID_Leggings = helper.PartID_Leggings
	PartID_Bracers  = helper.PartID_Bracers
	PartEquipCount  = helper.PartEquipCount
)

const (
	GuildMedal      = helper.GuildMedal
	GuildMedalCount = helper.GuildMedalCount
)

const (
	FashionPart_Weapon = helper.FashionPart_Weapon
	FashionPart_Armor  = helper.FashionPart_Armor
	FashionPart_Count  = helper.FashionPart_Count
)

const (
	JadePart_0    = helper.JadePart_0
	JadePart_1    = helper.JadePart_1
	JadePart_2    = helper.JadePart_2
	JadePart_3    = helper.JadePart_3
	JadePart_4    = helper.JadePart_4
	JadePart_5    = helper.JadePart_5
	JadePartCount = helper.JadePartCount
)

const (
	RareLv_White = iota
	RareLv_Green
	RareLv_Blue
	RareLv_Purple
	RareLv_Gold
	RareLv_Red
	RareLv_Max
)

func GetEquipSlotNum() int {
	return PartEquipCount
}

// 上面前7个是装备
const Fashion_slotid_start = FashionPart_Weapon
const Fashion_slotid_end = FashionPart_Armor

// 龙玉装备part
func GetJadeSlot(part string) int {
	return helper.GetJadeSlot(part)
}

// 获取一个Part代表的装备位置索引，如果不是装备，返回-1
func GetEquipSlot(part string) int {
	return helper.GetEquipSlot(part)
}

func GetWholeCharIdByAvatarId(avatar int) string {
	return gdWholeCharId[avatar]
}

// 获取avatar专属装备Part
func GetAvatarEquipSlot(part string) int {
	return helper.GetAvatarEquipSlot(part)
}

func LoadItemData(filepath string) {
	loadItemData(filepath)
}
func loadItemData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	// 读取武器装备表
	buffer, err := loadBin(filepath)
	errcheck(err)

	itemsList := &ProtobufGen.Item_ARRAY{}
	err = proto.Unmarshal(buffer, itemsList)
	errcheck(err)

	gdItems = itemsList.GetItems()
	gdItemsMap = make(map[string]*ProtobufGen.Item, len(gdItems))
	gdItemsFixedIDs = make(map[string]GDItem, len(gdItems))
	gdItemsAttrAddonMap = make(map[string]*avatarAttrAddon, len(gdItems))
	gdFashionMap = make(map[string]*ProtobufGen.Item, 100)
	gdJadeMap = make(map[string]*ProtobufGen.Item, 100)
	gdStarSoulMap = make(map[string]*ProtobufGen.Item)
	gdStarSoulMapWithParam = make(map[string]map[uint32]map[uint32]*ProtobufGen.Item)
	gdJadeSlot2Lvl2Item = make(map[int]map[int32]string, 10)
	gdPackage = make(map[string]*ProtobufGen.Item, len(gdItems))
	//gdItemsPart = make(map[string]int, len(gdItems))

	counter := uint32(1) //XXX 0 is reserved
	for _, item := range gdItems {
		var gditem GDItem
		if IsFixedIDItem(item.GetType()) {
			gditem.BagId = counter
		}
		gditem.ItemId = ItemIdx_t(counter)
		gdItemsFixedIDs[item.GetID()] = gditem
		gdItemsMap[item.GetID()] = item

		// 计算GS用
		addon := &avatarAttrAddon{}
		addon.AddEquipAddon(item.GetAttack(), item.GetDefense(), item.GetHP())
		gdItemsAttrAddonMap[item.GetID()] = addon
		if item.GetType() == "Fashion" {
			gdFashionMap[item.GetID()] = item
		} else if item.GetType() == "JADE" {
			gdJadeMap[item.GetID()] = item
			slot := GetJadeSlot(item.GetPart())
			if slot < 0 {
				panic(errors.New(fmt.Sprintf("jade part not found, %s %s", item.GetID(), item.GetPart())))
			}
			lvl2Item, ok := gdJadeSlot2Lvl2Item[slot]
			if !ok {
				lvl2Item = make(map[int32]string, JadeMaxLvl)
				gdJadeSlot2Lvl2Item[slot] = lvl2Item
			}
			lvl2Item = gdJadeSlot2Lvl2Item[slot]
			old, ok := lvl2Item[item.GetJadeLevel()]
			if ok {
				panic(errors.New(fmt.Sprintf("jade slot %d lvl %d item repeat %s %s", slot, item.GetJadeLevel(), item.GetID(), old)))
			}
			lvl2Item[item.GetJadeLevel()] = item.GetID()
			gdJadeSlot2Lvl2Item[slot] = lvl2Item
		} else if item.GetType() == "STARSOUL" {
			gdStarSoulMap[item.GetID()] = item
			if _, ok := gdStarSoulMapWithParam[item.GetPart()]; false == ok {
				gdStarSoulMapWithParam[item.GetPart()] = make(map[uint32]map[uint32]*ProtobufGen.Item)
			}
			if _, ok := gdStarSoulMapWithParam[item.GetPart()][item.GetStarHole()]; false == ok {
				gdStarSoulMapWithParam[item.GetPart()][item.GetStarHole()] = make(map[uint32]*ProtobufGen.Item)
			}
			gdStarSoulMapWithParam[item.GetPart()][item.GetStarHole()][uint32(item.GetRareLevel())] = item
		} else if item.GetType() == "Package" {
			gdPackage[item.GetID()] = item
		}

		counter++
	}
}

func loadJadeData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	// 读取武器装备表
	buffer, err := loadBin(filepath)
	errcheck(err)

	l := &ProtobufGen.JADEXP_ARRAY{}
	err = proto.Unmarshal(buffer, l)
	errcheck(err)

	var maxlvl uint32
	for _, e := range l.GetItems() {
		if e.GetJadeLevel() > maxlvl {
			maxlvl = e.GetJadeLevel()
		}
	}
	JadeMaxLvl = maxlvl
	gdJadeLvl2Xp = make([]uint32, maxlvl, maxlvl)
	for _, e := range l.GetItems() {
		gdJadeLvl2Xp[e.GetJadeLevel()-1] = e.GetJadeXP()
		if e.GetJadeXP() > JadeMaxExp {
			JadeMaxExp = e.GetJadeXP()
		}
	}
}

func GetItemIDbyIdx(idx ItemIdx_t) string {
	ic := uint32(idx) - 1
	if ic < 0 || int(ic) >= len(gdItems) {
		logs.Error("GetItemIDbyIdx ic Err %v", ic)
		return ""
	}
	return gdItems[ic].GetID()
}

func GetItemID(ItemID string) ItemIdx_t {
	if id, ok := gdItemsFixedIDs[ItemID]; ok {
		return id.ItemId
	}
	return 0
}

func GetItemDataByList(ids []string) *ProtobufGen.Item_ARRAY {
	result := make([]*ProtobufGen.Item, len(ids))
	for idx, iname := range ids {
		if item, ok := gdItemsMap[iname]; ok {
			result[idx] = item
		} else {
			result[idx] = &ProtobufGen.Item{}
		}
	}
	return &ProtobufGen.Item_ARRAY{
		Items: result,
	}
}

func GetProtoItem(ItemID string) (*ProtobufGen.Item, bool) {
	//logs.Trace("GetProtoItem %s", ItemID)
	item, ok := gdItemsMap[ItemID]
	return item, ok
}

func GetProtoItemAttrAddon(ItemID string) *avatarAttrAddon {
	item, ok := gdItemsAttrAddonMap[ItemID]
	if !ok {
		return nil
	} else {
		return item
	}
}

func IsItemEquip(ItemID string) bool {
	data, ok := GetProtoItem(ItemID)
	if !ok {
		return false
	}

	item_type := data.GetType()
	if item_type == "Weapon" ||
		item_type == "Ornament" ||
		item_type == "Armor" ||
		item_type == "GuildMedal" {
		return true
	}
	return false
}

// 这个是所有记录在背包中个数的
func IsItemNeedCountToBag(ItemID string) bool {
	data, ok := GetProtoItem(ItemID)
	if !ok {
		return false
	}

	item_type := data.GetType()
	if item_type == "Weapon" ||
		item_type == "Ornament" ||
		item_type == "Armor" {
		return true
	}
	return false
}

func IsFixedIDItem(item_type string) bool {
	if item_type == "Weapon" ||
		item_type == "Ornament" ||
		item_type == "Armor" ||
		item_type == "Fashion" ||
		item_type == "JADE" ||
		item_type == "GuildMedal" {
		return false
	}
	return true
}

func IsFixedIDItemID(itemID string) bool {
	if IsItemVirtual(itemID) {
		return true
	}
	data, ok := GetProtoItem(itemID)
	if !ok {
		return true
	}
	return IsFixedIDItem(data.GetType())
}

func GetFixedBagID(ItemID string) (uint32, bool) {
	if id, ok := gdItemsFixedIDs[ItemID]; ok {
		return id.BagId, (id.BagId != 0)
	}
	return 0, false
}

// 获取道具信息
func GetItemDataByIdx(idx ItemIdx_t) *ProtobufGen.Item {
	ic := uint32(idx) - 1
	if ic < 0 || int(ic) >= len(gdItems) {
		logs.Error("GetItemInfoByIdx ic Err %v", ic)
		return nil
	}
	return gdItems[ic]
}

// 道具是否在获得时折算成sc
// 如果是则顺便把sc数值传回去
// 见 http://wiki.taiyouxi.net/T751
func IsItemToSCWhenAdd(ItemID string) (bool, string, uint32) {
	item, ok := gdItemsMap[ItemID]
	if !ok {
		return false, "", 0
	}

	if item.GetType() == "VirtualItem" && item.GetAttrValue() > 0 {
		return true, item.GetAttrType(), item.GetAttrValue()
	}

	return false, "", 0
}

func IsItemTreasurebox(ItemId string) (bool, string, uint32) {
	item, ok := gdItemsMap[ItemId]
	if !ok {
		return false, "", 0
	}
	//logs.Warn("IsItemTreasurebox %v %v", ItemId, *item)
	if item.GetPart() == "Treasurebox" {
		return true, item.GetAttrType(), item.GetAttrValue()
	}

	return false, "", 0
}

func IsItemIdKnownBeforeGive(itemID string) bool {
	if itemID == helper.MatEevoUniversalItemID || itemID == helper.VI_HcByVIP {
		return true
	}
	t, _, _, _ := IsItemToWholeCharWhenAdd(itemID)
	if t {
		return true
	}
	return false
}

func IsItemToSCNoGoodWillWhenAdd(ItemID string) (bool, string, uint32) {
	item, ok := gdItemsMap[ItemID]
	if !ok {
		return false, "", 0
	}

	if item.GetType() == "VirtualItem" && item.GetPart() != "GeneralGoodwill" {
		return true, item.GetAttrType(), item.GetAttrValue()
	}

	return false, "", 0
}

func IsItemToBuffWhenAdd(ItemID string) bool {
	item, ok := gdItemsMap[ItemID]
	if !ok {
		return false
	}

	if item.GetType() == "BuffItem" {
		// BuffItem 不作处理
		return true
	}

	return false
}

func IsItemToWholeCharWhenAdd(ItemID string) (bool, string, uint32, *ProtobufGen.Item) {
	item, ok := gdItemsMap[ItemID]
	if !ok {
		return false, "", 0, nil
	}

	if item.GetType() == "WholeChar" {
		return true, item.GetAttrType(), item.GetAttrValue(), item
	}

	return false, "", 0, nil
}

func IsGeneralGoodwillItem(ItemID string) (bool, string, uint32) {
	item, ok := gdItemsMap[ItemID]
	if !ok {
		return false, "", 0
	}

	if item.GetType() == "VirtualItem" && item.GetPart() == "GeneralGoodwill" {
		// AttrValue里的值不读，缺省为1表示该道具作为副将友好度的基础单位
		return true, item.GetAttrType(), 1
	}

	return false, "", 0
}

func IsHeroPieceItem(ItemID string) (bool, string, uint32, *ProtobufGen.Item) {
	item, ok := gdItemsMap[ItemID]
	if !ok {
		return false, "", 0, nil
	}

	if item.GetType() == "VirtualItem" && item.GetPart() == "HeroGoodwill" {
		// AttrValue里的值不读，缺省为1表示该道具作为副将友好度的基础单位
		return true, item.GetAttrType(), 1, item
	}

	return false, "", 0, nil
}

func IsItemIAP(itemId string) (bool, uint32) {
	item, ok := gdItemsMap[itemId]
	if !ok {
		return false, 0
	}
	if item.GetType() == "VirtualItem" && item.GetPart() == "IAPITEM" {
		idx, err := strconv.Atoi(item.GetAttrType())
		if err != nil {
			logs.Error("Item Gamedata IsItemIAP Atoi err %v", err)
			return false, 0
		}
		return true, uint32(idx)
	}
	return false, 0
}

// 获取装备的TrickRuleID，用来随机附加属性
// 如果不是装备则返回false
func GetEquipTrickRuleID(ItemID string) (string, bool) {
	item, ok := gdItemsMap[ItemID]
	//logs.Trace("Item %s %v", ItemID, item)
	if !ok {
		return "", false
	}

	if IsFixedIDItem(item.GetType()) {
		return "", false
	}
	return item.GetTrickRuleID(), true
}

func IsSlotFashion(slot int) bool {
	return slot < FashionPart_Count
}

// 装备是否有天地人
func IsTCJ(ItemID string) bool {
	item, ok := gdItemsMap[ItemID]
	//logs.Trace("Item %s %v", ItemID, item)
	if !ok {
		return false
	}
	return item.GetIsTDR() > 0
}

func IsFashion(itemId string) (bool, *ProtobufGen.Item) {
	item, ok := gdFashionMap[itemId]
	return ok, item
}

func CalcFashionTime(fashionCfg *ProtobufGen.Item, now_time int64) (newTime int64) {
	if fashionCfg.GetTimeLimit() == fashion_perm_timemode {
		return fashion_perm_timemode
	}
	newTime = now_time + int64(fashionCfg.GetTimeLimit()*Hour2Second)
	return
}

const fashion_perm_timemode = 99999

func IsFashionPerm(fashionCfg *ProtobufGen.Item, fashionTime int64) bool {
	return fashionCfg.GetTimeLimit() == fashion_perm_timemode && fashionTime == fashion_perm_timemode
}

func IsJade(itemId string) (bool, *ProtobufGen.Item) {
	cfg, ok := gdJadeMap[itemId]
	return ok, cfg
}

func GetJadeExpByLvl(lvl int32) uint32 {
	if int(lvl) > len(gdJadeLvl2Xp) {
		return 0
	}
	return gdJadeLvl2Xp[lvl-1]
}

func GetJadeLvlByExp(exp uint32) int32 {
	var lvl int32
	for l, e := range gdJadeLvl2Xp {
		if exp >= e {
			lvl = int32(l + 1)
		} else {
			break
		}
	}
	return lvl
}

func GetJadeBySlotLvl(slot int, lvl int32) (bool, string) {
	lvl2Item, ok := gdJadeSlot2Lvl2Item[slot]
	if !ok {
		return false, ""
	}
	item, ok := lvl2Item[lvl]
	if !ok {
		return false, ""
	}
	return true, item
}

func Slot2String(slotId int) string {
	switch slotId {
	case PartID_Weapon:
		return "weapon"
	case PartID_Chest:
		return "chest"
	case PartID_Necklace:
		return "necklace"
	case PartID_Belt:
		return "belt"
	case PartID_Ring:
		return "ring"
	case PartID_Leggings:
		return "leg"
	case PartID_Bracers:
		return "bracers"
	default:
		return ""
	}
}

//CheckItemIsStarSoul ..
func CheckItemIsStarSoul(id string) bool {
	_, is := gdStarSoulMap[id]
	return is
}

//GetItemStarSoul ..
func GetItemStarSoul(id string) *ProtobufGen.Item {
	return gdStarSoulMap[id]
}

func GetPackage(itemId string) (bool, *ProtobufGen.Item) {
	if pkg, ok := gdPackage[itemId]; ok {
		return true, pkg
	}
	return false, nil
}
