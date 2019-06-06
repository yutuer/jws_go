package guild_boss

import (
	"math/rand"
	"time"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/base"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/common/guild_player_rank"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/common/player_state"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	BossChallengeStatNormal = iota
	BossChallengeStatLocked
	BossChallengeStatFighting
)

const (
	BossLevelNull = iota
	BossLevelNormal
	BossLevelMiddle
	BossLevelHigh
	BossLevelMaster
	BossLevelHell
)

const (
	BossTypNormal = iota
	BossTypBig
)

const (
	BossStatNormal = iota
	BossStatKilled
	BossStatUnlocked
)

type ActivityState struct {
	base.ActBase
	player_state.PlayerActivityStates

	LastDayDamages guild_player_rank.GuildPlayerRank `json:"ldd"`
	TodayDamages   guild_player_rank.GuildPlayerRank `json:"tdd"`
	Bosses         []BossState                       `json:"bosses"`
	BigBoss        BossState                         `json:"bigboss"`
	BossDegree     int                               `json:"degree"`
	Statictic      logiclog.DailyStatistics          `json:"stic"` // 为bi信息的加的统计信息
}

type BossState struct {
	Idx                int                               `json:"idx"`
	ID                 string                            `json:"id"`
	Hp                 int64                             `json:"hp"`
	TotalHp            int64                             `json:"totalhp"`
	GroupId            string                            `json:"group"`
	PartRewardDroped   int                               `json:"part"`
	CurrPlayerAcID     string                            `json:"acid"`
	CurrPlayerName     string                            `json:"name"`
	CurrPlayerAvatarID int                               `json:"avatar"`
	CurrPlayerState    int                               `json:"stat"`
	CurrPlayerStopTime int64                             `json:"stopTime"`
	SelfLoot           string                            `json:"loot"`
	ItemC              map[string]uint32                 `json:"item_c"`
	MVPRank            guild_player_rank.GuildPlayerRank `json:"mvp"`
	LevelId            string                            `json:"levelId"`
	LevelTime          int64                             `json:"lvTime"`
}

type BossStatistic struct {
	TS           int64    `json:"ts"`
	JoinTimes    int      `json:"j_ts"`
	JoinMemCount []string `json:"j_m_c"`
}

func (b *BossState) FromData(data *ProtobufGen.GUILDBOSS_BossData, rd *rand.Rand) {
	bossId, ok := gamedata.RandGuildBossEnemy(data.GetBossGroupID(), rd)
	if !ok {
		logs.Error("FromData Err by RandGuildBossEnemy %s", data.GetBossGroupID())
		return
	}

	acData, ok := gamedata.GetAcData(bossId)
	if !ok {
		logs.Error("FromData Err By GetAcData %s", bossId)
	}

	b.ID = bossId
	b.Hp = int64(acData.GetHitPoint())
	b.TotalHp = int64(acData.GetHitPoint())
	b.GroupId = data.GetBossGroupID()
	b.PartRewardDroped = 4 // 每%20给一个奖励 共四个
	b.SelfLoot = data.GetSelfLoot()
	b.ItemC = make(map[string]uint32, 4)
	if data.GetSelfLootB() != "" {
		b.ItemC[data.GetSelfLootB()] = data.GetSelfLootBNum()
	}
	if data.GetSelfLootC() != "" {
		b.ItemC[data.GetSelfLootC()] = data.GetSelfLootCNum()
	}
	if data.GetSelfLootD() != "" {
		b.ItemC[data.GetSelfLootD()] = data.GetSelfLootDNum()
	}
	b.LevelId = data.GetLevelDemo()
	stage := gamedata.GetStageData(b.LevelId)
	if stage == nil {
		logs.Error("stage info nil by %s", b.LevelId)
		b.LevelTime = 60
	} else {
		b.LevelTime = int64(stage.TimeLimit)
	}
}

func (b *BossState) updateChallengeStat(nowT int64) {
	if b.CurrPlayerAcID == "" {
		return
	} else {
		if nowT > b.CurrPlayerStopTime {
			b.CurrPlayerAcID = ""
			b.CurrPlayerName = ""
			b.CurrPlayerAvatarID = 0
			b.CurrPlayerState = BossChallengeStatNormal
			b.CurrPlayerStopTime = 0
			logs.Debug("challenge boss time expired, %d, %d", nowT, b.CurrPlayerStopTime)
		}
	}
}

func (b *BossState) OnMemberKick(acid string) {
	if b.CurrPlayerAcID == acid {
		b.CurrPlayerState = BossChallengeStatNormal
		b.CurrPlayerAcID = ""
		b.CurrPlayerName = ""
		b.CurrPlayerAvatarID = 0
		b.CurrPlayerStopTime = 0
	}
}

func (a *ActivityState) Init() {
	a.TodayDamages.InitOnRestart()
	a.BigBoss.MVPRank.InitOnRestart()
	for i := range a.Bosses {
		a.Bosses[i].MVPRank.InitOnRestart()
	}
}

func (a *ActivityState) UpdateInfo(info *helper.AccountSimpleInfo) {
	a.TodayDamages.UpdatePlayerInfo(info)
	a.BigBoss.MVPRank.UpdatePlayerInfo(info)
	for i := range a.Bosses {
		a.Bosses[i].MVPRank.UpdatePlayerInfo(info)
	}
}

func (a *ActivityState) Clean(guildUuid, name string) {
	a.PlayerActivityStates.Clean()
	a.log(guildUuid, name)

	if a.BigBoss.Hp <= 0 {
		cfg := gamedata.GetGuildBossDataByLv(uint32(a.BossDegree) + 1)
		if cfg != nil {
			guildLv := a.GetGuildHandler().GetGuildLv()
			if guildLv >= cfg.GetGuildLvReqirement() {
				a.BossDegree++
			}
		}

		if a.BossDegree > BossLevelHell {
			a.BossDegree = BossLevelHell
		}
	}

	rd := rand.New(rand.NewSource(time.Now().Unix()))

	bossData := gamedata.GetGuildBossDataByLv(uint32(a.BossDegree))
	if bossData == nil {
		logs.Error("guild bossData nil By %v", a.BossDegree)
		return
	}

	bossCount := len(bossData.GetBossData_Table())
	//logs.Warn("boss clean %v", bossCount)
	a.Bosses = make([]BossState, bossCount-1, bossCount-1)
	ts := bossData.GetBossData_Table()
	for i := 0; i < len(ts); i++ {
		logs.Trace("ActivityState Clean Boss %v %v", i, ts[i])
		if i < bossCount-1 {
			a.Bosses[i].FromData(ts[i], rd)
			a.Bosses[i].MVPRank.Clean()
			a.Bosses[i].Idx = i

		} else {
			a.BigBoss.FromData(ts[i], rd)
			a.BigBoss.Idx = i
			a.BigBoss.MVPRank.Clean()
		}
	}

	a.LastDayDamages = a.TodayDamages
	a.TodayDamages.Clean()

	logs.Trace("a.Bosses %v %v", a.Bosses, a.BigBoss)

}

func (a *ActivityState) GetBossStat(idx int) int64 {
	if idx < len(a.Bosses) {
		// common boss state
		if a.Bosses[idx].Hp > 0 {
			return BossStatNormal
		} else {
			return BossStatKilled
		}
	} else {
		// big boss state
		allPassed := true
		for i := 0; i < len(a.Bosses); i++ {
			if a.Bosses[i].Hp > 0 {
				allPassed = false
			}
		}
		if allPassed {
			if a.BigBoss.Hp > 0 {
				return BossStatNormal
			} else {
				return BossStatKilled
			}
		} else {
			return BossStatUnlocked
		}
	}
}

func (a *ActivityState) GetBossById(id, group string) *BossState {
	// Boss 很少
	logs.Trace("Boss %v --> %s %s", a.Bosses, id, group)
	for i := 0; i < len(a.Bosses); i++ {
		if a.Bosses[i].ID == id && a.Bosses[i].GroupId == group {
			return &a.Bosses[i]
		}
	}

	if a.BigBoss.ID == id && a.BigBoss.GroupId == group {
		return &a.BigBoss
	}

	return nil
}

func (a *ActivityState) log(guildUuid, name string) {
	boss := make([]logiclog.LogicInfo_GuildBoss, len(a.Bosses))
	for i, b := range a.Bosses {
		boss[i] = logiclog.LogicInfo_GuildBoss{
			BossId:     b.ID,
			LeftHpRate: fmt.Sprintf("%.2f", float64(b.Hp)/float64(b.TotalHp)),
		}
	}
	bigBoss := logiclog.LogicInfo_GuildBoss{
		BossId:     a.BigBoss.ID,
		LeftHpRate: fmt.Sprintf("%.2f", float64(a.BigBoss.Hp)/float64(a.BigBoss.TotalHp)),
	}
	logiclog.LogGuildBoss(guildUuid, name, a.BossDegree, boss, bigBoss, "")
}
