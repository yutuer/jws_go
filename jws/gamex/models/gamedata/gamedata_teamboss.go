package gamedata

import (
	"math"

	"time"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	BoxCfg             TBBoxCfg                              //设置配置
	gdTBossBoxData     []*ProtobufGen.TBOSSBOXDATA           //宝箱数据
	gdTBossHeroTypeMap map[uint32]*ProtobufGen.TBOSSHEROTYPE //根据日期随机阵容组合
	gdTBossDiffMap     map[uint32]*ProtobufGen.TBOSSMAINDATA //根据难度分的BOSS信息
	gdTBossDungeonMap  map[uint32]*ProtobufGen.TBOSSDUNGEON  //根据组合id 用来得到对应的场景和boss
	gdTBossDungeon     []*ProtobufGen.TBOSSDUNGEON           //阵容组合配置
	gdTBossEnemy       []*ProtobufGen.TBOSSENEMY             //敌人配置
	gdTBossBoxLoot     []*ProtobufGen.TBOSSBOXLOOT           //宝箱掉落配置
	gdTBossVipCtrl     []*ProtobufGen.TBOSSVIPCONTROL        //vip暗控配置
	gdTBossConfig      []*ProtobufGen.TBOSSCONFIG            //组队boss配置
	gdTRoopsMessage    []*ProtobufGen.TROOPSMESSAGE          //关卡兵配置
	gdBoxByDiff        map[uint32][]string                   // 不同难度的宝箱ID
)

type TBBoxCfg struct {
	HCOpenBoxNum   int //每天可以用钻石打开宝箱的次数
	BoxTimeMin     int //最小开箱时间
	TimeMinHc      int //最小单位花费钻石
	InviteColdDown int //发送邀请cd时间
	RedBoxCost     int //必中红箱子花费钻石
	TeamBackTime   int //X秒无法回到原队伍
	GoOutTeamTime  int //被踢后x秒无法回到员队伍
	RoomMax        int //每种难度最多显示多少个房间
	MaxWaitTime    int //最大等待时间
	BackstageTime  int
}

type TBossForRand ProtobufGen.TBOSSBOXLOOT
type TBossTypeForRand ProtobufGen.TBOSSHEROTYPE
type TBossIdAndSenceForRand ProtobufGen.TBOSSDUNGEON

func (ti TBossIdAndSenceForRand) GetWeight(index int) int {
	return int(ti.Choose_Table[index].GetChooseChance())
}

func (ti TBossIdAndSenceForRand) Len() int {
	return len(ti.Choose_Table)
}

func (tt TBossTypeForRand) GetWeight(index int) int {
	return int(tt.Type_Table[index].GetChooseChance())
}

func (tt TBossTypeForRand) Len() int {
	return len(tt.Type_Table)
}

func (tb TBossForRand) GetWeight(index int) int {
	return int(tb.Loot_Table[index].GetLootChance())
}

func (tb TBossForRand) Len() int {
	return len(tb.Loot_Table)
}

func loadTBBoxData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.TBOSSBOXDATA_ARRAY{}
	error := proto.Unmarshal(buffer, dataList)
	panicIfErr(error)
	gdTBossBoxData = dataList.GetItems()
	gdBoxByDiff = make(map[uint32][]string)
	for _, items := range dataList.GetItems() {
		gdBoxByDiff[items.GetDropLevel()] = append(gdBoxByDiff[items.GetDropLevel()], items.GetID())
	}
}

func loadTBBossConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	BoxCfg.InitTBossCfgFromLoad(buffer)

}

func (bcf *TBBoxCfg) InitTBossCfgFromLoad(buffer []byte) {
	dataList := &ProtobufGen.TBOSSCONFIG_ARRAY{}
	err := proto.Unmarshal(buffer, dataList)
	panicIfErr(err)
	for _, items := range dataList.GetItems() {
		bcf.BoxTimeMin = int(items.GetBoxTimeMin())
		bcf.TimeMinHc = int(items.GetTimeMinHC())
		bcf.HCOpenBoxNum = int(items.GetHCOpenBoxNum())
		bcf.InviteColdDown = int(items.GetInviteColdDown())
		bcf.RedBoxCost = int(items.GetRedBoxCost())
		bcf.TeamBackTime = int(items.GetTeamBackTime())
		bcf.GoOutTeamTime = int(items.GetGoOutTeamTime())
		bcf.RoomMax = int(items.GetRoomMax())
		bcf.MaxWaitTime = int(items.GetWaitTimeServer())
		bcf.BackstageTime = int(items.GetBackstageTime())
	}
}

func loadTBBossMainData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	initBossMainData(buffer)
}

func initBossMainData(buffer []byte) {
	dataList := &ProtobufGen.TBOSSMAINDATA_ARRAY{}
	err := proto.Unmarshal(buffer, dataList)
	panicIfErr(err)
	gdTBossDiffMap = make(map[uint32]*ProtobufGen.TBOSSMAINDATA)
	for _, items := range dataList.GetItems() {
		gdTBossDiffMap[items.GetBossLevel()] = items
	}
}

func loadTBBossDungeon(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.TBOSSDUNGEON_ARRAY{}
	error := proto.Unmarshal(buffer, dataList)
	panicIfErr(error)
	gdTBossDungeon = dataList.GetItems()
	gdTBossDungeonMap = make(map[uint32]*ProtobufGen.TBOSSDUNGEON)
	for _, items := range dataList.GetItems() {
		gdTBossDungeonMap[items.GetTeamTypeID()] = items
	}
}

func loadTBBossHeroType(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.TBOSSHEROTYPE_ARRAY{}
	error := proto.Unmarshal(buffer, dataList)
	panicIfErr(error)
	gdTBossHeroTypeMap = make(map[uint32]*ProtobufGen.TBOSSHEROTYPE)
	for _, items := range dataList.GetItems() {
		gdTBossHeroTypeMap[items.GetDateID()] = items
	}
}

func loadTBBossEnemy(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.TBOSSENEMY_ARRAY{}
	error := proto.Unmarshal(buffer, dataList)
	panicIfErr(error)
	gdTBossEnemy = dataList.GetItems()
}

func loadTBBoxLoot(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.TBOSSBOXLOOT_ARRAY{}
	error := proto.Unmarshal(buffer, dataList)
	panicIfErr(error)
	gdTBossBoxLoot = dataList.GetItems()
}

func loadTBBossVipCtrl(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.TBOSSVIPCONTROL_ARRAY{}
	error := proto.Unmarshal(buffer, dataList)
	panicIfErr(error)
	gdTBossVipCtrl = dataList.GetItems()
}

func loadTRoopsMessage(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.TROOPSMESSAGE_ARRAY{}
	error := proto.Unmarshal(buffer, dataList)
	panicIfErr(error)
	gdTRoopsMessage = dataList.GetItems()
}

func GetTBossDiffMap() map[uint32]*ProtobufGen.TBOSSMAINDATA {
	return gdTBossDiffMap
}

func GetTBossDungeonMap() map[uint32]*ProtobufGen.TBOSSDUNGEON {
	return gdTBossDungeonMap
}

func GetTBossHeroTypeMap() map[uint32]*ProtobufGen.TBOSSHEROTYPE {
	return gdTBossHeroTypeMap
}

func GetTBossMainDataByDiff(diff uint32) *ProtobufGen.TBOSSMAINDATA {
	return gdTBossDiffMap[diff]
}

func GetHighestLevelByCorpLv(corpLv uint32) uint32 {
	var diff uint32
	for k, v := range gdTBossDiffMap {
		if corpLv >= v.GetPlayerNeedLevel() && diff < k {
			diff = k
		}
	}
	return diff
}

func GetTBossVipCtrl(vipLv uint32) *ProtobufGen.TBOSSVIPCONTROL {
	for _, item := range gdTBossVipCtrl {
		if item.GetVIPLower() <= vipLv && vipLv <= item.GetVIPUpper() {
			return item
		}
	}
	return nil
}

//计算花钻开宝箱需要花费的钻石数
func CalTBOpenCost(leftTime int64) int {
	minTime := BoxCfg.BoxTimeMin
	minHC := BoxCfg.TimeMinHc
	if minTime <= 0 {
		logs.Error("invalid teamboss box min time %d", minTime)
	}
	if minHC <= 0 {
		logs.Error("invalid teamboss box min HC %d", minHC)
	}
	count := int(math.Ceil(float64(leftTime) / float64(minTime)))
	return count * minHC
}

func GetTBBossData(id string) *ProtobufGen.TBOSSENEMY {
	for _, item := range gdTBossEnemy {
		if item.GetWBossID() == id {
			logs.Debug("tb boss data item: %v", item.GetWBossID())
			return item
		}
	}
	return nil
}

func GetTBEnemyLevelData(lv uint32) *ProtobufGen.TBOSSMAINDATA {
	for _, item := range gdTBossDiffMap {
		if item.GetBossLevel() == lv {
			return item
		}
	}
	return nil
}

func GetTBEnemyData(id string) *ProtobufGen.TROOPSMESSAGE {
	for _, item := range gdTRoopsMessage {
		if item.GetLevelID() == id {
			return item
		}
	}
	return nil
}

func RandomTBBox(dropId uint32) string {
	for _, item := range gdTBossBoxLoot {
		if item.GetBoxDropGroup() == dropId {
			randIndex := util.RandomItem(TBossForRand(*item))
			return item.GetLoot_Table()[randIndex].GetBoxID()
		}
	}
	return ""
}

//随机一个红宝箱id
func GetRedBoxId(diff uint32) string {
	for _, item := range gdItems {
		if item.GetRareLevel() == RareLv_Red {
			for _, boxId := range gdBoxByDiff[diff] {
				if boxId == item.GetID() {
					return boxId
				}
			}
		}
	}
	return ""
}

//判断一个宝箱是否是红色以上
func IsRedOrGoldenBox(boxId string) bool {
	for _, item := range gdItems {
		if item.GetRareLevel() >= RareLv_Gold {
			for _, box := range gdTBossBoxData {
				if boxId == item.GetID() && boxId == box.GetID() {
					return true
				}
			}
		}
	}
	return false
}

func GetTBBoxDataByBoxId(boxId string) *ProtobufGen.TBOSSBOXDATA {
	for _, item := range gdTBossBoxData {
		if item.GetID() == boxId {
			return item
		}
	}
	return nil
}

func GetTBBoxNeedTime(boxId string) int64 {
	for _, item := range gdTBossBoxData {
		if item.GetID() == boxId {
			return int64(item.GetOpenNeedTime())
		}
	}
	return -1
}

func GetTBBoxLootTableByBoxID(boxId string) []*ProtobufGen.TBOSSBOXDATA_LootRule {
	for _, item := range gdTBossBoxData {
		if item.GetID() == boxId {
			return item.Loot_Table
		}
	}
	return nil
}

//随机一个组合ID
func GetTBTeamTypeID() uint32 {
	date := uint32(time.Now().Weekday())
	for day, item := range gdTBossHeroTypeMap {
		if date == uint32(time.Sunday) && day == 7 {
			randIndex := util.RandomItem(TBossTypeForRand(*item))
			return item.GetType_Table()[randIndex].GetTeamTypeID()
		} else if date == day {
			randIndex := util.RandomItem(TBossTypeForRand(*item))
			return item.GetType_Table()[randIndex].GetTeamTypeID()
		}
	}
	return 0
}

//随机一个bossID
func GetTBBossID(dayTypeID uint32) string {
	for typeId, item := range gdTBossDungeonMap {
		if typeId == dayTypeID {
			randIndex := util.RandomItem(TBossIdAndSenceForRand(*item))
			return item.GetChoose_Table()[randIndex].GetWbossID()
		}
	}
	return ""
}

//随机一个场景id
func GetTBSceneID(dayTypeID uint32) string {
	for typeId, item := range gdTBossDungeonMap {
		if typeId == dayTypeID {
			randIndex := util.RandomItem(TBossIdAndSenceForRand(*item))
			return item.GetChoose_Table()[randIndex].GetLevelInfoID()
		}
	}
	return ""
}
