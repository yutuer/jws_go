package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

// AstrologyConfig ..
type AstrologyConfig struct {
	holes   map[uint32]*ProtobufGen.STARMAP
	souls   map[uint32]*ProtobufGen.STARSOUL
	upgrade map[uint32]*starSoulUpgradeHole
	augur   map[uint32]*ProtobufGen.AUGUR
	common  *ProtobufGen.STARMAPCONFIG

	minAugurLevel uint32
	maxAugurLevel uint32
}

var gAstrologyConfig *AstrologyConfig

func getAstrologyConfigInstance() *AstrologyConfig {
	if nil == gAstrologyConfig {
		gAstrologyConfig = &AstrologyConfig{
			holes:   map[uint32]*ProtobufGen.STARMAP{},
			souls:   map[uint32]*ProtobufGen.STARSOUL{},
			upgrade: map[uint32]*starSoulUpgradeHole{},
			augur:   map[uint32]*ProtobufGen.AUGUR{},

			minAugurLevel: 0xFFFFFFFF,
			maxAugurLevel: 0,
		}
	}

	return gAstrologyConfig
}

func (a *AstrologyConfig) addUpgradeElem(elem *ProtobufGen.STARSOULUPGRADE) {
	up, exist := a.upgrade[elem.GetStarHole()]
	if false == exist {
		up = &starSoulUpgradeHole{
			rares: map[uint32]*starSoulUpgradeRare{},
		}
		a.upgrade[elem.GetStarHole()] = up
	}
	up.addELem(elem)
}

type starSoulUpgradeHole struct {
	rares map[uint32]*starSoulUpgradeRare
}

func (s *starSoulUpgradeHole) addELem(elem *ProtobufGen.STARSOULUPGRADE) {
	rare, exist := s.rares[elem.GetStarSoulRareLevel()]
	if false == exist {
		rare = &starSoulUpgradeRare{
			sets: map[uint32]*ProtobufGen.STARSOULUPGRADE{},
		}
		s.rares[elem.GetStarSoulRareLevel()] = rare
	}
	rare.sets[elem.GetStarSoulUpgradeLevel()] = elem
}

type starSoulUpgradeRare struct {
	sets map[uint32]*ProtobufGen.STARSOULUPGRADE
}

func loadAstrologyStarMapConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.STARMAP_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("Astrology STARMAP_ARRAY empty")
	}

	aCfg := getAstrologyConfigInstance()
	for _, starmap := range dataList.GetItems() {
		aCfg.holes[starmap.GetStarHoleID()] = starmap
	}
}

func loadAstrologyStarSoulConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.STARSOUL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("Astrology STARSOUL_ARRAY empty")
	}

	aCfg := getAstrologyConfigInstance()
	for _, soul := range dataList.GetItems() {
		aCfg.souls[soul.GetRareLevel()] = soul
	}
}

func loadAstrologyStarSoulUpgradeConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.STARSOULUPGRADE_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("Astrology STARSOULUPGRADE_ARRAY empty")
	}

	aCfg := getAstrologyConfigInstance()
	for _, up := range dataList.GetItems() {
		aCfg.addUpgradeElem(up)
	}
}

func loadAstrologyAugurConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.AUGUR_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("Astrology AUGUR_ARRAY empty")
	}

	aCfg := getAstrologyConfigInstance()
	for _, augur := range dataList.GetItems() {
		aCfg.augur[augur.GetAugurLevel()] = augur

		if aCfg.minAugurLevel > augur.GetAugurLevel() {
			aCfg.minAugurLevel = augur.GetAugurLevel()
		}
		if aCfg.maxAugurLevel < augur.GetAugurLevel() {
			aCfg.maxAugurLevel = augur.GetAugurLevel()
		}
	}
}

func loadAstrologyStarMapConfigConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.STARMAPCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("Astrology STARMAPCONFIG_ARRAY empty")
	}

	aCfg := getAstrologyConfigInstance()
	aCfg.common = dataList.Items[0]
}

//GetAstrologyAugurMinLevel 占星最低等级
func GetAstrologyAugurMinLevel() uint32 {
	return getAstrologyConfigInstance().minAugurLevel
}

//GetAstrologyAugurMaxLevel 占星最高等级
func GetAstrologyAugurMaxLevel() uint32 {
	return getAstrologyConfigInstance().maxAugurLevel
}

//GetAstrologyAugurCfg 取占星配置
func GetAstrologyAugurCfg(lv uint32) *ProtobufGen.AUGUR {
	return getAstrologyConfigInstance().augur[lv]
}

//GetAstrologySkipNum 取一键占星数量上限
func GetAstrologySkipNum() uint32 {
	return getAstrologyConfigInstance().common.GetOneKeyAugurNum()
}

//GetAstrologyMarqueeMin 取跑马灯发送的最低星魂品质
func GetAstrologyMarqueeMin() uint32 {
	return getAstrologyConfigInstance().common.GetHorseLampLimit()
}

//GetAstrologyIntoLimit 取星魂镶嵌条件 return:(star, level)
func GetAstrologyIntoLimit(holeID uint32) (uint32, uint32) {
	hole := getAstrologyConfigInstance().holes[holeID]
	if nil == hole {
		return 0xFFFFFFFF, 0xFFFFFFFF
	}

	return hole.GetHeroStarLimit(), hole.GetHeroLevelLimit()
}

//GetAstrologyHeroType 取武将的属性类型
func GetAstrologyHeroType(id uint32) string {
	heroCfg := GetHeroData(int(id))
	if nil == heroCfg {
		return "_no_hero_"
	}

	return heroCfg.StatSoulPart
}

//GetAstrologySoulCfg 取星魂的物品配置
func GetAstrologySoulCfg(soulID string) *ProtobufGen.Item {
	return gdStarSoulMap[soulID]
}

//GetAstrologySoulIDByParam 根据参数取星魂的物品ID
func GetAstrologySoulIDByParam(hero uint32, hole uint32, rare uint32) *ProtobufGen.Item {
	if _, ok := gdStarSoulMapWithParam[GetAstrologyHeroType(hero)]; false == ok {
		return nil
	}
	if _, ok := gdStarSoulMapWithParam[GetAstrologyHeroType(hero)][hole]; false == ok {
		return nil
	}
	return gdStarSoulMapWithParam[GetAstrologyHeroType(hero)][hole][rare]
}

//GetAstrologyAllSoulIDs 取所有的星魂物品ID
func GetAstrologyAllSoulIDs() []string {
	ids := []string{}
	for id := range gdStarSoulMap {
		ids = append(ids, id)
	}
	return ids
}

//GetAstrologyResolveLimit 一键分解星魂的品质上限
func GetAstrologyResolveLimit() uint32 {
	return getAstrologyConfigInstance().common.GetQuickResolveLimit()
}

//CheckAstrologyUpgrade 检查星魂是否能升级
func CheckAstrologyUpgrade(hole uint32, rare uint32, upgrade uint32) bool {
	tAstrologyConfig := getAstrologyConfigInstance()
	upHole := tAstrologyConfig.upgrade[hole]
	if nil == upHole {
		return false
	}

	upRare := upHole.rares[rare]
	if nil == upRare {
		return false
	}

	up := upRare.sets[upgrade]
	if nil == up {
		return false
	}
	return true
}

//GetAstrologyUpgradeMaterial 取星魂升级消耗的材料
func GetAstrologyUpgradeMaterial(hole uint32, rare uint32, upgrade uint32) map[string]uint32 {
	materials := map[string]uint32{}

	tAstrologyConfig := getAstrologyConfigInstance()
	upHole := tAstrologyConfig.upgrade[hole]
	if nil == upHole {
		return materials
	}

	upRare := upHole.rares[rare]
	if nil == upRare {
		return materials
	}

	up := upRare.sets[upgrade]
	if nil != up {
		for _, c := range up.GetSSUpgrade_Template() {
			materials[c.GetSSUpgradeMaterial()] = materials[c.GetSSUpgradeMaterial()] + c.GetSSUpgradeMaterialCount()
		}
	}
	return materials
}

//GetAstrologySoulAttr 取星魂加的属性
func GetAstrologySoulAttr(hole uint32, rare uint32, upgrade uint32) map[uint32]float32 {
	attr := map[uint32]float32{}

	tAstrologyConfig := getAstrologyConfigInstance()
	upHole := tAstrologyConfig.upgrade[hole]
	if nil == upHole {
		return attr
	}

	upRare := upHole.rares[rare]
	if nil == upRare {
		return attr
	}

	up := upRare.sets[upgrade]
	if nil != up {
		for _, c := range up.GetSSAttr_Template() {
			attr[c.GetProperty()] = attr[c.GetProperty()] + c.GetValue()
		}
	}

	return attr
}

//AstrologyTranslateSoulToMaterial 转换星魂为材料(软通)
func AstrologyTranslateSoulToMaterial(hole uint32, rare uint32, upgrade uint32) map[string]uint32 {
	materials := map[string]uint32{}

	tAstrologyConfig := getAstrologyConfigInstance()
	soulBaseCfg := tAstrologyConfig.souls[rare]
	if nil != soulBaseCfg {
		for _, c := range soulBaseCfg.GetSSResolve_Template() {
			materials[c.GetSSRResolveMaterial()] = materials[c.GetSSRResolveMaterial()] + c.GetSSRResolveMaterialCount()
		}
	}

	upHole := tAstrologyConfig.upgrade[hole]
	if nil == upHole {
		return materials
	}

	upRare := upHole.rares[rare]
	if nil == upRare {
		return materials
	}

	for cu := int(upgrade); cu >= 0; cu-- {
		up := upRare.sets[uint32(cu)]
		if nil != up {
			for _, c := range up.GetSSUpgrade_Template() {
				materials[c.GetSSUpgradeMaterial()] = materials[c.GetSSUpgradeMaterial()] + c.GetSSUpgradeMaterialCount()
			}
		}
	}

	return materials
}
