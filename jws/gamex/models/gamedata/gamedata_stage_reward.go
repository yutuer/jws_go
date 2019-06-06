package gamedata

import (
	"strings"

	"sort"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	LEVEL_TYPE_TEST  = iota
	LEVEL_TYPE_MAIN  //主线关卡
	LEVEL_TYPE_ELITE // 精英关卡
	LEVEL_TYPE_GOLDLEVEL
	LEVEL_TYPE_BOSS
	LEVEL_TYPE_EXPLEVEL
	LEVEL_TYPE_PVP
	LEVEL_TYPE_GATEENEMY //兵临城下
	LEVEL_TYPE_DCLEVEL
	LEVEL_TYPE_TRIAL
	LEVEL_TYPE_TEAMBOSS  // RAID
	LEVEL_TYPE_MINILEVEL //主线关卡
	LEVEL_TYPE_Gank
	LEVEL_TYPE_HELL               // 地狱关卡
	LEVEL_TYPE_GUILDBOSS          //军团bos
	LEVEL_TYPE_FENGHUOSINGLE      //组队pve单人模式
	LEVEL_TYPE_FENGHUO            //组队pve
	LEVEL_TYPE_TEAMPVP            //比武场
	LEVEL_TYPE_TESTHERO           //名将体验
	LEVEL_TYPE_EXPEDITION         //远征
	LEVEL_TYPE_GVG                // 军团战
	LEVEL_TYPE_FESTIVAL           // 节日
	LEVEL_TYPE_HERODIFF_TU        //出奇制胜武将差异化
	LEVEL_TYPE_HERODIFF_ZHAN      //出奇制胜武将差异化
	LEVEL_TYPE_HERODIFF_HU        //出奇制胜武将差异化
	LEVEL_TYPE_HERODIFF_SHI       //出奇制胜武将差异化
	LEVEL_TYPE_WORLD_BOSS    = 28 //世界boss
	LEVEL_TYPE_TEAM_BOSS     = 29 //组队boss
	LEVEL_TYPE_Count         = 30
)

type stageRewardLimit struct {
	Item_group_id  string
	Num            int32
	Space          int32
	Offset         int32
	MItem_group_id string
}

type StageData struct {
	Id                string   // * 关卡ID
	Chapter           string   // * 章节
	Type              int32    // * 0:测试，1:主线，2:精英，3:金币关，4:BOSS战
	Energy            int32    // * 体力消耗
	HighEnergy        int32    // * 包子消耗
	MaxDailyAccess    int32    // * 每日最大访问
	TimeLimit         int32    // * 限时
	TimeGoal          int32    // * 加星时间
	HpGoal            float32  // * 加星生命比例
	PreLevelID        []string // * 前续关卡ID列表
	CorpLvRequirement int32    // * 队伍等级限制
	LevelRequirement  int32    // * 角色等级限制
	RoleOnly          int32    // * 角色专属
	GameModeId        uint32   // * 活动id外键
	LevelIndex        int32    // * 普通关卡序号
	EliteLevelIndex   int32    // * 精英关卡序号
	HellLevelIndex    int32    // * 地狱关卡序号
}

type stageRewardData struct {
	SCReward     int64
	XpReward     uint32
	CorpXpReward uint32

	ManualXpReward uint32
	SweepItem      string
	SweepCount     uint32

	FirstLootAddonItem  string
	FirstLootAddonCount uint32
}

type stageRandRewardData struct {
	Reward []string
	P      []uint32 // idx项的发奖概率 万分比
}

var (
	gdStageRewardLimitConfig      map[string][]stageRewardLimit
	gdFirstStageRewardLimitConfig map[string][]stageRewardLimit
	gdStageRewardRandConfig       map[string]stageRandRewardData
	gdStageRewardConfig           map[string]stageRewardData
	gdFirstStageRewardConfig      map[string]stageRewardData
	gdStageData                   map[string]StageData
	gdStagePurchasePolicy         map[string]string
	gdStageTimeLimit              map[int32]int32
)

func GetAllStageData() map[string]StageData {
	return gdStageData
}

func GetStageRewardLimitCfg(stage_id string, isFirst bool) []stageRewardLimit {
	cfg, ok := gdStageRewardLimitConfig[stage_id]
	if isFirst {
		cfgFirst, okFirst := gdFirstStageRewardLimitConfig[stage_id]
		if okFirst {
			cfg = cfgFirst
		}
	}

	if !ok {
		return []stageRewardLimit{}[:]
	}

	return cfg[:]
}

func GetStageRewardRandCfg(stage_id string) *stageRandRewardData {
	cfg, ok := gdStageRewardRandConfig[stage_id]
	if !ok {
		return nil
	}

	return &cfg
}

func GetStageData(stage_id string) *StageData {
	cfg, ok := gdStageData[stage_id]
	if !ok {
		return nil
	}
	return &cfg
}

func GetStage2Chapter(stage_id string) string {
	stage := GetStageData(stage_id)
	if stage == nil {
		return ""
	}
	return stage.Chapter
}

// 掉落软通和经验的表配置
func GetStageReward(stage_id string, isFirst bool) *stageRewardData {
	cfg, ok := gdStageRewardConfig[stage_id]
	if isFirst {
		cfgFirst, okFirst := gdFirstStageRewardConfig[stage_id]
		if okFirst {
			cfg = cfgFirst
		}
	}
	if ok {
		return &cfg
	} else {
		return nil
	}
}

func loadStageRewardLimit(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	// 读取武器装备表
	buffer, err := loadBin(filepath)
	errcheck(err)

	stageRewards := &ProtobufGen.STAGELIMITREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, stageRewards)
	errcheck(err)

	rewards := stageRewards.GetItems()
	gdStageRewardLimitConfig = make(map[string][]stageRewardLimit)
	gdStageRewardConfig = make(map[string]stageRewardData)
	for _, r := range rewards {
		limit_table := r.GetSRewardLimit_Table()
		limits := make([]stageRewardLimit, 0, len(limit_table))
		for _, a := range limit_table {
			limits = append(limits,
				stageRewardLimit{
					a.GetItemGroupID(),
					a.GetLootNum(),
					a.GetLootSpace(),
					a.GetOffset(),
					a.GetMItemGroupID()})
		}
		gdStageRewardLimitConfig[r.GetID()] = limits
		gdStageRewardConfig[r.GetID()] = stageRewardData{
			int64(r.GetSC()),
			r.GetXP(),
			r.GetCorpXP(),
			r.GetManualXP(),
			r.GetSweepItem(),
			r.GetSweepCount(),
			r.GetFirstLootItem(),
			r.GetFirstLootCount(),
		}
	}
}

func loadFirstStageRewardLimit(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	// 读取武器装备表
	buffer, err := loadBin(filepath)
	errcheck(err)

	stageRewards := &ProtobufGen.FIRSTSTAGELIMITREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, stageRewards)
	errcheck(err)

	rewards := stageRewards.GetItems()
	gdFirstStageRewardLimitConfig = make(map[string][]stageRewardLimit)
	gdFirstStageRewardConfig = make(map[string]stageRewardData)
	for _, r := range rewards {
		limit_table := r.GetSRewardLimit_Table()
		limits := make([]stageRewardLimit, 0, len(limit_table))
		for _, a := range limit_table {
			limits = append(limits,
				stageRewardLimit{
					a.GetItemGroupID(),
					a.GetLootNum(),
					a.GetLootSpace(),
					a.GetOffset(),
					a.GetMItemGroupID()})
		}
		if len(limits) > 0 {
			gdFirstStageRewardLimitConfig[r.GetID()] = limits
		}

		if r.GetSC() > 0 || r.GetXP() > 0 || r.GetCorpXP() > 0 || r.GetManualXP() > 0 {
			gdFirstStageRewardConfig[r.GetID()] = stageRewardData{
				SCReward:       int64(r.GetSC()),
				XpReward:       r.GetXP(),
				CorpXpReward:   r.GetCorpXP(),
				ManualXpReward: r.GetManualXP(),
			}
		}
	}
}

func loadStageRewardRand(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	// 读取武器装备表
	buffer, err := loadBin(filepath)
	errcheck(err)

	stageRewards := &ProtobufGen.STAGERANDREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, stageRewards)
	errcheck(err)

	rewards := stageRewards.GetItems()
	gdStageRewardRandConfig = make(map[string]stageRandRewardData)
	for _, r := range rewards {
		rand_table := r.GetSRewardRand_Table()

		rands := stageRandRewardData{}
		rands.Reward = make([]string, 0, len(rand_table))
		rands.P = make([]uint32, 0, len(rand_table))

		for _, a := range rand_table {
			rands.Reward = append(rands.Reward, a.GetItemGroupID())
			rands.P = append(rands.P, a.GetSRandRate())
		}
		gdStageRewardRandConfig[r.GetStageID()] = rands
	}
}

func loadStageData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	buffer, err := loadBin(filepath)
	errcheck(err)

	stage_data := &ProtobufGen.LEVEL_INFO_ARRAY{}
	err = proto.Unmarshal(buffer, stage_data)
	errcheck(err)

	datas := stage_data.GetItems()
	gdStageData = make(map[string]StageData, len(datas))
	gdStagePurchasePolicy = make(map[string]string)
	for _, r := range datas {
		pre_lv := strings.Split(r.GetPreLevelID(), ",")
		gdStageData[r.GetLevelID()] = StageData{
			r.GetLevelID(),
			r.GetChapterID(),
			r.GetLevelType(),
			r.GetEnergy(),
			r.GetHighEnergy(),
			r.GetMaxDailyAccess(),
			r.GetTimeLimit(),
			r.GetTimeGoal(),
			r.GetHpGoal(),
			pre_lv,
			r.GetTeamRequirement(),
			r.GetLevelRequirement(),
			r.GetRoleOnly(),
			uint32(r.GetGameModeID()),
			r.GetLevelIndex(),
			r.GetEliteLevelIndex(),
			r.GetHardLevelIndex(),
		}
		if r.GetLevelPurchasePolicy() != "" {
			gdStagePurchasePolicy[r.GetLevelID()] = r.GetLevelPurchasePolicy()
		}
	}
	gdStageTimeLimit = make(map[int32]int32, 0)
	for _, r := range datas {
		gdStageTimeLimit[r.GetLevelType()] = r.GetTimeLimit()
	}
}

func GetStageTimeLimit(typ int32) int32 {
	return gdStageTimeLimit[typ]
}

var gdAvatarRanders [AVATAR_NUM_CURR + 1]util.RandIntSet // 上场人数对应的随机池缓冲
var gdAvatarInStagePower uint32
var gdAvatarNotInStagePower uint32
var gdAvatarInStagePowerAdden uint32 // gdAvatarInStagePower - gdAvatarNotInStagePower

func loadStageAvatarPower(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	buffer, err := loadBin(filepath)
	errcheck(err)

	data := &ProtobufGen.ROLEWEIGHT_ARRAY{}
	err = proto.Unmarshal(buffer, data)
	errcheck(err)

	datas := data.GetItems()
	if len(datas) < 1 {
		logs.Error("loadStageAvatarPower Err Data Miss")
	}

	gdAvatarInStagePower = datas[0].GetRoleFightWeight()
	gdAvatarNotInStagePower = datas[0].GetRoleRestWeight()
	gdAvatarInStagePowerAdden = gdAvatarInStagePower - gdAvatarNotInStagePower

	// 注意这个是权重
	for i := 0; i < len(gdAvatarRanders); i++ {
		gdAvatarRanders[i].Init(AVATAR_NUM_MAX)
	}

	for i := 0; i < len(gdAvatarRanders); i++ {
		// 先按照小概率把所有人加进去， 等到真正用的时候，再按照高概率与低概率之差追加上场的武将
		for avatar_id := 0; avatar_id < AVATAR_NUM_CURR; avatar_id++ {
			ok := gdAvatarRanders[i].Add(avatar_id, gdAvatarNotInStagePower)
			if !ok {
				logs.Error("loadStageAvatarPower Err Add Rander %v", gdAvatarRanders)
			}
		}
	}
}

func GetStageAvatarSelectRander(avatar_in_stage_num int) (bool, uint32, util.RandIntSet) {
	if avatar_in_stage_num < 0 || avatar_in_stage_num >= len(gdAvatarRanders) {
		return false, 0, util.RandIntSet{}
	}
	return true, gdAvatarInStagePowerAdden, gdAvatarRanders[avatar_in_stage_num]
}

var (
	gdChapterAward    map[string]map[uint32]*ProtobufGen.CHAPTERAWARD // chapterId -> goalNum -> chapterAward
	gdChapterIdx2Star map[string][]int
)

func loadChapterReward(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	buffer, err := loadBin(filepath)
	errcheck(err)

	data := &ProtobufGen.CHAPTERAWARD_ARRAY{}
	err = proto.Unmarshal(buffer, data)
	errcheck(err)

	gdChapterAward = map[string]map[uint32]*ProtobufGen.CHAPTERAWARD{}
	gdChapterIdx2Star = map[string][]int{}
	for _, ch := range data.GetItems() {
		info, ok := gdChapterAward[ch.GetChapterID()]
		if !ok {
			info = map[uint32]*ProtobufGen.CHAPTERAWARD{}
			gdChapterAward[ch.GetChapterID()] = info
		}
		info[ch.GetGoal()] = ch

		idx2Star, ok := gdChapterIdx2Star[ch.GetChapterID()]
		if !ok {
			idx2Star = make([]int, 0, 3)
			gdChapterIdx2Star[ch.GetChapterID()] = idx2Star
		}
		idx2Star = gdChapterIdx2Star[ch.GetChapterID()]
		idx2Star = append(idx2Star, int(ch.GetGoal()))
		gdChapterIdx2Star[ch.GetChapterID()] = idx2Star
	}
	ks := make([]string, 0, len(gdChapterIdx2Star))
	for k, _ := range gdChapterIdx2Star {
		ks = append(ks, k)
	}
	for _, k := range ks {
		v := gdChapterIdx2Star[k]
		sort.Ints(v)
		gdChapterIdx2Star[k] = v
	}
}

func ChapterAwardId2Star(chapterId string, index int32) int {
	idx2Star := gdChapterIdx2Star[chapterId]
	if idx2Star == nil || int(index) >= len(idx2Star) {
		return -1
	}
	return idx2Star[index]
}

func ChapterIsExist(chapterId string) bool {
	_, ok := gdChapterAward[chapterId]
	return ok
}

func ChapterGoalAward(chapterId string, goal uint32) *ProtobufGen.CHAPTERAWARD {
	chapter, ok := gdChapterAward[chapterId]
	if !ok {
		return nil
	}

	award, ok := chapter[goal]
	if !ok {
		return nil
	}
	return award
}

func StagesPurchasePolicy() map[string]string {
	return gdStagePurchasePolicy
}
