package merge

import "fmt"

const (
	prepare_B_table    = "profile|bag|general|pguild|store|simpleinfo|friend|anticheat|tmp|guild:Info|expeditenemy"
	prepare_table_2    = "guild:account2guild|guild:id2uuid|guild:uuid|guild:idseed|fishreward|global:levelfinish|globalcount:acid7day"
	prepare_table_rank = "RankCorpGs|RankCorpGsSvrOpn|RankSimplePvp|RankGuildGS|RankGuildGsSvrOpn|RankCorpTrial|RankBalance|RankCorpHeroStar|BIHour"
	prepare_except     = "topN"
	profile_prefix     = "profile:"
)

var (
	prepare_A_table = fmt.Sprintf("%s|%s", prepare_B_table, prepare_table_2)
	account_prefix  = []string{"profile", "bag", "general", "pguild", "store", "simpleinfo", "friend", "anticheat", "tmp"}
)
