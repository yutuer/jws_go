package account

import (
	"time"

	"reflect"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type TmpProfile struct {
	dbkey        db.ProfileDBKey
	dirtiesCheck map[string]interface{}

	Ver      int64 `redis:"version"`
	LootRand uint16
	//json
	LootContent []byte
	// 关尾掉落
	StageRewards gamedata.PriceDataSet `redis:"stageRewards"`

	//LastTime   int64 `redis:"lasttime"`
	CreateTime int64 `redis:"createtime"`

	GoldLevelPoint uint32 // 金币关积分
	ExpLevelPoint  uint32 // 经验关积分
	DCLevelPoint   uint32 // 天命关积分

	// pvp信息
	SimplePvp         PlayerCurrSimplePvpState `redis:"pvp"`
	TmpSkipEquipCount int                      `redis:"tskipec"`

	Level_enter_time int64 // 关卡、pvp进入时间，为反作弊用，不存db

	Last_Level_Prepare    string
	Last_Level_Prepare_TS int64
	Last_Level_Declare    string
	Last_Level_Declare_TS int64
	Last_GetProto_TS      int64
	Last_GetInfo_TS       int64

	TrialFirst      bool // 首次开启爬塔，临时标记
	ExpeditionFirst bool // 首次开启远征，临时标记

	//
	CurrRoomNum int

	// gve
	CurrWaitting    bool
	CurrFighting    bool
	CurrGameEndTime int64
	GameID          string
	GameSecret      string
	GameUrl         string
	GameIsHard      bool
	GameIsUseHc     bool
	GameIsDouble    bool
	GameRewards     []string
	GameCounts      []uint32
	IsBot           bool
	CurrCount       int
	MatchBeginTime  int64

	BossIdx int64

	// 切磋信息
	GankFightAcid           string
	GankFightName           string
	GankFightRecordTimStamp int64
	GankIds                 int

	// gvg
	gvgData helper.GVGData

	GVGCity int
}

func NewTmpProfile(account db.Account) TmpProfile {
	now_t := time.Now().Unix()
	return TmpProfile{
		dbkey: db.ProfileDBKey{
			Account: account,
			Prefix:  "tmp",
		},
		//Ver:        helper.CurrDBVersion,
		//LastTime:   now_t,
		CreateTime: now_t,
	}
}

func (p *TmpProfile) DBName() string {
	return p.dbkey.String()
}

func (p *TmpProfile) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
	key := p.DBName()
	if p.GameRewards == nil {
		p.GameRewards = []string{}
	}

	if p.GameCounts == nil {
		p.GameCounts = []uint32{}
	}

	if forceDirty {
		p.dirtiesCheck = nil
	}
	err, newDirtyCheck, chged := driver.DumpToHashDBCmcBufferCheckDirty(
		cb, key, p, p.dirtiesCheck)
	if err != nil {
		return err
	}
	if !game.Cfg.IsRunModeProd() {
		if !reflect.DeepEqual(p.dirtiesCheck, newDirtyCheck) {
			logs.Trace("Save TmpProfile %s %v", p.dbkey.Account.String(), chged)
		} else {
			logs.Trace("Save TmpProfile clean %s", p.dbkey.Account.String())
		}
	}
	p.dirtiesCheck = newDirtyCheck
	return nil
}

func (p *TmpProfile) DBLoad(logInfo bool) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	key := p.DBName()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)

	// RESTORE_ERR_Profile_No_Data 表示玩家第一次登陆游戏，没有存档，这不视为Bug
	// 外面的逻辑需要根据此判断是否是第一次登陆游戏
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return nil
	}
	p.dirtiesCheck = driver.GenDirtyHash(p)

	if p.GameRewards == nil {
		p.GameRewards = []string{}
	}

	if p.GameCounts == nil {
		p.GameCounts = []uint32{}
	}

	return nil
}

func (p *TmpProfile) GetLevelEnterTime() int64 {
	return p.Level_enter_time
}

func (p *TmpProfile) SetLevelEnterTime(time int64) {
	p.Level_enter_time = time
}

func (p *TmpProfile) GetSimplePvpState() *PlayerCurrSimplePvpState {
	return &p.SimplePvp
}

func (p *TmpProfile) CleanStageData() {
	p.LootContent = []byte{}
}

func (tmp *TmpProfile) SetGankFightInfo(acid, name string, ts int64, ids int) {
	tmp.GankFightAcid = acid
	tmp.GankFightName = name
	tmp.GankFightRecordTimStamp = ts
	tmp.GankIds = ids
}

func (tmp *TmpProfile) GetGankFightInfo() (string, string, int64, int) {
	return tmp.GankFightAcid, tmp.GankFightName, tmp.GankFightRecordTimStamp, tmp.GankIds
}

func (tmp *TmpProfile) SetGVEData(gameID, secret, gameUrl string,
	rewards []string, cs []uint32, isBot bool) {
	tmp.CurrFighting = true
	fightTime, _ := gamedata.GetGVEGameCfg()
	tmp.CurrGameEndTime = time.Now().Unix() + fightTime*60 // 会晚一些 只是做保护用
	tmp.GameID = gameID
	tmp.GameSecret = secret
	tmp.GameUrl = gameUrl
	tmp.GameRewards = rewards
	tmp.GameCounts = cs
	tmp.IsBot = isBot
}

func (tmp *TmpProfile) CleanGVEData() {
	tmp.CurrWaitting = false
	tmp.CurrFighting = false
	tmp.GameID = ""
	tmp.GameSecret = ""
	tmp.GameUrl = ""
	tmp.GameIsHard = false
	tmp.GameIsUseHc = false
	tmp.GameIsDouble = false
	tmp.GameRewards = []string{}
	tmp.GameCounts = []uint32{}
	tmp.IsBot = false
}

func (tmp *TmpProfile) IsCurrWaittingGVE() bool {
	return tmp.CurrWaitting
}

func (tmp *TmpProfile) GetGVEData() (bool, string, string, string, bool) {
	if !tmp.CurrFighting || time.Now().Unix() >= tmp.CurrGameEndTime {
		return false, "", "", "", false
	} else {
		return true, tmp.GameID, tmp.GameSecret, tmp.GameUrl, tmp.IsBot
	}
}

func (tmp *TmpProfile) GetGVGData() *helper.GVGData {
	return &tmp.gvgData
}

func (tmp *TmpProfile) SetGVGData(
	gvgAvatarState [helper.GVG_AVATAR_COUNT]helper.AvatarState, gvgDestinySkill [helper.DestinyGeneralSkillMax]int, gvgEnemyAcID string,
	gvgEAvatarState [helper.GVG_AVATAR_COUNT]helper.AvatarState, gvgEDestinySkill [helper.DestinyGeneralSkillMax]int,
	gvgEnemyData [helper.GVG_AVATAR_COUNT]*helper.Avatar2Client,
	roomID string, url string) {
	tmp.gvgData = helper.GVGData{
		GvgAvatarState:   gvgAvatarState,
		GvgDestinySkill:  gvgDestinySkill,
		GvgEnemyAcID:     gvgEnemyAcID,
		GvgEAvatarState:  gvgEAvatarState,
		GvgEDestinySkill: gvgEDestinySkill,
		GvgEnemyData:     gvgEnemyData,
		Url:              url,
		RoomID:           roomID,
	}
}

func (tmp *TmpProfile) CleanGVGData() {
	tmp.SetGVGData([helper.GVG_AVATAR_COUNT]helper.AvatarState{}, [helper.DestinyGeneralSkillMax]int{}, "",
		[helper.GVG_AVATAR_COUNT]helper.AvatarState{}, [helper.DestinyGeneralSkillMax]int{-1, -1, -1},
		[helper.GVG_AVATAR_COUNT]*helper.Avatar2Client{}, "", "")
}
