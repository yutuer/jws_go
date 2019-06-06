package account

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

type PlayerActGiftByTime struct {
	Gifts    []actGiftByCond `json:"gifts"`
	LastTime int64           `json:"last"`
}

func (p *PlayerActGiftByTime) OnAfterLogin() {
	if p.Gifts == nil {
		p.Gifts = make([]actGiftByCond, 0, gamedata.ActGiftInitLen)
	}
}

func (p *PlayerActGiftByTime) GetData(id int) *actGiftByCond {
	if id < 0 || id >= len(p.Gifts) {
		return nil
	}
	return &p.Gifts[id]
}

func (p *PlayerActGiftByTime) update(a *Account) {
	nowT := a.Profile.GetProfileNowTime()
	if !gamedata.IsSameDayCommon(nowT, p.LastTime) {
		// 全部重新计算
		p.Gifts = make([]actGiftByCond, 0, gamedata.ActGiftInitLen)
		p.LastTime = nowT
	}
}

func (p *PlayerActGiftByTime) GetAllInfo(a *Account) []byte {
	// 考虑到活动可能会过期与新增,这里要新建一套,最后替换回去
	newGifts := make([]actGiftByCond, 0, gamedata.ActGiftInitLen)

	p.update(a)

	data := gamedata.GetActivityGiftByTime()
	mainDatas := gamedata.GetActivityGiftByCond()

	resConds := actGiftByCondsToCLient{
		Gifts: make([][]byte, 0, gamedata.ActGiftInitLen),
	}

	if len(data) == 0 {
		return encode(resConds)
	}

	for _, m := range mainDatas {
		if m.ID == data[0].ID {
			resConds.Begin = m.TimeBegin
			resConds.End = m.TimeEnd
			resConds.Title = m.Title
			resConds.Desc = m.Desc
			resConds.Index = m.Index
			resConds.Typ = int(m.ID)
		}
	}

	for i := 0; i < len(data); i++ {
		g := data[i]
		giftTyp := int(resConds.Typ)

		currData := p.GetData(i)
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
			currData.Cond.Param1 = a.Profile.GetProfileNowTime()
			a.Profile.GetCondition().RegCondition(&currData.Cond)
		}

		progress, all := GetConditionProgress(
			&currData.Cond,
			a, g.Cond.Ctyp,
			g.Cond.Param1, g.Cond.Param2,
			g.Cond.Param3, g.Cond.Param4)

		// 写进新的地方
		newGifts = append(newGifts, *currData)

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

	p.Gifts = newGifts
	return encode(resConds)
}

func (p *PlayerActGiftByTime) RefreshRedPoint(a *Account) bool {
	for _, actData := range p.Gifts {
		if actData.IsHasGet != 1 {
			data := gamedata.GetActivityGiftByTime()
			progress, all := GetConditionProgress(
				&actData.Cond,
				a, data[actData.ID].Cond.Ctyp,
				data[actData.ID].Cond.Param1,
				data[actData.ID].Cond.Param2,
				data[actData.ID].Cond.Param3,
				data[actData.ID].Cond.Param4)
			if progress >= all { // 这里并没有加条件宽松，在领奖的地方有宽松条件
				return true
			}
		}
	}
	return false
}
