package account

import (
	"fmt"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"reflect"

	"math"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/anticheatlog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 客户端参数index
const (
	LEVEL_START_TIME     = iota // 关卡开始时间
	LEVEL_END_TIME              // 关卡结束时间
	ATTACK_MAX_POINT_FD         // 最大单招伤害
	DEFENCE_MAX_POINT_FD        // 最大防御力
	ENERGY_CURRENT              // 最后结算时候的MP
	ENERGY_GAIN_TOTAL           // 获取的MP
	ENERGY_COST                 // 消耗的MP
	ENERGY_GAIN_BY_OTHER        // 其他途径获取的MP
	ENERGY_INIT                 // 初始的的MP
	SKILL1_CD_MINTIME           // GY1 ZF1 SX1 / PVP: GY1 ZF1 SX1技能施放最小CD
	SKILL2_CD_MINTIME           // GY9 ZF2 SX9 / PVP: GY9 ZF2 SX9技能施放最小CD
	SKILL3_CD_MINTIME           // GY2 ZF3 SX2 / PVP: GY2 ZF9 SX2技能施放最小CD
	SKILL4_CD_MINTIME           // GY3 ZF9 SX3技能施放最小CD
	SKILL5_CD_MINTIME           // GY4 ZF4 SX4技能施放最小CD
	ENERGY_MAX                  // 战斗中的MP上限
	HP_MAX                      // 战斗中的HP上限
	Pass_Time                   // 通关时间
	Hero1_id                    // 英雄1id
	Hero1_gs                    // 英雄1gs
	Hero2_id                    // 英雄2id
	Hero2_gs                    // 英雄2gs
	Hero3_id                    // 英雄3id
	Hero3_gs                    // 英雄3gs
	Hero4_id                    // 英雄4id
	Hero4_gs                    // 英雄4gs
	Hero5_id                    // 英雄5id
	Hero5_gs                    // 英雄5gs
	Hero6_id                    // 英雄6id
	Hero6_gs                    // 英雄6gs
	Hero7_id                    // 英雄7id
	Hero7_gs                    // 英雄7gs
	Hero8_id                    // 英雄8id
	Hero8_gs                    // 英雄8gs
	Hero9_id                    // 英雄9id
	Hero9_gs                    // 英雄9gs

	AC_Count // 反作弊参数数量
)

// 检查条目index
const (
	CheckerMaxDamage     = iota // 0 最大单招伤害
	CheckerMaxDefence           // 1 最大防御力
	CheckerMPCostGet            // 2 MP消耗和获取
	CheckerLevelTime            // 3 通关时间
	CheckerSkillCD              // 4 技能CD
	CheckerHPMPRageLimit        // 5 HP/MP上限
	CheckerDamageResist
	CheckerHPMPRageCostGet
	CheckerHeroGS   // 8 gs检查
	CheckerPassTime // 9 通关时间检查
	CheckerCount
)

const MAX_CHECK_AVATAR = 9

const (
	Anticheat_Typ_SimplePVP    = "SimplePVP"
	Anticheat_Typ_TeamPVP      = "TeamPVP"
	Anticheat_Typ_Expedition   = "Expedition"
	Anticheat_Typ_GuildBoss    = "GuildBoss"
	Anticheat_Typ_BossFight    = "BossFight"
	Anticheat_Typ_GVG          = "GVG"
	Anticheat_Typ_LevelStage   = "LevelStage"
	Anticheat_Typ_Trial        = "Trial"
	Anticheat_Typ_FestivalBoss = "FestivalBoss"
	Anticheat_Typ_HeroDiff     = "HeroDiff"
	Anticheat_Typ_Wspvp        = "Wspvp"
	Anticheat_Typ_CSRob        = "CSRob"
	Anticheat_Typ_WB           = "WorldBoss"
	Anticheat_Typ_GateEnemy    = "GateEnemy"
)

type checker func(param anticheat_param) (isOk bool)

var (
	checkers     [CheckerCount]checker
	checker_strs = []string{
		CheckerMaxDamage:       "MaxDamage",
		CheckerMaxDefence:      "MaxDefence",
		CheckerMPCostGet:       "MPCostGet",
		CheckerLevelTime:       "LevelTime",
		CheckerHPMPRageLimit:   "HPMPRageLimit",
		CheckerSkillCD:         "SkillCD",
		CheckerDamageResist:    "DamageResist",
		CheckerHPMPRageCostGet: "HPMPRageCostGet",
		CheckerHeroGS:          "CheckerHeroGS",
		CheckerPassTime:        "CheckerPassTime",
	}
	hero_idx_id = map[int]int{0: Hero1_id, 1: Hero2_id, 2: Hero3_id, 3: Hero4_id, 4: Hero5_id, 5: Hero6_id, 6: Hero7_id,
		7: Hero8_id, 8: Hero9_id}
	hero_idx_gs = map[int]int{0: Hero1_gs, 1: Hero2_gs, 2: Hero3_gs, 3: Hero4_gs, 4: Hero5_gs, 5: Hero6_gs, 6: Hero7_gs,
		7: Hero8_gs, 8: Hero9_gs}
)

func init() {
	checkers[CheckerMaxDamage] = checkMaxDamage
	checkers[CheckerMaxDefence] = checkMaxDefence
	checkers[CheckerMPCostGet] = checkMpCostGet
	checkers[CheckerLevelTime] = checkLevelTime
	checkers[CheckerSkillCD] = checkSkillCD
	checkers[CheckerHPMPRageLimit] = checkHPMPRageLimit
	checkers[CheckerDamageResist] = checkDamageResist
	checkers[CheckerHPMPRageCostGet] = checkHPMPRageCostGet
	checkers[CheckerHeroGS] = checkHeroGS
	checkers[CheckerPassTime] = checkPassTime
}

type PlayerAntiCheat struct {
	AntiCheat
}

func NewPlayerAntiCheat(account db.Account) PlayerAntiCheat {
	AccountID := account

	cheat := PlayerAntiCheat{
		AntiCheat: *NewAntiCheat(AccountID),
	}
	return cheat
}

type AntiCheat struct {
	dbkey        db.ProfileDBKey
	dirtiesCheck map[string]interface{}
	Ver          int64 `redis:"version"`

	Cheated [CheckerCount]bool `redis:"cheated"`
	BanTime int64              `redis:"bantime"`
}

func NewAntiCheat(account db.Account) *AntiCheat {
	re := &AntiCheat{
		dbkey: db.ProfileDBKey{
			Account: account,
			Prefix:  "anticheat",
		},
		//Ver:        helper.CurrDBVersion,
	}
	return re
}

type anticheat_param struct {
	acid      string
	info      []float32
	avatars   []int
	account   *Account
	isPvp     bool
	levelTime int64
	gs        int
	typ       string
}

type AnticheatLogInfo struct {
	Version         string  `json:"version"`
	Type            string  `json:"Type,omitempty"`
	Damage          float32 `json:"Damage,omitempty"`
	ClientLevelTime float32 `json:"ClientLevelTime,omitempty"`
	ServerLevelTime int64   `json:"ServerLevelTime,omitempty"`
	ClientParamIdx  int     `json:"ClientParamIdx,omitempty"`
	ClientCD        float32 `json:"ClientCD,omitempty"`
	ServerCD        float32 `json:"ServerCD,omitempty"`
	Avatar          int
	Avatars         []int
	ClientInfo      []float32
	Attrs           gamedata.AvatarAttr
	IsPvp           bool
	GS              int
	HeroGS          int
	CheckedHeroGS   int
	Mid             Avatar2ClientJson
	Typ             string
}

func checkMaxDamage(param anticheat_param) (isFail bool) {
	/*
		damage =（attack * 技能倍率 + 技能附加值）* 暴击伤害 * （1 + 特技加成百分比）
		pvedamage = damage * (1 + 常规加成)
		pvpdamage = damage*（1 + pvp加成）*(1 + 常规加成)
	*/
	curAttr := param.account.Profile.GetData().CorpAttrs
	var skillLevel uint32
	skills := param.account.Profile.GetAvatarSkill().GetByAvatar(param.account.Profile.CurrAvatar)
	for _, l := range skills {
		if l > skillLevel {
			skillLevel = l
		}
	}
	atSkillcfg := gamedata.GetAntiCheatSkillCfg(skillLevel)
	if atSkillcfg == nil {
		logs.Error("anticheat checkMaxDamage skill %d cfg not found", skillLevel)
		return
	}
	antiCheatCommonCfg := gamedata.GetAntiCheatCommon()
	damage := curAttr.ATK*atSkillcfg.GetDamageScale() + float32(atSkillcfg.GetExtraDamage())
	damage = damage * curAttr.CritValue * (1 + antiCheatCommonCfg.GetTrickBonus())
	if param.isPvp {
		damage = damage * (1 + antiCheatCommonCfg.GetPVPBonus()) * (1 + antiCheatCommonCfg.GetCommonBonus())
	} else {
		damage = damage * (1 + antiCheatCommonCfg.GetCommonBonus())
	}
	logs.Trace("[anticheat] checkMaxDamage client %v server %v", param.info[ATTACK_MAX_POINT_FD], damage)
	if param.info[ATTACK_MAX_POINT_FD] > damage {
		logs.Warn("[anticheat-fail] checkMaxDamage client %v server %v", param.info[ATTACK_MAX_POINT_FD], damage)
		var mid_info Avatar2ClientJson
		mid_info.FromAccount(param.account, param.account.Profile.CurrAvatar)
		anticheatlog.Trace(param.acid, checker_strs[CheckerMaxDamage],
			AnticheatLogInfo{
				Avatar:     param.account.Profile.CurrAvatar,
				ClientInfo: param.info,
				Attrs:      curAttr,
				Damage:     damage,
				IsPvp:      param.isPvp,
				GS:         param.gs,
				Mid:        mid_info,
			}, "anticheat_fail")
		return true
	}
	return
}

func checkMaxDefence(param anticheat_param) (isFail bool) {
	/*
		角色面板防御力
	*/
	curAttr := param.account.Profile.GetData().CorpAttrs
	curDef := curAttr.DEF + 10 // 公式里增加宽限值为10
	logs.Trace("[anticheat] CheckMaxDefence client %v server %v", param.info[DEFENCE_MAX_POINT_FD], curDef)
	if param.info[DEFENCE_MAX_POINT_FD] > curDef {
		logs.Warn("[anticheat-fail] CheckMaxDefence client %v server %v", param.info[DEFENCE_MAX_POINT_FD], curDef)
		var mid_info Avatar2ClientJson
		mid_info.FromAccount(param.account, param.account.Profile.CurrAvatar)
		anticheatlog.Trace(param.acid, checker_strs[CheckerMaxDefence],
			AnticheatLogInfo{
				Avatar:     param.account.Profile.CurrAvatar,
				ClientInfo: param.info,
				Attrs:      curAttr,
				IsPvp:      param.isPvp,
				GS:         param.gs,
				Mid:        mid_info,
			}, "anticheat_fail")
		return true
	}
	return
}

func checkMpCostGet(param anticheat_param) (isFail bool) {
	/*
		ENERGY_INIT= ENERGY_CURRENT - ENERGY_GAIN_TOTAL - ENERGY_GAIN_BY_OTHER - ENERGY_COST;
	*/
	logs.Trace("[anticheat] CheckMPCostGet client %v ", param.info)
	if int64(param.info[ENERGY_INIT]) != int64(param.info[ENERGY_CURRENT]-
		param.info[ENERGY_GAIN_TOTAL]-
		param.info[ENERGY_GAIN_BY_OTHER]-
		param.info[ENERGY_COST]) {
		logs.Warn("[anticheat-fail] CheckMPCostGet client %v ", param.info)
		curAttr := param.account.Profile.GetData().CorpAttrs
		var mid_info Avatar2ClientJson
		mid_info.FromAccount(param.account, param.account.Profile.CurrAvatar)
		anticheatlog.Trace(param.acid, checker_strs[CheckerMPCostGet],
			AnticheatLogInfo{
				Avatar:     param.account.Profile.CurrAvatar,
				ClientInfo: param.info,
				Attrs:      curAttr,
				IsPvp:      param.isPvp,
				GS:         param.gs,
				Mid:        mid_info,
			}, "anticheat_fail")
		return true
	}
	return
}

func checkLevelTime(param anticheat_param) (isFail bool) {
	/*
		客户端时间 > 服务器的时间 * 1.2 + 5s
	*/
	clientTime := param.info[LEVEL_END_TIME] - param.info[LEVEL_START_TIME]
	logs.Trace("[anticheat] CheckLevelTime client %v server %v", clientTime, param.levelTime)
	if float64(clientTime) > (float64(param.levelTime)*1.2 + float64(5)) {
		logs.Warn("[anticheat-fail] CheckLevelTime client %v server %v", clientTime, param.levelTime)
		curAttr := param.account.Profile.GetData().CorpAttrs
		var mid_info Avatar2ClientJson
		mid_info.FromAccount(param.account, param.account.Profile.CurrAvatar)
		anticheatlog.Trace(param.acid, checker_strs[CheckerLevelTime],
			AnticheatLogInfo{
				Avatar:          param.account.Profile.CurrAvatar,
				ClientInfo:      param.info,
				Attrs:           curAttr,
				IsPvp:           param.isPvp,
				GS:              param.gs,
				Mid:             mid_info,
				ClientLevelTime: clientTime,
				ServerLevelTime: param.levelTime,
			}, "anticheat_fail")
		return true
	}
	return
}
func checkSkillCD(param anticheat_param) (isFail bool) {
	avatarId := param.account.Profile.CurrAvatar
	isFail = isFail || _checkSkillCD(avatarId, SKILL1_CD_MINTIME, param)
	isFail = isFail || _checkSkillCD(avatarId, SKILL2_CD_MINTIME, param)
	isFail = isFail || _checkSkillCD(avatarId, SKILL3_CD_MINTIME, param)
	isFail = isFail || _checkSkillCD(avatarId, SKILL4_CD_MINTIME, param)
	isFail = isFail || _checkSkillCD(avatarId, SKILL5_CD_MINTIME, param)
	return
}

func _checkSkillCD(avatarId, clientParamIdx int, param anticheat_param) (isFail bool) {
	skillIdx := transform2skillId(avatarId, clientParamIdx, param.isPvp)
	if skillIdx < 0 {
		return false
	}
	if param.info[clientParamIdx] > 100 {
		return false
	}
	level := param.account.Profile.GetAvatarSkill().GetByAvatar(param.account.Profile.CurrAvatar)[skillIdx]
	serverCd := gamedata.GetSkillLevelConfig(avatarId, skillIdx).CDTime[level]
	logs.Trace("[anticheat] CheckSkillCD skillidx %d skilllevel %d client %v server %v",
		skillIdx,
		level,
		param.info[clientParamIdx],
		serverCd)
	if param.info[clientParamIdx] < serverCd {
		logs.Warn("[anticheat-fail] CheckSkillCD skillidx %d skilllevel %d client %v server %v",
			skillIdx,
			level,
			param.info[clientParamIdx],
			serverCd)
		curAttr := param.account.Profile.GetData().CorpAttrs
		var mid_info Avatar2ClientJson
		mid_info.FromAccount(param.account, param.account.Profile.CurrAvatar)
		anticheatlog.Trace(param.acid, checker_strs[CheckerSkillCD],
			AnticheatLogInfo{
				Avatar:         param.account.Profile.CurrAvatar,
				ClientInfo:     param.info,
				Attrs:          curAttr,
				IsPvp:          param.isPvp,
				GS:             param.gs,
				Mid:            mid_info,
				ClientParamIdx: clientParamIdx,
				ClientCD:       param.info[clientParamIdx],
				ServerCD:       serverCd,
			}, "anticheat_fail")
		return true
	}
	return false
}

func checkHPMPRageLimit(param anticheat_param) (isFail bool) {

	curAttr := param.account.Profile.GetData().CorpAttrs
	curHp := curAttr.HP + 10
	logs.Trace("[anticheat] checkHPMPRageLimit HP client %v server %v", param.info[HP_MAX], curHp)
	if param.info[HP_MAX] > curHp {
		logs.Warn("[anticheat-fail] checkHPMPRageLimit HP client %v server %v", param.info[HP_MAX], curHp)
		var mid_info Avatar2ClientJson
		mid_info.FromAccount(param.account, param.account.Profile.CurrAvatar)
		anticheatlog.Trace(param.acid, checker_strs[CheckerHPMPRageLimit],
			AnticheatLogInfo{
				Avatar:     param.account.Profile.CurrAvatar,
				ClientInfo: param.info,
				Attrs:      curAttr,
				IsPvp:      param.isPvp,
				GS:         param.gs,
				Mid:        mid_info,
				Type:       "HP",
			}, "anticheat_fail")
		isFail = isFail || true
	}
	logs.Trace("[anticheat] checkHPMPRageLimit MP client %v server %v", param.info[ENERGY_MAX], 100)
	if param.info[ENERGY_MAX] > 100 { // mp上限写死100
		logs.Warn("[anticheat-fail] checkHPMPRageLimit MP client %v server %v", param.info[ENERGY_MAX], 100)
		var mid_info Avatar2ClientJson
		mid_info.FromAccount(param.account, param.account.Profile.CurrAvatar)
		anticheatlog.Trace(param.acid, checker_strs[CheckerHPMPRageLimit],
			AnticheatLogInfo{
				Avatar:     param.account.Profile.CurrAvatar,
				ClientInfo: param.info,
				Attrs:      curAttr,
				IsPvp:      param.isPvp,
				GS:         param.gs,
				Mid:        mid_info,
				Type:       "MP",
			}, "anticheat_fail")
		isFail = isFail || true
	}
	return
}

func checkDamageResist(param anticheat_param) (isFail bool) {

	return
}

func checkHPMPRageCostGet(param anticheat_param) (isFail bool) {

	return
}

func checkHeroGS(param anticheat_param) (isFail bool) {
	cfg := game.Cfg.AntiCheat[CheckerHeroGS]
	for i := 0; i < MAX_CHECK_AVATAR; i++ {
		if hero_idx_id[i] > len(param.info) || hero_idx_gs[i] > len(param.info) {
			break
		}
		id := int(param.info[hero_idx_id[i]])
		if id >= AVATAR_NUM_CURR {
			continue
		}
		gs := param.account.Profile.GetData().HeroGs[int(param.info[hero_idx_id[i]])]
		ch_gs := int(param.info[hero_idx_gs[i]])
		logs.Debug("checkHeroGS %s hero %d gs %d %d", param.account.AccountID.String(),
			id, gs, ch_gs)
		if ch_gs-gs > int(cfg.ParamInt) {
			logs.Warn("checkHeroGS %s hero %d gs %d %d", param.account.AccountID.String(),
				id, gs, ch_gs)
			anticheatlog.Trace(param.acid, checker_strs[CheckerHeroGS],
				AnticheatLogInfo{
					Avatar:        id,
					HeroGS:        gs,
					CheckedHeroGS: ch_gs,
					ClientInfo:    param.info,
					Typ:           param.typ,
				}, "anticheat_fail")
			isFail = isFail || true
		}
	}
	return
}

// 检查战斗相关
func (ac *AntiCheat) CheckFightRelAll(acid string, info []float32,
	account *Account, typ string, levelTime int64) (failed []int) {
	failed = make([]int, 0, CheckerCount)
	if game.Cfg.AntiCheatValid {
		for i, _ := range checkers {
			if ac.check(i, anticheat_param{
				acid:      acid,
				info:      info,
				account:   account,
				levelTime: levelTime,
				typ:       typ,
			}) {
				failed = append(failed, i)
			}
		}
	}
	return
}

// 检查通关时间
func checkPassTime(param anticheat_param) (isFail bool) {
	if Pass_Time > len(param.info) {
		logs.Warn("arg len err, len: %v", param.info)
		return false
	}
	passTime := param.info[Pass_Time]
	var time int32
	switch param.typ {
	case Anticheat_Typ_SimplePVP:
		time = gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_PVP)
	case Anticheat_Typ_TeamPVP:
		time = int32(gamedata.GetTPvpCommonCfg().GetServerServeTime())
	case Anticheat_Typ_Expedition:
		time = gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_EXPEDITION)
	case Anticheat_Typ_GuildBoss:
		time = gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_GUILDBOSS)
	case Anticheat_Typ_BossFight:
		time = gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_BOSS)
	case Anticheat_Typ_GVG:
		time = gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_GVG)
	case Anticheat_Typ_LevelStage:
		time = 10 * 60
	case Anticheat_Typ_Trial:
		time = gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_TRIAL)
	case Anticheat_Typ_FestivalBoss:
		time = gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_FESTIVAL)
	case Anticheat_Typ_HeroDiff:
		time1 := float64(gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_HERODIFF_TU))
		time2 := float64(gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_HERODIFF_HU))
		time3 := float64(gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_HERODIFF_ZHAN))
		time4 := float64(gamedata.GetStageTimeLimit(gamedata.LEVEL_TYPE_HERODIFF_SHI))
		time = int32(math.Max(time4, math.Max(time3, math.Max(time1, time2))))
	case Anticheat_Typ_Wspvp:
		time = int32(gamedata.WsPvpMainCfg.Config.GetFightMaxTime())
	case Anticheat_Typ_CSRob:
		time = int32(gamedata.CSRobRCConfig.GetLevelTime())
	case Anticheat_Typ_WB:
		break
	case Anticheat_Typ_GateEnemy:
		t, _ := gamedata.GetGVEGameCfg()
		time = int32(t * 60)
	}
	logs.Debug("valid time: %v", time)
	if time == 0 || int64(passTime) <= int64(float32(time)*float32(1.1)) {
		return false
	}
	return true
}

func (ac *AntiCheat) check(index int, param anticheat_param) (isFail bool) {
	if index >= CheckerCount {
		logs.Error("anticheat check param err %d", index)
		return true
	}
	cfg := game.Cfg.AntiCheat[index]
	if cfg.IsCheck {
		ch := checkers[index]
		isFail = ch(param)
	}
	if cfg.IsRecord && isFail {
		ac.Cheated[index] = true
	}
	if cfg.IsBan && isFail && time.Now().Unix()-ac.BanTime > game.Cfg.BanTime {
		account, err := db.ParseAccount(param.acid)
		if err != nil {
			logs.Error("anticheat ParseAccount err %s ", param.acid)
			return
		}
		url := fmt.Sprintf("%s/%s?time=%d&gid=%d", game.Cfg.BanUrl, account.UserId, game.Cfg.BanTime, account.GameId)
		code, body, err := httpGet(url)
		if err != nil {
			logs.Error("anticheat ban err %s", err.Error())
			return
		}
		if code != http.StatusOK {
			var res map[string]interface{}
			if err := json.Unmarshal(body, &res); err != nil {
				logs.Error("anticheat ban body unmarshal err %s", err.Error())
				return
			}
			logs.Error("anticheat ban fail status %d msg %v", code, res)
			return
		}
		ac.BanTime = time.Now().Unix()
	}
	return
}

func (ac *AntiCheat) DBName() string {
	return ac.dbkey.String()
}

func (ac *AntiCheat) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
	key := ac.DBName()

	if forceDirty {
		ac.dirtiesCheck = nil
	}
	err, newDirtyCheck, chged := driver.DumpToHashDBCmcBufferCheckDirty(
		cb, key, ac, ac.dirtiesCheck)
	if err != nil {
		return err
	}
	if !game.Cfg.IsRunModeProd() {
		if !reflect.DeepEqual(ac.dirtiesCheck, newDirtyCheck) {
			logs.Trace("Save PlayerStores %s %v", ac.dbkey.Account.String(), chged)
		} else {
			logs.Trace("Save PlayerStores clean %s", ac.dbkey.Account.String())
		}
	}
	ac.dirtiesCheck = newDirtyCheck
	return nil
}

func (ac *AntiCheat) DBLoad(logInfo bool) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	key := ac.DBName()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, ac, false, logInfo)

	// RESTORE_ERR_Profile_No_Data 表示玩家第一次登陆游戏，没有存档，这不视为Bug
	// 外面的逻辑需要根据此判断是否是第一次登陆游戏
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return nil
	}
	ac.dirtiesCheck = driver.GenDirtyHash(ac)
	return nil
}

func httpGet(url string) (int, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("anticheat HttpGet err %s", err.Error())
		return resp.StatusCode, []byte{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logs.Error("anticheat HttpGet ReadAll err %s", err.Error())
		return resp.StatusCode, []byte{}, err
	}

	return resp.StatusCode, body, nil
}

func transform2skillId(avatarId, clientParam int, isPvp bool) int {
	return _transform2skillId(avatarId, clientParam, isPvp) - 1
}

// 反作弊根据客户端参数返回技能idx，是没减一的
func _transform2skillId(avatarId, clientParam int, isPvp bool) int {
	switch clientParam {
	case SKILL1_CD_MINTIME:
		return 1
	case SKILL2_CD_MINTIME:
		if avatarId == 1 {
			return 2
		}
		return 9
	case SKILL3_CD_MINTIME:
		if isPvp {
			if avatarId == 1 {
				return 9
			} else {
				return 2
			}
		} else {
			if avatarId == 1 {
				return 3
			} else {
				return 2
			}
		}
	case SKILL4_CD_MINTIME:
		if isPvp {
			return 0
		}
		if avatarId == 1 {
			return 9
		}
		return 3
	case SKILL5_CD_MINTIME:
		if isPvp {
			return 0
		}
		return 4
	}
	return 0
}
