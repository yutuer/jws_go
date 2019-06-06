package rank

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func TableRankCorpGs(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpGs", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func tableRankCorpGsSvrOpn(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpGsSvrOpn", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankSimplePvp(sid uint) string {
	return fmt.Sprintf("%d:%d:RankSimplePvp", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankGuildGS(sid uint) string {
	return fmt.Sprintf("%d:%d:RankGuildGS", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func tableRankGuildGsSvrOpn(sid uint) string {
	return fmt.Sprintf("%d:%d:RankGuildGsSvrOpn", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankGuildGateEnemy(sid uint) string {
	return fmt.Sprintf("%d:%d:RankGuildGateEnemy", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpTrial(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpTrial", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpHeroStar(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpHeroStar", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpHeroDiffTU(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpHeroDiff:TU", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpHeroDiffZHAN(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpHeroDiff:ZHAN", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpHeroDiffHU(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpHeroDiff:HU", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpHeroDiffSHI(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpHeroDiff:SHI", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankDestiny(sid uint) string {
	return fmt.Sprintf("%d:%d:RankDestiny", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankJade(sid uint) string {
	return fmt.Sprintf("%d:%d:RankJade", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankEquipStarLv(sid uint) string {
	return fmt.Sprintf("%d:%d:RankEquipStarLv", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpLv(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpLv", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankSwingStarLv(sid uint) string {
	return fmt.Sprintf("%d:%d:RankSwingStarLv", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankHeroDestinyLv(sid uint) string {
	return fmt.Sprintf("%d:%d:RankHeroDestinyLv", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankJadeTwo(sid uint) string {
	return fmt.Sprintf("%d:%d:RankRankJadeTwo", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankExclusiveWeapon(sid uint) string {
	return fmt.Sprintf("%d:%d:RankExclusiveWeapon", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankWuShuangGs(sid uint) string {
	return fmt.Sprintf("%d:%d:RankWuShuangGs", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankAstrology(sid uint) string {
	return fmt.Sprintf("%d:%d:RankAstrology", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpOfWei(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpOfWei", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpOfShu(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpOfShu", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpOfWu(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpOfWu", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableRankCorpOfQunXiong(sid uint) string {
	return fmt.Sprintf("%d:%d:RankCorpOfQunXiong", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
