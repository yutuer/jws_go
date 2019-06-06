package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ExchangeHeroPiece : 请求兑换武将碎片
// 武将碎片兑换令牌
func (p *Account) ExchangeHeroPieceHandler(req *reqMsgExchangeHeroPiece, resp *rspMsgExchangeHeroPiece) uint32 {
	heroId := req.HeroAvatar
	starLevel := p.Profile.GetHero().HeroStarLevel[heroId]
	if starLevel < 25 {
		logs.Warn("<ExchangeHeroPieceHandler> bad hero level %d", starLevel)
		return errCode.CommonConditionFalse
	}
	heroCfg := gamedata.GetHeroData(int(heroId))
	if heroCfg == nil {
		logs.Warn("<ExchangeHeroPieceHandler> bad hero id %d", heroId)
		return errCode.CommonInvalidParam
	}
	pieceCountId := heroCfg.Piece
	pieceCostCount := heroCfg.SurplusCurrencyCount
	scGotId := heroCfg.SurplusCurrencyId
	scGotCount := heroCfg.SurplusCurrencyCount2

	if req.IsTen {
		realCount := p.getRealExchangeHeroPieceCount(int(heroId), pieceCostCount)
		if realCount <= 0 {
			return errCode.CommonCountLimit
		}
		pieceCostCount = pieceCostCount * realCount
		scGotCount = scGotCount * realCount
	}

	costData := &gamedata.CostData{}
	costData.AddItem(pieceCountId, pieceCostCount)

	reason := fmt.Sprintf("exchange hero piece %d", heroId)
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}

	giveData := &gamedata.CostData{}
	giveData.AddItem(scGotId, uint32(scGotCount))
	if ok := account.GiveBySync(p.Account, giveData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}

	resp.mkInfo(p)
	return 0
}

func (p *Account) getRealExchangeHeroPieceCount(avatarId int, costUnitCount uint32) uint32 {
	maxCount := p.Profile.GetHero().HeroStarPiece[avatarId] / costUnitCount
	if maxCount < 10 {
		return maxCount
	} else {
		return 10
	}
}

// DrawHeroPieceGacha : 武将碎片抽奖
// 武将碎片抽奖
func (p *Account) DrawHeroPieceGachaHandler(req *reqMsgDrawHeroPieceGacha, resp *rspMsgDrawHeroPieceGacha) uint32 {

	return 0
}
