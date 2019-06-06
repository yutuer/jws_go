package helper

const MatchPlayerNum = 3
const MatchMinPlayerNum = 1
const RegTimePreSeconds = 10
const MatchDefaultToken = "DefaultToken"

//const MatchPostUrlAddress = "/api/v1/match"
const MatchPostUrlAddressV2 = "/api/v2/match"
const FenghuoPostUrlAddressV1 = "/api/v1/create"
const TeamBossPostUrlAddressV1 = "/api/v1/tb_create"

//const MatchCancelPostUrlAddress = "/api/v1/matchCancel"
const OnMatchSuccessPostUrl = "/api/v1/matchSucess"
const OnFenghuoSuccessPostUrl = "/api/v1/fenghuoSucess"

const OnTBSuccessPostUrl = "/api/v1/tb_success"

const OnGVGSuccessPostUrl = "/api/v1/gvg_success"

const TeamBossToken = "TeamBoss"

const GVGToken = "GVG"

func FmtMatchToken(token string) string {
	return "Match_" + token
}
