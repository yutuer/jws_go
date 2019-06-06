package gamedata

import (
	"errors"
	"strconv"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdConds               map[uint32]*ProtobufGen.CONDITION
	gdSimplePvpStartStage string
	gdSimplePvpStartLv    int
)

const (
	Mod_ActiveGift    = 0  // 七日登陆
	Mod_SCGacha       = 1  // 金币宝箱
	Mod_Arousal       = 2  // 主将觉醒
	Mod_HCGacha       = 3  // 钻石宝箱
	Mod_GoldLevel     = 4  // 金币关
	Mod_EquipResolve  = 5  // 装备熔炼
	Mod_PevBoss       = 6  // 名将乱入
	Mod_Rank          = 7  // 排行榜
	Mod_DailyTask     = 8  // 日常任务
	Mod_Store         = 9  // 商人
	Mod_EquipAbstract = 10 // 洗炼
	Mod_ExpLevel      = 11 // 精铁关
	Mod_SimplePvp     = 13 // Pvp
	Mod_DCLevel       = 34 // 天命关
	Mod_Trial         = 37 // 爬塔开启条件
	Mod_TeamPvp       = 40 // 组队pvp
	Mod_GrowFund      = 46 // 成长基金
	Mod_EatBaozi      = 55 // 吃包子
	Mod_Expedition    = 64 // 远征
	Mod_WhiteGacha    = 73 // 白盒宝箱
	Mod_MagicPet      = 84 // 灵宠
	Mod_LuckyWheel 	  = 86 //幸运转盘
	Mod_BattleArmy	  = 88 // 战阵系统
)

func loadConditionConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.CONDITION_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	gdConds = map[uint32]*ProtobufGen.CONDITION{}

	for _, cond := range lv_ar.GetItems() {
		gdConds[cond.GetConditionID()] = cond
	}

	// pvp
	pvpData, ok := gdConds[Mod_SimplePvp]
	if !ok || pvpData.GetConditionValue() == "" {
		panic(errors.New("pvp Condition Err"))
		return
	}
	if pvpData.GetConditionType() == 0 {
		gdSimplePvpStartLv, err = strconv.Atoi(pvpData.GetConditionValue())
		if err != nil {
			panic(err)
		}
	} else if pvpData.GetConditionType() == 1 {
		gdSimplePvpStartStage = pvpData.GetConditionValue()
	} else {
		panic(errors.New("pvp GetConditionType Err"))
	}

}

func GetCond(modeId uint32) *ProtobufGen.CONDITION {
	return gdConds[modeId]
}

func GetSimplePvpStartStage() string {
	return gdSimplePvpStartStage
}

func GetSimplePvpStartLv() int {
	return gdSimplePvpStartLv
}
