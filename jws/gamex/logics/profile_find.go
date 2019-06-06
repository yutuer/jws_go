package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/update/data_update"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) GetOtherAccountData(r servers.Request) *servers.Response {
	req := &struct {
		Req
		UID       string `codec:"uid"`
		CountryId int64  `codec:"country_id"` // 势力排行榜使用
	}{}

	resp := &struct {
		SyncResp
		Avatar    []byte `codec:"avatar"`
		CountryId int64  `codec:"country_id"` // 势力排行榜使用
	}{}

	initReqRsp(
		"Attr/GetOtherAccountRsp",
		r.RawBytes,
		req, resp, p)

	acid := p.AccountID.String()

	otherAcID := req.UID
	dbAccountID, err := db.ParseAccount(otherAcID)
	if err != nil {
		return rpcErrorWithMsg(resp, 6, fmt.Sprintf("ERR_Param %s", otherAcID))
	}
	otherAccount, err := account.LoadPvPAccount(dbAccountID)
	if err != nil {
		logs.SentryLogicCritical(acid, "LoadAccount %s Err By %s",
			dbAccountID, err.Error())

		return rpcError(resp, 1)
	}

	// 数据更新不涉及结构变化
	// 此处处理数据更新, version最后的更新也是在这里做的
	err = data_update.Update(otherAccount.Profile.Ver, true, otherAccount)
	if err != nil {
		logs.SentryLogicCritical(otherAccount.AccountID.String(),
			"data_update Err By %s",
			err.Error())
		return rpcError(resp, 1)
	}

	a := helper.Avatar2Client{}
	err = account.FromAccount(&a, otherAccount, otherAccount.Profile.CurrAvatar)
	if err != nil {
		logs.SentryLogicCritical(p.AccountID.String(),
			"FromAccount Err by %s",
			err.Error())
		return rpcError(resp, 2)
	}

	resPos, resScore := rank.GetModule(p.AccountID.ShardId).RankSimplePvp.GetPos(otherAcID)
	resScore /= rank.SimplePvpScorePow
	a.SimplePvpScore = resScore
	a.SimplePvpRank = resPos
	a.Gs = GetCurrGS(otherAccount)
	if a.GuildUUID == p.GuildProfile.GuildUUID {
		a.GuildName = p.GuildProfile.GuildName
	}
	countryId := int(req.CountryId)
	if req.CountryId != 0 {
		addCountryInfo(&a, otherAccount, countryId)
		resp.CountryId = req.CountryId
	}
	resp.Avatar = encode(a)

	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func addCountryInfo(a *helper.Avatar2Client, acc *account.Account, countryId int) {
	if countryId <= helper.Country_Invalid || countryId >= helper.Country_Count {
		return
	}
	bestHeroInfoByCountry := acc.GetBestHeroByCountry()
	a.ShowCountry = countryId
	a.HeroIdsByCountry = make([]int, 0)
	a.HeroGsByCountry = make([]int, 0)
	a.HeroBaseGsByCountry = make([]int, 0)
	for _, info := range bestHeroInfoByCountry[countryId] {
		a.HeroIdsByCountry = append(a.HeroIdsByCountry, info.HeroId)
		a.HeroGsByCountry = append(a.HeroGsByCountry, info.HeroGs)
		a.HeroBaseGsByCountry = append(a.HeroBaseGsByCountry, info.HeroBaseGs)
	}
}
