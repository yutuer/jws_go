package logics

import "vcs.taiyouxi.net/jws/gamex/logiclog"

func (s *SyncResp) mkFashionAllInfo(p *Account) {
	acid := p.AccountID.String()

	//
	// 每次刷新时装，删除过期的
	// 要在刷新GS之前做，因为时装会过期，并也会影响GS计算，所以先刷新一下
	//
	p.FashionRefresh(s)

	if s.SyncFashionBagAllNeed {
		all := p.Profile.GetFashionBag().GetFashionAll()
		s.SyncFashionBagAll = make([][]byte, 0, len(all))
		for _, f := range all {
			s.SyncFashionBagAll = append(s.SyncFashionBagAll, encode(f))
		}
	}

	if s.SyncFashionBagUpdateNeed {
		fashions := p.Profile.GetFashionBag().GetFashions2Client(s.fashion_update)
		s.SyncFashionBagUpdate = make([][]byte, 0, len(fashions))
		for _, f := range fashions {
			s.SyncFashionBagUpdate = append(s.SyncFashionBagUpdate, encode(f))
		}
		// logiclog
		addFashions := make(map[string]int64, len(s.fashion_update_o))
		for _, f := range fashions {
			if _, ok := s.fashion_update_o[f.ID]; ok {
				addFashions[f.TableID] = 1
				logiclog.LogGiveItemUseSelf(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
					p.Profile.ChannelId, s.items_update_reason, f.TableID, 0, 1, 0,
					p.Profile.GetVipLevel(), func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			}
		}
		if len(addFashions) > 0 {
			logiclog.LogGiveItem(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
				p.Profile.ChannelId, s.fashion_update_reason, addFashions, p.Profile.GetVipLevel(),
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
		}
	}

	if s.SyncFashionBagDelNeed {
		s.SyncFashionBagDel = s.fashion_del
		// logiclog
		logiclog.LogCostItem(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId, s.fashion_del_reason, s.fashion_del_o, p.Profile.GetVipLevel(),
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	}
}
