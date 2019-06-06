package title_rank

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func TableTitleSimplePvpRank(sid uint) string {
	return fmt.Sprintf("%d:%d:RankSimplePvpForTitle", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableTitleTeamPvpRank(sid uint) string {
	return fmt.Sprintf("%d:%d:RankTeamPvpForTitle", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableTitle7DayGsRank(sid uint) string {
	return fmt.Sprintf("%d:%d:Rank7DayGsForTitle", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableTitleWuShuangRank(sid uint) string {
	return fmt.Sprintf("%d:%d:RankWuShuangForTitle", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
