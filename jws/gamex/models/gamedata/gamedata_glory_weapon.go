package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	weaponControlMaxNum = 9
	weaponControlOffset = 2
)

type GloryWeaponConfig struct {
	ConfigArray      []*ProtobufGen.GLORYWEAPON
	MaxWeaponQuality uint32                               // 最大神兵品质
	WeaponMap        map[int]*ProtobufGen.GLORYWEAPONLIST // <avatarId, config>
	PolicyArray      []*ProtobufGen.GWDEVELOPRANDPOLICY
}

var GloryWeaponCfg GloryWeaponConfig

// 品质对应消耗所需的总碎片数量
func GetAllNeedGWChips(quilty int) uint32 {
	var count uint32
	for i := 1; i <= quilty; i++ {
		count += GetEvolveGloryWeaponCfg(i).GetGWChipsCount()
	}
	return count
}

// 神兵升品表
func loadGloryWeaponData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.GLORYWEAPON_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)

	GloryWeaponCfg.ConfigArray = data_ar.GetItems()
	GloryWeaponCfg.MaxWeaponQuality = data_ar.Items[len(data_ar.Items)-1].GetGloryWeaponQuality()
}

func loadGloryWeaponListData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.GLORYWEAPONLIST_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)
	GloryWeaponCfg.WeaponMap = make(map[int]*ProtobufGen.GLORYWEAPONLIST)
	for _, config := range data_ar.Items {
		GloryWeaponCfg.WeaponMap[int(config.GetID())] = config
	}
}

func loadGWDevelopRandPolicyData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.GWDEVELOPRANDPOLICY_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)
	GloryWeaponCfg.PolicyArray = data_ar.Items
}

func GetActivateGloryWeaponCost(avatarId int) (string, int) {
	listCfg := GloryWeaponCfg.WeaponMap[avatarId]
	if listCfg == nil {
		return "", 0
	}
	costItemId := listCfg.GetGWChipsID()
	costItemCount := GloryWeaponCfg.ConfigArray[0].GetGWChipsCount()
	return costItemId, int(costItemCount)
}

func ContainsGloryWeapon(avatarId int) bool {
	_, ok := GloryWeaponCfg.WeaponMap[avatarId]
	return ok
}

func GetGloryWeaponListCfg(avatarId int) *ProtobufGen.GLORYWEAPONLIST {
	return GloryWeaponCfg.WeaponMap[avatarId]
}

func GetEvolveGloryWeaponCost(avatarId, quality int) ([]string, []uint32) {
	// xi
	listCfg := GloryWeaponCfg.WeaponMap[avatarId]
	if listCfg == nil {
		return nil, nil
	}
	costId := make([]string, 2)
	costCount := make([]uint32, 2)
	costId[0] = listCfg.GetGWChipsID() // 消耗的碎片ID从list配表里面取
	costCount[0] = GloryWeaponCfg.ConfigArray[quality-1].GetGWChipsCount()
	costId[1] = GloryWeaponCfg.ConfigArray[quality-1].GetGWBreakMaterialID()
	costCount[1] = GloryWeaponCfg.ConfigArray[quality-1].GetGWBreakMaterialCount()
	return costId, costCount
}

func GetEvolveGloryWeaponCfg(quality int) *ProtobufGen.GLORYWEAPON {
	return GloryWeaponCfg.ConfigArray[quality-1]
}

func GetWeaponAttrById(cfg *ProtobufGen.GLORYWEAPON, attrId int) *ProtobufGen.GLORYWEAPON_GWAttr {
	for _, it := range cfg.GetGWAttr_Template() {
		if it.GetProperty() == uint32(attrId) {
			return it
		}
	}
	return nil
}

func getGwPolicyConfig(policyId float32) *ProtobufGen.GWDEVELOPRANDPOLICY {
	for _, it := range GloryWeaponCfg.PolicyArray {
		if it.GetGWRandPolicyID() == policyId {
			return it
		}
	}
	return nil
}

type PolicyArray []*ProtobufGen.GWDEVELOPRANDPOLICY_GWRandRange

func (p PolicyArray) GetWeight(index int) int {
	return int(p[index].GetWeight())
}

func (p PolicyArray) Len() int {
	return len(p)
}

func RandPolicyConfig(policyId float32) *ProtobufGen.GWDEVELOPRANDPOLICY_GWRandRange {
	policyCfg := getGwPolicyConfig(policyId)
	policeList := PolicyArray(policyCfg.GetGWRandRange_Template())
	logs.Debug("%d %v", policyId, policeList)
	index := util.RandomItem(policeList)
	return policeList[index]
}
