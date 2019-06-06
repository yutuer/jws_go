package rank

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

const RankTopSize = 100
const RankBalanceSize = 1000
const RankTopSizeToClient = 100
const GSRankRandTopSizeToClient = 20
const GetChannelSize = 2048
const AddChannelSize = 2048

const SimplePvpScorePow = 1000
const SimplePvpInitScoreReal = 2000
const SimplePvpInitScore = SimplePvpInitScoreReal * SimplePvpScorePow

const CleanSize int64 = 1000
const RankScorePowBase = 100000

const AVATAR_NUM_MAX = helper.AVATAR_NUM_MAX
const AVATAR_NUM_CURR = helper.AVATAR_NUM_CURR
const AVATAR_SLOT_MAX = helper.EQUIP_SLOT_MAX
const AVATAR_SKILL_MAX = helper.AVATAR_SKILL_MAX

const rank_mail_title = "[TODO]rank_mail_title"
const rank_mail_info = "[TODO]rank_mail_info"

type PairPosScore struct {
	Pos   int
	Score int64
}
