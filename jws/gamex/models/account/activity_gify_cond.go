package account

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

type actGiftByCond struct {
	ID       int                `json:"id"`
	Typ      int                `json:"typ"`
	Cond     gamedata.Condition `json:"cond"`
	IsHasGet int                `json:"isget"`
}

func (a *actGiftByCond) SetHasGet() {
	a.IsHasGet = 1
}

type actGiftByCondToCLient struct {
	ID           int      `codec:"id"`
	Typ          int      `codec:"typ"`
	IsHasGet     int      `codec:"hasget"`
	CondTyp      int      `codec:"cond_t"`
	CondPrograss int      `codec:"cond_p"`
	CondAll      int      `codec:"cond_a"`
	RewardIDs    []string `codec:"rid"`
	RewardCounts []uint32 `codec:"rc"`
	Desc         string   `codec:"desc"`
}

type actGiftByCondsToCLient struct {
	Gifts [][]byte `codec:"gifts_"`
	Begin int64    `codec:"begin"`
	End   int64    `codec:"end"`
	Title string   `codec:"title"`
	Desc  string   `codec:"desc"`
	Typ   int      `codec:"typ_id"`
	Index int      `codec:"index"`
}

type PlayerActGiftByConds struct {
	Gifts   []actGiftByCond `json:"gifts"`
	giftMap map[int][]*actGiftByCond
}

func (p *PlayerActGiftByConds) addByGifts(g *actGiftByCond) {
	gfits, ok := p.giftMap[g.Typ]
	if !ok || gfits == nil {
		p.giftMap[g.Typ] = make([]*actGiftByCond, 0, gamedata.ActGiftInitLen)[:]
	}
	gfits = p.giftMap[g.Typ]
	for g.ID >= len(gfits) {
		gfits = append(gfits, nil)
	}
	gfits[g.ID] = g
	p.giftMap[g.Typ] = gfits
	return
}

func (p *PlayerActGiftByConds) OnAfterLogin() {
	if p.Gifts == nil {
		p.Gifts = make([]actGiftByCond, 0, gamedata.ActGiftInitLen*gamedata.ActTypInitLen)
	}
}

func (p *PlayerActGiftByConds) initMap() {
	p.giftMap = make(map[int][]*actGiftByCond, gamedata.ActTypInitLen)
	for idx, _ := range p.Gifts {
		p.addByGifts(&p.Gifts[idx])
	}
}

func (p *PlayerActGiftByConds) GetData(id, typ int) *actGiftByCond {
	if p.giftMap == nil {
		p.initMap()
	}
	t, tOk := p.giftMap[typ]
	if !tOk || t == nil {
		return nil
	}
	if id < 0 || id >= len(t) {
		return nil
	}
	return t[id]
}

func (p *PlayerActGiftByConds) GetAllInfo(a *Account) [][]byte {
	// 考虑到活动可能会过期与新增,这里要新建一套,最后替换回去
	newGifts := make([]actGiftByCond, 0, gamedata.ActGiftInitLen*gamedata.ActTypInitLen)

	res := make([][]byte, 0, gamedata.ActTypInitLen)
	data := gamedata.GetActivityGiftByCond()
	for _, gifts := range data {
		// 旗下没有Gift不添加, 同时排出了time的奖励
		if len(gifts.Gifts) == 0 {
			continue
		}

		giftTyp := int(gifts.ID)
		resConds := actGiftByCondsToCLient{
			Gifts: make([][]byte, 0, gamedata.ActGiftInitLen),
			Begin: gifts.TimeBegin,
			End:   gifts.TimeEnd,
			Title: gifts.Title,
			Desc:  gifts.Desc,
			Typ:   giftTyp,
			Index: gifts.Index,
		}

		for i := 0; i < len(gifts.Gifts); i++ {
			g := gifts.Gifts[i]

			currData := p.GetData(i, giftTyp)
			if currData == nil {
				cond := NewCondition(
					g.Cond.Ctyp,
					g.Cond.Param1,
					g.Cond.Param2,
					g.Cond.Param3,
					g.Cond.Param4)
				p.Gifts = append(p.Gifts, actGiftByCond{
					ID:   i,
					Typ:  giftTyp,
					Cond: *cond,
				})
				currData = &p.Gifts[len(p.Gifts)-1]
				p.addByGifts(currData)
				a.Profile.GetCondition().RegCondition(&currData.Cond)
			}

			// 写进新的地方
			newGifts = append(newGifts, *currData)

			progress, all := GetConditionProgress(
				&currData.Cond,
				a, g.Cond.Ctyp,
				g.Cond.Param1, g.Cond.Param2,
				g.Cond.Param3, g.Cond.Param4)
			resConds.Gifts = append(resConds.Gifts, encode(actGiftByCondToCLient{
				ID:           i,
				Typ:          giftTyp,
				IsHasGet:     currData.IsHasGet,
				CondTyp:      int(currData.Cond.Ctyp),
				CondPrograss: progress,
				CondAll:      all,
				RewardIDs:    g.Reward.Ids,
				RewardCounts: g.Reward.Counts,
				Desc:         g.Desc,
			}))
		}

		res = append(res, encode(resConds))
	}
	p.Gifts = newGifts
	p.initMap()
	return res
}

func (p *PlayerActGiftByConds) RefreshRedPoint(a *Account) bool {
	nowT := a.Profile.GetProfileNowTime()
	for _, actData := range p.Gifts {
		if actData.IsHasGet != 1 {
			data := gamedata.GetActivityGiftByCondData(uint32(actData.ID), uint32(actData.Typ), nowT)
			if data == nil {
				return false
			}
			progress, all := GetConditionProgress(
				&actData.Cond,
				a, data.Cond.Ctyp,
				data.Cond.Param1, data.Cond.Param2,
				data.Cond.Param3, data.Cond.Param4)

			if progress >= all {
				return true
			}
		}
	}
	return false
}
