package logics

import (
	"time"

	"sort"

	"encoding/json"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/hero_diff"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/modules/worship"
	"vcs.taiyouxi.net/jws/gamex/modules/ws_pvp"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const MaxHeroIdxCount = 3

type PlayerInfoInRank struct {
	UID          string               `codec:"uid"`
	Name         string               `codec:"n"`
	CorpLv       int                  `codec:"cl"`
	Gs           int                  `codec:"gs"`
	Sorce        int64                `codec:"score"`
	Worship      int                  `codec:"worship"`
	MaxHeroIdx   [MaxHeroIdxCount]int `codec:"mhero"`
	MaxHeroStar  [MaxHeroIdxCount]int `codec:"mherostar"`
	GsHeroIds    []int                `codec:"gshero"`
	GsHeroGs     []int                `codec:"gsherogs"`
	GsHeroBaseGs []int                `codec:"gsherobgs"`
	Sid          string               `codec:"sid"`
	GuildName    string               `codec:"gname"`
}

type GuildInfoInRank struct {
	GuildID       int64  `codec:"id"`
	GuildUUID     string `codec:"uuid"`
	Name          string `codec:"n"`
	ChiefName     string `codec:"chief"`
	Lv            uint32 `codec:"lv"`
	Sorce         int64  `codec:"sorce"`
	GuildMemCount int    `codec:"memc"`
	GuildMemMax   int    `codec:"memmax"`
}

const (
	corpRankTyp_GS           = 0  // 请求战队总战力排行榜
	corpRankTyp_SimplePvp    = 1  // 请求竞技场积分排行榜
	corpRankTyp_Trial        = 2  // 请求最高爬塔榜
	corpRankTyp_Gs_SerOpen   = 3  // 请求战队总战力开服活动排行榜
	corpRankTyp_HeroStar     = 4  // 请求最强主将榜
	corpRankTyp_HeroDiffTU   = 5  // 请求出奇制胜屠榜单
	corpRankTyp_HeroDiffZHAN = 6  // 请求出奇制胜战榜单
	corpRankTyp_HeroDiffHU   = 7  // 请求出奇制胜屠护单
	corpRankTyp_HeroDiffSHI  = 8  // 请求出奇制胜士榜单
	corpRankTyp_WSPVP        = 9  // 无双争霸PVP
	corpRankType_WSPVP_BEST9 = 10 // 无双争霸9v9PVP
	corpRankType_Wei         = 11 // 魏国势力排行榜
	corpRankType_Shu         = 12 // 蜀国势力排行榜
	corpRankType_Wu          = 13 // 吴国势力排行榜
	corpRankType_QunXiong    = 14 // 群雄势力排行榜
)

const (
	guildRankTyp_Act         = 0 // 公会活跃度排行榜
	guildRankTyp_ActWeek     = 1 // 公会周活跃度排行榜
	guildRankTyp_ActWeekLast = 2 // 公会上周活跃度排行榜
	guildRankTyp_Gs          = 3 // 公会Gs排行榜
	guildRankTyp_GateEnemy   = 4 // 公会兵临城下排行榜
	guildRankTyp_Gs_SerOpen  = 5 // 公会Gs开服活动排行榜
)

type playerHeroStarMax struct {
	HeroIdx  [gamedata.AVATAR_NUM_MAX]int
	HeroStar [gamedata.AVATAR_NUM_MAX]int
	HeroCurr int
}

func (p *playerHeroStarMax) Init(stars []uint32) {
	for i := 0; i < len(stars); i++ {
		if stars[i] > 0 {
			p.HeroIdx[p.HeroCurr] = i
			p.HeroStar[p.HeroCurr] = int(stars[i])
			p.HeroCurr++
		}
	}
}

// Len is the number of elements in the collection.
func (g *playerHeroStarMax) Len() int {
	return g.HeroCurr
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (g *playerHeroStarMax) Less(i, j int) bool {
	datai := gamedata.GetHeroData(g.HeroIdx[i])
	dataj := gamedata.GetHeroData(g.HeroIdx[j])
	if datai == nil || dataj == nil {
		return false
	}
	if datai.RareLv > dataj.RareLv {
		return true
	}

	if datai.RareLv == dataj.RareLv {
		return g.HeroStar[i] > g.HeroStar[j]
	}

	return false
}

// Swap swaps the elements with indexes i and j.
func (g *playerHeroStarMax) Swap(i, j int) {
	if i == j {
		return
	}
	t := g.HeroStar[i]
	g.HeroStar[i] = g.HeroStar[j]
	g.HeroStar[j] = t

	t2 := g.HeroIdx[i]
	g.HeroIdx[i] = g.HeroIdx[j]
	g.HeroIdx[j] = t2
}

func (g *playerHeroStarMax) GetMaxHero() (idx, star [MaxHeroIdxCount]int) {
	sort.Sort(g)
	for i := 0; i < MaxHeroIdxCount; i++ {
		idx[i] = g.HeroIdx[i]
		star[i] = g.HeroStar[i]
	}
	return
}

func getMaxHeroStarFromSimpleInfo(a *helper.AccountSimpleInfo) (idx, star [MaxHeroIdxCount]int) {
	p := playerHeroStarMax{}
	p.Init(a.AvatarStarLvl[:])
	return p.GetMaxHero()
}

func getHeroDiffFreqHeroStarFromSimpleInfo(a *helper.AccountSimpleInfo, extraData []byte) (idx, star [MaxHeroIdxCount]int) {
	v := hero_diff.HeroDiffRankData{}
	err := json.Unmarshal(extraData, &v)
	if err != nil {
		logs.Error("fatal error for hero diff, extra data error by %v for byte: %v", err, string(extraData))
		idx = [MaxHeroIdxCount]int{0, 1, 2}
		return
	}
	return getHeroDiffFreqHeroStar(a, v.FreqAvatar)

}

func getHeroDiffFreqHeroStar(a *helper.AccountSimpleInfo, id []int) (idx, star [MaxHeroIdxCount]int) {
	for i, _ := range star {
		idx[i] = -1
	}
	for i := 0; i < MaxHeroIdxCount && i < len(id); i++ {
		idx[i] = id[i]
		star[i] = int(a.AvatarStarLvl[id[i]])
	}
	return
}

func (p *Account) GetRank(r servers.Request) *servers.Response {
	req := &struct {
		Req
		RankID int `codec:"id"`
	}{}
	resp := &struct {
		Resp
		Info                   [][]byte `codec:"topN"`
		Self                   []byte   `codec:"self"`
		PlayerRankPos          int      `codec:"pos"`
		PlayerRankPosYesterday int      `codec:"posy"`
	}{}

	initReqRsp(
		"PlayerAttr/PlayerRankRsp",
		r.RawBytes,
		req, resp, p)

	var res *rank.RankByCorpGetRes
	var res_yesterday *rank.RankByCorpGetRes
	acid := p.AccountID.String()

	simpleInfo := p.Account.GetSimpleInfo()

	switch req.RankID {
	case corpRankTyp_GS:
		res = rank.GetModule(p.AccountID.ShardId).RankCorpGs.Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankCorpGs.GetCorpInfo(&simpleInfo).Score
	case corpRankTyp_SimplePvp:
		res = rank.GetModule(p.AccountID.ShardId).RankSimplePvp.Get(acid)
		res.Score /= rank.SimplePvpScorePow
		for i := 0; i < len(res.TopN); i++ {
			res.TopN[i].Score /= rank.SimplePvpScorePow
		}
		p.Profile.GetFirstPassRank().OnRank(
			gamedata.FirstPassRankTypSimplePvp,
			int(res.Score))
	case corpRankTyp_Trial:
		res = rank.GetModule(p.AccountID.ShardId).RankByCorpTrial.Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByCorpTrial.GetCorpInfo(&simpleInfo).Score
	case corpRankTyp_Gs_SerOpen:
		res = rank.GetModule(p.AccountID.ShardId).RankCorpGsSevOpn.Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankCorpGsSevOpn.GetCorpInfo(&simpleInfo).Score
	case corpRankTyp_HeroStar:
		res = rank.GetModule(p.AccountID.ShardId).RankByHeroStar.Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByHeroStar.GetCorpInfo(&simpleInfo).Score
	case corpRankTyp_HeroDiffTU:
		res = rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[gamedata.HeroDiff_TU].Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[gamedata.HeroDiff_TU].GetCorpInfo(&simpleInfo).Score
	case corpRankTyp_HeroDiffZHAN:
		res = rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[gamedata.HeroDiff_ZHAN].Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[gamedata.HeroDiff_ZHAN].GetCorpInfo(&simpleInfo).Score
	case corpRankTyp_HeroDiffHU:
		res = rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[gamedata.HeroDiff_HU].Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[gamedata.HeroDiff_HU].GetCorpInfo(&simpleInfo).Score
	case corpRankTyp_HeroDiffSHI:
		res = rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[gamedata.HeroDiff_SHI].Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[gamedata.HeroDiff_SHI].GetCorpInfo(&simpleInfo).Score
	case corpRankTyp_WSPVP:
		res = p.getWspvpTopN()
	case corpRankType_WSPVP_BEST9:
		res = p.getWspvpBest9TopN()
	case corpRankType_Wei:
		res = rank.GetModule(p.AccountID.ShardId).RankByCorpGsOfWei.Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByCorpGsOfWei.GetCorpInfo(&simpleInfo).Score
	case corpRankType_Shu:
		res = rank.GetModule(p.AccountID.ShardId).RankByCorpGsOfShu.Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByCorpGsOfShu.GetCorpInfo(&simpleInfo).Score
	case corpRankType_Wu:
		res = rank.GetModule(p.AccountID.ShardId).RankByCorpGsOfWu.Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByCorpGsOfWu.GetCorpInfo(&simpleInfo).Score
	case corpRankType_QunXiong:
		res = rank.GetModule(p.AccountID.ShardId).RankByCorpGsOfQunXiong.Get(acid)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankByCorpGsOfQunXiong.GetCorpInfo(&simpleInfo).Score
	default:
		logs.SentryLogicCritical(acid, "GetRank Err by Unknown %d", req.RankID)
		return rpcError(resp, 1)
	}

	logs.Debug("<Get Rank> rankId=%d res=%v", req.RankID, res)

	var worshipDataMap map[string]int
	if req.RankID == corpRankTyp_GS {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		topWorship := worship.Get(p.AccountID.ShardId).GetWorshipData(ctx)
		if topWorship != nil {
			worshipDataMap = make(map[string]int, len(topWorship))
			for _, worshipData := range topWorship {
				worshipDataMap[worshipData.AccountID] = worshipData.WorshipTime
			}
		} else {
			logs.Warn("topWorship Get Err")
		}
	}
	var selfWorshipTime int
	if worshipDataMap != nil {
		w, ok := worshipDataMap[acid]
		if ok {
			selfWorshipTime = w
		}
	}
	selfDataInRank := PlayerInfoInRank{
		UID:       acid,
		Name:      p.Profile.Name,
		CorpLv:    int(p.GetCorpLv()),
		Gs:        simpleInfo.CurrCorpGs,
		Sorce:     res.Score,
		Worship:   selfWorshipTime,
		GuildName: p.GuildProfile.GuildName,
	}
	resp.PlayerRankPos = res.Pos

	resp.Info = make([][]byte, 0, rank.RankTopSizeToClient)
	for r := 0; r < len(res.TopN) && r < rank.RankTopSizeToClient; r++ {
		t := res.TopN[r]
		var worshipTime int
		if worshipDataMap != nil {
			w, ok := worshipDataMap[t.ID]
			if ok {
				worshipTime = w
			}
		}
		topPlayerInRank := PlayerInfoInRank{
			UID:          t.ID,
			Name:         t.Name,
			CorpLv:       t.CorpLv,
			Gs:           t.Gs,
			Sorce:        t.Score,
			Worship:      worshipTime,
			GsHeroIds:    t.Info.GsHeroIds,
			GsHeroGs:     t.Info.GsHeroGs,
			GsHeroBaseGs: t.Info.GsHeroBaseGs,
			GuildName:    t.Info.GuildName,
			Sid:          t.Sid,
		}
		if t.Score == 0 {
			topPlayerInRank = PlayerInfoInRank{}
		} else {
			if req.RankID == corpRankTyp_HeroStar {
				topPlayerInRank.MaxHeroIdx, topPlayerInRank.MaxHeroStar =
					getMaxHeroStarFromSimpleInfo(&t.Info)
			} else if req.RankID == corpRankTyp_HeroDiffTU ||
				req.RankID == corpRankTyp_HeroDiffZHAN ||
				req.RankID == corpRankTyp_HeroDiffHU ||
				req.RankID == corpRankTyp_HeroDiffSHI {
				topPlayerInRank.MaxHeroIdx, topPlayerInRank.MaxHeroStar =
					getHeroDiffFreqHeroStarFromSimpleInfo(&t.Info, t.ExtraData)
			}

		}
		if t.ID == p.AccountID.String() {
			selfDataInRank = PlayerInfoInRank{
				UID:       acid,
				Name:      p.Profile.Name,
				CorpLv:    t.CorpLv,
				Gs:        t.Gs,
				Sorce:     t.Score,
				Worship:   selfWorshipTime,
				GuildName: t.Info.GuildName,
			}
			resp.PlayerRankPos = r + 1

		}
		resp.Info = append(resp.Info, encode(topPlayerInRank))
	}
	if req.RankID == corpRankTyp_HeroStar {
		selfDataInRank.MaxHeroIdx, selfDataInRank.MaxHeroStar =
			getMaxHeroStarFromSimpleInfo(&simpleInfo)
	} else if req.RankID == corpRankTyp_HeroDiffTU ||
		req.RankID == corpRankTyp_HeroDiffZHAN ||
		req.RankID == corpRankTyp_HeroDiffHU ||
		req.RankID == corpRankTyp_HeroDiffSHI {
		heroDiff := p.Profile.GetHeroDiff()
		selfDataInRank.MaxHeroIdx, selfDataInRank.MaxHeroStar =
			getHeroDiffFreqHeroStar(&simpleInfo, heroDiff.GetTopNFreqHero(getHeroDiffRankTyp(req.RankID), gamedata.HeroDiffRankShowAvatarCount))
	} else if req.RankID == corpRankType_WSPVP_BEST9 {
		if resp.PlayerRankPos > len(res.TopN) || resp.PlayerRankPos == 0 {
			resp.PlayerRankPos = 0 // 如果这个人不在排行榜，排名改成0
			selfDataInRank.Gs = int(simpleInfo.WuShuangGs)
			//logs.Debug("9BestGs : %d", selfDataInRank.Gs)
		}

	}
	logs.Debug("self data in rank: %v", selfDataInRank)
	resp.Self = encode(selfDataInRank)

	if res_yesterday != nil {
		resp.PlayerRankPosYesterday = res_yesterday.Pos
	}

	return rpcSuccess(resp)
}

func (p *Account) getWspvpTopN() *rank.RankByCorpGetRes {
	res := new(rank.RankByCorpGetRes)
	wspvpRanks := ws_pvp.GetModule(p.AccountID.ShardId).GetTopN()
	res.TopN = make([]rank.CorpDataInRank, len(wspvpRanks))
	for i, rankPlayer := range wspvpRanks {
		res.TopN[i] = rank.CorpDataInRank{
			Name:   rankPlayer.Name,
			CorpLv: rankPlayer.CorpLevel,
			Score:  int64(rankPlayer.Rank),
			Info: helper.AccountSimpleInfo{
				GuildName: rankPlayer.GuildName,
			},
			ID:  rankPlayer.Acid,
			Sid: rankPlayer.SidStr,
		}
	}
	res.Pos = p.Profile.WSPVPPersonalInfo.Rank
	res.Score = int64(res.Pos)
	return res
}

func (p *Account) getWspvpBest9TopN() *rank.RankByCorpGetRes {
	res := new(rank.RankByCorpGetRes)
	wspvpRanks := ws_pvp.GetModule(p.AccountID.ShardId).GetBest9TopN()
	res.TopN = make([]rank.CorpDataInRank, len(wspvpRanks))
	for i, rankPlayer := range wspvpRanks {
		res.TopN[i] = rank.CorpDataInRank{
			Name:   rankPlayer.Name,
			CorpLv: rankPlayer.CorpLevel,
			Score:  int64(rankPlayer.Rank),
			Info: helper.AccountSimpleInfo{
				GuildName: rankPlayer.GuildName,
			},
			ID:  rankPlayer.Acid,
			Sid: rankPlayer.SidStr,
			Gs:  int(rankPlayer.Gs),
		}
	}
	res.Pos = ws_pvp.GetBest9RankByOne(p.GetWSPVPGroupId(), p.AccountID.String())
	res.Score = int64(res.Pos)
	return res
}

func getHeroDiffRankTyp(rankTyp int) int {
	switch rankTyp {
	case corpRankTyp_HeroDiffTU:
		return gamedata.HeroDiff_TU
	case corpRankTyp_HeroDiffZHAN:
		return gamedata.HeroDiff_ZHAN
	case corpRankTyp_HeroDiffHU:
		return gamedata.HeroDiff_HU
	case corpRankTyp_HeroDiffSHI:
		return gamedata.HeroDiff_SHI
	default:
		return gamedata.HeroDiff_TU
	}
}

func (p *Account) GetSomeoneFromGSRank(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Num int `codec:"num"`
	}{}
	resp := &struct {
		SyncResp
		Info [][]byte `codec:"info"`
	}{}

	initReqRsp(
		"Attr/GetFromGSRankRsp",
		r.RawBytes,
		req, resp, p)

	acid := p.AccountID.String()
	res := rank.GetModule(p.AccountID.ShardId).RankCorpGs.Get(acid)

	var topNCount int
	for _, t := range res.TopN {
		if t.Name != "" {
			topNCount++
		}
	}

	// 注释下面这段代码，让查找范围是整个top
	//if topNCount > rank.GSRankRandTopSizeToClient {
	//	topNCount = rank.GSRankRandTopSizeToClient
	//}

	resp.Info = make([][]byte, 0, rank.GSRankRandTopSizeToClient)
	randArray := util.Shuffle1ToNSelf(
		topNCount,
		p.GetRand())

	for i := 0; i < topNCount && i < req.Num && i < rank.GSRankRandTopSizeToClient; i++ {
		idx := randArray[i]
		if res.TopN[idx].Info.CurrCorpGs <= 0 {
			logs.Error("GetSomeoneFromGSRank gs <= 0 %v", res.TopN[idx].Info)
			continue
		}
		resp.Info = append(resp.Info, encode(res.TopN[idx].Info))
	}

	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) GetGuildRank(r servers.Request) *servers.Response {
	req := &struct {
		Req
		RankID int `codec:"id"`
	}{}
	resp := &struct {
		Resp
		Info  [][]byte `codec:"topN"`
		Pos   int      `codec:"pos"`
		Score int64    `codec:"s"`
	}{}

	initReqRsp(
		"PlayerGuild/PlayerGuildRankRsp",
		r.RawBytes,
		req, resp, p)

	guildUUID := p.GuildProfile.GuildUUID
	var guildInfo *guild.GuildInfo
	var gInfo *guild_info.GuildInfoBase
	var re guild.GuildRet
	if guildUUID != "" {
		guildInfo, re = guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
		if re.HasError() {
			logs.Error("GetGuildRank GetGuildInfo err %v", re)
			guildInfo = nil
		} else {
			gInfo = &guildInfo.GuildInfoBase
		}
	}
	var sinfo *guild_info.GuildSimpleInfo
	if gInfo != nil {
		sinfo = &gInfo.Base
	}

	var res *rank.RankByGuildGetRes
	switch req.RankID {
	//case guildRankTyp_Act:
	//	res = rank.RankGuildAct.Get(p.GuildProfile.GuildUUID)
	//	res.Score = rank.RankGuildAct.GetInfoInRank(gInfo).Score
	//case guildRankTyp_ActWeek:
	//	res = rank.RankGuildActWeek.Get(p.GuildProfile.GuildUUID)
	//	res.Score = rank.RankGuildActWeek.GetInfoInRank(gInfo).Score
	//case guildRankTyp_ActWeekLast:
	//	res = rank.RankGuildActWeekLast.Get(p.GuildProfile.GuildUUID)
	//	res.Score = rank.RankGuildActWeekLast.GetInfoInRank(gInfo).Score
	case guildRankTyp_Gs:
		res = rank.GetModule(p.AccountID.ShardId).RankGuildGs.Get(p.GuildProfile.GuildUUID)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankGuildGs.GetInfoInRank(sinfo).Score
	case guildRankTyp_GateEnemy:
		res = rank.GetModule(p.AccountID.ShardId).RankGuildGateEnemy.Get(p.GuildProfile.GuildUUID)
		res.Score = rank.GetModule(p.AccountID.ShardId).RankGuildGateEnemy.GetInfoInRank(sinfo).Score
	case guildRankTyp_Gs_SerOpen:
		res = rank.GetModule(p.AccountID.ShardId).RankGuildSevOpn.Get(p.GuildProfile.GuildUUID)
	}

	s := rank.RankTopSizeToClient
	if req.RankID == guildRankTyp_Gs_SerOpen {
		s = gamedata.GetGuildRankCount()
	}

	resp.Info = make([][]byte, 0, s)
	for _, t := range res.TopN[:s] {
		resp.Info = append(resp.Info, encode(GuildInfoInRankForClient{
			GuildID:       guild.GuildItoa(t.ID),
			GuildUUID:     t.UUID,
			Name:          t.Name,
			Lv:            t.Lv,
			Sorce:         t.Score,
			ChiefName:     t.ChiefName,
			GuildMemMax:   t.GuildMemMax,
			GuildMemCount: t.GuildMemCount,
		}))
	}

	resp.Pos = res.Pos
	resp.Score = res.Score

	return rpcSuccess(resp)
}

type GuildInfoInRankForClient struct {
	GuildID       string `codec:"id"`
	GuildUUID     string `codec:"uuid"`
	Name          string `codec:"n"`
	ChiefName     string `codec:"chief"`
	Lv            uint32 `codec:"lv"`
	Sorce         int64  `codec:"sorce"`
	GuildMemCount int    `codec:"memc"`
	GuildMemMax   int    `codec:"memmax"`
}
