package gamedata

import (
	"fmt"
	"strconv"
	"strings"

	"sort"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	GoldLevel_ModeId = 1
	ExpLevel_ModeId  = 2
)

type validWeekDayCouple struct {
	beginTimeSlot int64
	endTimeSlot   int64
}

type gameModeInfo struct {
	id            uint32
	validWeekTime []*validWeekDayCouple
	info          *ProtobufGen.MODECONTROL
}

var (
	gdGameModeConfig map[uint32]*gameModeInfo
	gdGameModeIds    []uint32
)

func loadModeControlConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.MODECONTROL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGameModeConfig = map[uint32]*gameModeInfo{}
	gdGameModeIds = []uint32{}
	for _, mc := range dataList.GetItems() {
		info := gameModeInfo{
			mc.GetModeID(),
			[]*validWeekDayCouple{},
			mc,
		}
		for _, cfgWeekday := range strings.Split(mc.GetGetTicketDay(), ",") {
			cwd, err := strconv.Atoi(cfgWeekday)
			if err != nil || cwd < 1 || cwd > 7 {
				logs.Error("loadModeControlConfig modeId %v cfgWeekDay err %v", mc.GetModeID(), cfgWeekday)
				continue
			}
			b := util.WeeklyTimeFromString(cwd, mc.GetGetTicketTime(), mc.GetGetTicketTime())
			info.validWeekTime = append(info.validWeekTime, &validWeekDayCouple{b, b + 48})
		}
		gdGameModeConfig[mc.GetModeID()] = &info
		gdGameModeIds = append(gdGameModeIds, mc.GetModeID())
	}
}

func GetNowValidGameModeIndex(id uint32, u int64) (bool, int64, error) {
	info, ok := gdGameModeConfig[id]
	if !ok {
		return false, -1, fmt.Errorf("gamedata GetNowValidGameModeIndex id %v not found", id)
	}
	curSlot := util.WeekTime(u, info.info.GetGetTicketTime())
	for _, couple := range info.validWeekTime {
		if curSlot >= couple.beginTimeSlot && curSlot < couple.endTimeSlot {
			return true, util.WeeklySlot2Unix(couple.beginTimeSlot, info.info.GetGetTicketTime(), u), nil
		}
	}
	return false, -1, nil
}

func GetGameModeCfg(id uint32) *ProtobufGen.MODECONTROL {
	if cfg, ok := gdGameModeConfig[id]; ok {
		return cfg.info
	}
	return nil
}

func GetGameMode2CondModId(id uint32) (uint32, bool) {
	switch id {
	case CounterTypeGoldLevel:
		return Mod_GoldLevel, true
	case CounterTypeFineIronLevel:
		return Mod_ExpLevel, true
	case CounterTypeDCLevel:
		return Mod_DCLevel, true
	case CounterTypeTeamPvp:
		return Mod_TeamPvp, true
	case CounterTypeBoss:
		return Mod_PevBoss, true
	case CounterTypeEatBaozi:
		return Mod_EatBaozi, true
	default:
		logs.Error("GameMode2Condition not found gamemode id : %d", id)
		return 0, false
	}
}

func SweepVipVaild(gameModeId, vip uint32) bool {
	vipCfg := GetVIPCfg(int(vip))
	if vipCfg != nil {
		switch gameModeId {
		case CounterTypeBoss:
			return vipCfg.BossFightSweep
		case CounterTypeGoldLevel:
			return vipCfg.GoldLevelSweep
		case CounterTypeFineIronLevel:
			return vipCfg.IronLevelSweep
		case CounterTypeDCLevel:
			return vipCfg.DcLevelSweep
		}
	}
	return true
}

func GetGameModeCfgTimes(id uint32) (uint32, error) {
	info, ok := gdGameModeConfig[id]
	if !ok {
		return 0, fmt.Errorf("gamedata GetGameModeCfgTimes id %v not found", id)
	}
	return *info.info.GetTicketNumber, nil
}

func GetGameModeCDSec(id uint32) (int64, uint32, error) {
	info, ok := gdGameModeConfig[id]
	if !ok {
		return 0, 0, fmt.Errorf("gamedata GetGameModeCDSec id %v not found", id)
	}
	if info.info.GetTicketCDM() <= 0 {
		return 0, 0, nil
	}
	return int64(info.info.GetTicketCDM()) * 60, info.info.GetCostTicketCD(), nil
}

func GetAllGameModeId() []uint32 {
	return gdGameModeIds
}

type goldLevelKey struct {
	modeId     uint32
	difficulty uint32
	minReward  uint32
	cfg        *ProtobufGen.GOLDLEVEL
}

type goldLevelList []goldLevelKey

func (pq goldLevelList) Len() int      { return len(pq) }
func (pq goldLevelList) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq goldLevelList) Less(i, j int) bool {
	return pq[i].difficulty > pq[j].difficulty
}

var (
	gdLevel2KeyConfig          map[string]goldLevelKey
	gdGoldLevelOrderByDiffDesc []string
)

func loadGoldLevelConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GOLDLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdLevel2KeyConfig = make(map[string]goldLevelKey, len(dataList.GetItems()))
	gdGoldLevelOrderByDiffDesc = make([]string, 0, len(dataList.GetItems()))
	_l := make(goldLevelList, 0, len(dataList.GetItems()))
	for _, gl := range dataList.GetItems() {
		key := goldLevelKey{*gl.ModeID, *gl.Difficulty, gl.GetReward(), gl}
		gdLevel2KeyConfig[*gl.Level] = key
		_l = append(_l, key)
	}
	sort.Sort(_l)
	for _, i := range _l {
		gdGoldLevelOrderByDiffDesc = append(gdGoldLevelOrderByDiffDesc, i.cfg.GetLevel())
	}
}

func IsGoldLevel(levelId string) (ok bool) {
	if _, ok = gdLevel2KeyConfig[levelId]; !ok {
		return false
	}
	return true
}

func GetGoldLevelCfg(levelId string) *ProtobufGen.GOLDLEVEL {
	if c, ok := gdLevel2KeyConfig[levelId]; ok {
		return c.cfg
	}
	return nil
}

func GetGoldLevelMinReward(levelId string) uint32 {
	r, ok := gdLevel2KeyConfig[levelId]
	if !ok {
		return 0
	}
	return r.minReward
}

func GoldLevelCorpLvl2MaxTotalReward(corpLvl uint32) (uint32, uint32) {
	var maxDiff uint32
	var totalReward uint32
	for lvl, v := range gdLevel2KeyConfig {
		if v.difficulty > maxDiff {
			stage_data := GetStageData(lvl)
			if corpLvl >= uint32(stage_data.CorpLvRequirement) {
				maxDiff = v.difficulty
				totalReward = v.cfg.GetTotal()
			}
		}
	}
	return totalReward, maxDiff
}

func GoldLevelOrderByDiffDesc() []string {
	return gdGoldLevelOrderByDiffDesc
}

var (
	gdExpLevel2Info           map[string]*ProtobufGen.EXPLEVEL
	gdExpLevelOrderByDiffDesc []string
)

type expLevelList []*ProtobufGen.EXPLEVEL

func (pq expLevelList) Len() int      { return len(pq) }
func (pq expLevelList) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq expLevelList) Less(i, j int) bool {
	return pq[i].GetDifficulty() > pq[j].GetDifficulty()
}

func loadExpLevelCfg(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.EXPLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdExpLevel2Info = make(map[string]*ProtobufGen.EXPLEVEL, len(dataList.GetItems()))
	gdExpLevelOrderByDiffDesc = make([]string, 0, len(dataList.GetItems()))
	_l := make(expLevelList, 0, len(dataList.GetItems()))
	for _, el := range dataList.GetItems() {
		gdExpLevel2Info[el.GetLevel()] = el
		_l = append(_l, el)
	}
	sort.Sort(_l)
	for _, i := range _l {
		gdExpLevelOrderByDiffDesc = append(gdExpLevelOrderByDiffDesc, i.GetLevel())
	}
}

type ExpLevelAward struct {
	ItemID string
	Count  uint32
}

func GetExpLevelMinAward(stageId string) (uint32, bool) {
	if el, ok := gdExpLevel2Info[stageId]; ok {
		return el.GetNumber(), true
	}
	return 0, false
}

func ExpLevelCorpLvl2MaxTotalReward(corpLvl uint32) (uint32, uint32) {
	var maxDiff uint32
	var totalReward uint32
	for lvl, v := range gdExpLevel2Info {
		if v.GetDifficulty() > maxDiff {
			stage_data := GetStageData(lvl)
			if corpLvl >= uint32(stage_data.CorpLvRequirement) {
				maxDiff = v.GetDifficulty()
				totalReward = v.GetTotal()
			}
		}
	}
	return totalReward, maxDiff
}

func ExpLevelOrderByDiffDesc() []string {
	return gdExpLevelOrderByDiffDesc
}

func GetExpLevelCfg(lvlId string) *ProtobufGen.EXPLEVEL {
	return gdExpLevel2Info[lvlId]
}

var (
	gdDCLevel2CfgInfo        map[string]*ProtobufGen.DCLEVEL
	gdDCLevelOrderByDiffDesc []string
)

type dcLevelList []*ProtobufGen.DCLEVEL

func (pq dcLevelList) Len() int      { return len(pq) }
func (pq dcLevelList) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq dcLevelList) Less(i, j int) bool {
	return pq[i].GetDifficulty() > pq[j].GetDifficulty()
}

func loadDCLevelCfg(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.DCLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdDCLevel2CfgInfo = make(map[string]*ProtobufGen.DCLEVEL, len(dataList.GetItems()))
	_l := make(dcLevelList, 0, len(dataList.GetItems()))
	gdDCLevelOrderByDiffDesc = make([]string, 0, len(dataList.GetItems()))
	for _, dcl := range dataList.GetItems() {
		gdDCLevel2CfgInfo[dcl.GetLevel()] = dcl
		_l = append(_l, dcl)
	}
	sort.Sort(_l)
	for _, i := range _l {
		gdDCLevelOrderByDiffDesc = append(gdDCLevelOrderByDiffDesc, i.GetLevel())
	}
}

func GetDCLevelMinAward(stageId string) (uint32, bool) {
	if dc, ok := gdDCLevel2CfgInfo[stageId]; ok {
		return dc.GetNumber(), true
	}
	return 0, false
}

func DCLevelCorpLvl2MaxTotalReward(corpLvl uint32) (uint32, uint32) {
	var maxDiff uint32
	var totalReward uint32
	for lvl, v := range gdDCLevel2CfgInfo {
		if v.GetDifficulty() > maxDiff {
			stage_data := GetStageData(lvl)
			if corpLvl >= uint32(stage_data.CorpLvRequirement) {
				maxDiff = v.GetDifficulty()
				totalReward = v.GetTotal()
			}
		}
	}
	return totalReward, maxDiff
}

func DCLevelOrderByDiffDesc() []string {
	return gdDCLevelOrderByDiffDesc
}

func GetDCLevelCfg(lvlId string) *ProtobufGen.DCLEVEL {
	return gdDCLevel2CfgInfo[lvlId]
}
