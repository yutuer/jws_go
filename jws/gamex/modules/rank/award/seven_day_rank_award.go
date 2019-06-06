package award

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/modules/title_rank"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GuildSevOpnRank struct {
	Uuid string
	Pos  int
}

var (
	guildAwardChan chan GuildSevOpnRank
)

func init() {
	guildAwardChan = make(chan GuildSevOpnRank, 128)
}

func GetGuildAwardChan() <-chan GuildSevOpnRank {
	return guildAwardChan
}

func AwardByCorpGs(shardId uint, topNAcid []string, topNScore []int64,
	acid2Score map[string]int64, gsPowBase int64) {

	// 发个人排名奖
	for idx, acid := range topNAcid {
		rank := idx + 1
		score := topNScore[idx]
		rcfg := gamedata.GetRankAward(uint32(rank), score/gsPowBase)
		if rcfg == nil {
			break
		}
		items := make([]string, 0, len(rcfg.GetAwardItems_Template()))
		counts := make([]uint32, 0, len(rcfg.GetAwardItems_Template()))
		for _, r := range rcfg.GetAwardItems_Template() {
			items = append(items, r.GetItemID())
			counts = append(counts, r.GetAmount())
		}

		mail_sender.BatchSend7DayPlayerRankReward(shardId, acid, rank, items, counts)

		logs.Trace("SevenDayRank person award rank %d acid %s", rank, acid)
	}
	// 称号
	title_rank.GetModule(shardId).Set7DayGsRank(topNAcid)

	// 发个人gs阶段奖
	for acid, score := range acid2Score {
		gs := score / gsPowBase
		fcfg := gamedata.GetFightAward(gs)
		if fcfg != nil {
			items := make([]string, 0, len(fcfg.GetAwardItems_Template()))
			counts := make([]uint32, 0, len(fcfg.GetAwardItems_Template()))
			for _, r := range fcfg.GetAwardItems_Template() {
				items = append(items, r.GetItemID())
				counts = append(counts, r.GetAmount())
			}

			mail_sender.BatchSend7DayPlayerGsReward(shardId, acid, items, counts)

			logs.Trace("SevenDayRank gs award  acid %s gs %d", acid, gs)
		}
	}
}

func AwardGuild(topNUuid []string) {
	for idx, uuid := range topNUuid {
		if uuid == "" {
			break
		}
		rank := idx + 1
		rCfg, _ := gamedata.GetGuildRankAwardCfg(rank)
		if rCfg == nil {
			break
		}
		// send
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		select {
		case guildAwardChan <- GuildSevOpnRank{
			Uuid: uuid,
			Pos:  rank,
		}:
		case <-ctx.Done():
			logs.Error("[SeverOpenRank] AwardGuild chann full, cmd put timeout")
		}
		cancel()
	}
}
