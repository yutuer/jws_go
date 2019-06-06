package account

import (
	"errors"
	"math"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/message"
	"vcs.taiyouxi.net/jws/gamex/modules/global_count"
	accountDB "vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const SimplePvpFightRecordCount = 50
const SimplePvpFightRecordKey = "SimplePvpRecord"
const SimplePvpFightRecordCount2Client = 10

type SimplePvpFightRecord struct {
	IsSuccess         bool   `codec:"s" json:"s"`
	RankChange        int    `codec:"rc" json:"rc"`
	AvatarID          int    `codec:"aid" json:"aid"`
	AvatarStarLv      int    `codec:"aslv" json:"aslv"`
	EnemyAvatarID     int    `codec:"eaid" json:"eaid"`
	EnemyAvatarStarLv int    `codec:"ealv" json:"ealv"`
	EnemyLv           uint32 `codec:"elv" json:"elv"`
	EnemyID           string `codec:"eid" json:"eid"`
	EnemyName         string `codec:"ename" json:"ename"`
	EnemyGS           int    `codec:"egs" json:"egs"`
	EnemyRank         int    `codec:"er" json:"er"`
	Time              int64  `codec:"t" json:"t"`
	IsAttack          bool   `codec:"attack" json:"attack"`
	RecordNum         uint32 `codec:"record" json:"record"`
}

func (s SimplePvpFightRecord) SendToDB(accountID string) {

	// 打上History Record 次数戳
	account, err := accountDB.ParseAccount(accountID)
	if err != nil {
		logs.SentryLogicCritical(accountID,
			"Get Account Info Err By %s", err.Error())
		return
	}
	counts := global_count.GetRecordCount(account.ShardId, account.GameId)
	value, ok := counts[global_count.SimplePvpRecord]
	if !ok {
		logs.SentryLogicCritical(accountID,
			"Get SimplePvp History Record Err")
		return
	}
	s.RecordNum = value

	b, err := json.Marshal(s)
	if err != nil {
		logs.SentryLogicCritical(accountID,
			"SimplePvpFightRecord Marshal Err By %s", err.Error())
		return
	}

	msg := message.PlayerMsg{
		Params: []string{string(b)},
	}
	message.SendPlayerMsgs(accountID,
		SimplePvpFightRecordKey,
		SimplePvpFightRecordCount,
		msg)
}

func loadSimplePvpFightRecord(accountID string) []SimplePvpFightRecord {
	res, err := message.LoadPlayerMsgs(accountID,
		SimplePvpFightRecordKey, SimplePvpFightRecordCount)
	if err != nil {
		return []SimplePvpFightRecord{}
	}

	records := make([]SimplePvpFightRecord, 0, len(res))
	for i := 0; i < len(res); i++ {
		newMsg := SimplePvpFightRecord{}
		if len(res[i].Params) < 1 {
			continue
		}
		err := json.Unmarshal([]byte(res[i].Params[0]), &newMsg)
		if err != nil {
			continue
		}
		account, err := accountDB.ParseAccount(accountID)
		if err != nil {
			logs.SentryLogicCritical(accountID,
				"Get Account Info Err By %s", err.Error())
			continue
		}
		counts := global_count.GetRecordCount(account.ShardId, account.GameId)
		value, ok := counts[global_count.SimplePvpRecord]
		if !ok {
			logs.SentryLogicCritical(accountID,
				"Get SimplePvp History Record Err")
			continue
		}
		if value == newMsg.RecordNum {
			records = append(records, newMsg)
		}
	}
	return records[:]
}

type PlayerSimplePvp struct {
	PvpCountByAvatar [AVATAR_NUM_CURR]int `json:"count"`
	PvpCountAll      int                  `json:"pca"`
	LastPvpCountTime int64                `json:"lc"`
	Score            int64                `json:"score"`
	PvpDefAvatar     int                  `json:"def"`
	SwitchCount      int                  `json:"swt_c"`
	LastGetTime      int64                `json:"lg_c"`
	PvpCountToday    int                  `json:"pct"`
	OpenedChestIDs   []uint32             `json:"open_c_id"`
	ResetChestTime   int64                `json:"rs_c_t`
	ResetSwitchTime  int64                `json:"rs_sw_t"`
}

type PlayerCurrSimplePvpState struct {
	Enemys              [SimplePvpEnemyCount]helper.Avatar2Client     `json:"es"`
	EnemySimpleInfos    [SimplePvpEnemyCount]helper.AccountSimpleInfo `json:"einfos"`
	IsPvping            bool                                          `json:"pvp"`
	CurrPvpAvatar       int                                           `json:"curr"`
	CurrEnemy           helper.Avatar2Client                          `json:"ces"`
	CurrEnemySimpleInfo helper.AccountSimpleInfo                      `json:"ceinfo"`
}

func (p *PlayerSimplePvp) CanOpenChest(id uint32) bool {
	for _, chestID := range p.OpenedChestIDs {
		if chestID == id {
			return false
		}
	}
	return true
}

func (p *PlayerSimplePvp) SetChestOpen(id uint32) {
	p.OpenedChestIDs = append(p.OpenedChestIDs, id)
}

func (p *PlayerSimplePvp) UpdateChestInfo(now_t int64) {
	if now_t < p.ResetChestTime {
		return
	}
	p.ResetChestTime = util.DailyBeginUnixByStartTime(now_t,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	p.ResetChestTime += util.DaySec
	p.OpenedChestIDs = make([]uint32, 0, 10)
	p.PvpCountToday = 0

}

func (p *PlayerSimplePvp) UpdateSwitchCount(now_t int64) {
	if now_t < p.ResetSwitchTime {
		return
	}
	p.ResetSwitchTime = util.DailyBeginUnixByStartTime(now_t,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	p.ResetSwitchTime += util.DaySec
	p.SwitchCount = 0
}

func (p *PlayerSimplePvp) IsDroid() bool {
	return p.PvpCountAll < SimplePvpDroidCount
}

func (p *PlayerSimplePvp) CanAutoSwitch(nowTime int64) bool {
	return p.LastGetTime+gamedata.GetSimplePvpConfig().RefreshEnemyTime <= nowTime
}

func (p *PlayerCurrSimplePvpState) AddEnemy(
	enemys [SimplePvpEnemyCount]helper.Avatar2Client,
	enemySimpleInfos [SimplePvpEnemyCount]helper.AccountSimpleInfo) {
	p.Enemys = enemys
	p.EnemySimpleInfos = enemySimpleInfos
}

func (p *PlayerSimplePvp) GetScore() int64 {
	return p.Score
}

func (p *PlayerCurrSimplePvpState) GetCurrEnemy() (helper.Avatar2Client, helper.AccountSimpleInfo) {
	return p.CurrEnemy, p.CurrEnemySimpleInfo
}

func (p *PlayerCurrSimplePvpState) GetCurrAvatar() int {
	return p.CurrPvpAvatar
}

func (sp *PlayerCurrSimplePvpState) Reset() {
	sp.IsPvping = false
	sp.CurrEnemy = helper.Avatar2Client{}
	sp.CurrEnemySimpleInfo = helper.AccountSimpleInfo{}
	sp.Enemys = [SimplePvpEnemyCount]helper.Avatar2Client{}
	sp.EnemySimpleInfos = [SimplePvpEnemyCount]helper.AccountSimpleInfo{}
	sp.CurrPvpAvatar = 0
}

func (p *PlayerSimplePvp) GetSimplePvpHistory(accountID string) []SimplePvpFightRecord {
	return loadSimplePvpFightRecord(accountID)
}

func (p *PlayerSimplePvp) AddSimplePvpHistory(accountID string, r SimplePvpFightRecord) {
	r.SendToDB(accountID)
}

func (p *PlayerSimplePvp) CheckPvpEnd(sp *PlayerCurrSimplePvpState) {
	logs.Trace("PlayerCurrSimplePvpState %v", *sp)
}

func (p *PlayerSimplePvp) OnPvpBegin(curr_avatar, enemy_idx int, sp *PlayerCurrSimplePvpState) error {
	logs.Trace("OnPvpBegin %v", *sp)

	if enemy_idx < 0 || enemy_idx >= len(sp.Enemys) {
		return errors.New("EnemyIdxErr")
	}

	sp.CurrEnemy = sp.Enemys[enemy_idx]
	if sp.CurrEnemy.GetAcId() == "" {
		return errors.New("NoEnemyInfo")
	}
	sp.CurrEnemySimpleInfo = sp.EnemySimpleInfos[enemy_idx]

	sp.IsPvping = true
	sp.CurrPvpAvatar = curr_avatar
	// 增加每日PVP次数
	p.PvpCountToday += 1

	return nil
}

func (p *PlayerSimplePvp) OnPvpEnd(is_success int, sp *PlayerCurrSimplePvpState) (int64, int64, error) {
	logs.Trace("OnPvpEnd %v", *sp)

	p.PvpCountAll += 1

	if !sp.IsPvping {
		return 0, 0, errors.New("NoPvping")
	}

	Ra := float64(p.GetScore())
	Rb := float64(sp.CurrEnemy.SimplePvpScore)

	//logs.Warn("PlayerSimplePvp End by %d %v %v", is_success, Ra, Rb)

	a, b := p.mkScore(is_success, Ra, Rb)
	//logs.Warn("PlayerSimplePvp End by %d %v %v", is_success, a, b)
	sa_d := int64(math.Floor(a * SimplePvpScorePow))
	sb_d := int64(math.Floor(b * SimplePvpScorePow))

	//logs.Warn("PlayerSimplePvp End by %d %d %d", is_success, sa_d, sb_d)

	p.Score += int64((math.Floor(a)))

	sp.Reset()

	return sa_d, sb_d, nil
}

/*
   每个角色有自己的一个等级分。
   玩家的一个角色满足入榜条件后会被赋予“初始等级分”。

   当两个角色交战时，按照以下算法来调整双方的等级分：
   Ra：A玩家赛前的等级分
   Rb：B玩家赛前的等级分
   Ea：预期A玩家的胜负值，Ea=1/(1+10^[(Rb-Ra)/400])
   Eb：预期B玩家的胜负值，Eb=1/(1+10^[(Ra-Rb)/400])
   Sa，Sb：实际胜负值，胜=1，平=0.5，负=0

   A的等级分变量 = BasicPvpEloK*(Sa-Ea)
   B的等级分变量 = BasicPvpEloK*(Sb-Eb)
   然后将胜方的等级分变量乘以 CommonConfig.BasicPvpWinnerBonus = 3

   A，B两个玩家赛后的等级分分别变为：
   R’a = Ra + A的等级分变量
   R’b = Rb + B的等级分变量

   BasicPvpEloK：是一个极限值，代表理论上最多可以赢一个玩家的得分和失分。
   暂时约定BasicPvpEloK=16，配置在CommonConfig表里。
*/

func (p *PlayerSimplePvp) mkScore(is_success int, Ra, Rb float64) (a, b float64) {
	var BasicPvpEloK float64 = float64(gamedata.GetSimplePvpConfig().PVPElok)
	var BasicPvpWinnerBonus float64 = gamedata.GetSimplePvpConfig().WinnerScoreX
	Ra_Rb := Ra - Rb
	Rb_Ra := Rb - Ra
	Ea := float64(1.0) / (float64(1.0) + math.Pow(10, (Rb_Ra/400.0)))
	Eb := float64(1.0) / (float64(1.0) + math.Pow(10, (Ra_Rb/400.0)))

	var Sa float64 = 0.5
	var Sb float64 = 0.5
	var Xa float64 = 1
	var Xb float64 = 1
	if is_success < 0 {
		Sa = 0.0
		Sb = 1.0
		Xb = BasicPvpWinnerBonus
	} else if is_success > 0 {
		Sa = 1.0
		Sb = 0.0
		Xa = BasicPvpWinnerBonus
	}
	//logs.Warn("PlayerSimplePvp End by %d %v %v - %v - %v %v", is_success,
	//	Ra_Rb, Rb_Ra, BasicPvpEloK, Ea, Eb)

	a = Xa * BasicPvpEloK * (Sa - Ea)
	b = Xb * BasicPvpEloK * (Sb - Eb)
	return
}
