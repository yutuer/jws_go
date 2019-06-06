package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

func (s *SyncResp) mkJadeAllInfo(p *Account) {
	acid := p.AccountID.String()

	if s.SyncFullJadeNeed {
		js := p.Profile.GetJadeBag().GetFullJade2Client()
		s.SyncFullJade = make([][]byte, 0, len(js))
		for _, j := range js {
			s.SyncFullJade = append(s.SyncFullJade, encode(j))
		}
	}

	if s.SyncUpdateJadeNeed {
		jades := p.Profile.GetJadeBag().GetJades2Client(s.jades_update)
		s.SyncUpdateJade = make([][]byte, 0, len(jades))
		for _, j := range jades {
			s.SyncUpdateJade = append(s.SyncUpdateJade, encode(j))
		}
		// logiclog
		addjades := make(map[string]int64, len(s.jades_update_o))
		deljades := make(map[string]int64, len(s.jades_update_o))
		for idx, oldCount := range s.jades_update_o {
			jade := jades[idx]
			newCount := jade.Count
			if newCount > oldCount {
				addjades[jade.TableID] = newCount - oldCount
				logiclog.LogGiveItemUseSelf(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
					p.Profile.ChannelId, s.items_update_reason, jade.TableID, oldCount, newCount-oldCount, newCount,
					p.Profile.GetVipLevel(), func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			} else if newCount < oldCount {
				deljades[jade.TableID] = oldCount - newCount
				logiclog.LogCostItemUseSelf(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
					p.Profile.ChannelId, s.items_update_reason, jade.TableID, oldCount, newCount-oldCount, newCount,
					p.Profile.GetVipLevel(), func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			}
		}
		if len(addjades) > 0 {
			logiclog.LogGiveItem(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
				p.Profile.ChannelId, s.jades_update_reason, addjades, p.Profile.GetVipLevel(),
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

		}
		if len(deljades) > 0 {
			logiclog.LogCostItem(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
				p.Profile.ChannelId, s.jades_update_reason, deljades, p.Profile.GetVipLevel(),
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
		}
	}

	if s.SyncDelJadeNeed {
		s.SyncDelJade = s.jades_del
		// logiclog
		logiclog.LogCostItem(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId, s.jades_del_reason, s.jades_del_o, p.Profile.GetVipLevel(),
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	}

	if s.avatar_jade_need_sync {
		s.JadeSlotMax = gamedata.JadePartCount
		ids, jades := p.Profile.GetEquipJades().CurrAll2Client()
		s.AvatarIdJade = ids
		s.AvatarJade = make([][]byte, 0, len(jades))
		for _, j := range jades {
			s.AvatarJade = append(s.AvatarJade, encode(j))
		}
	}

	if s.destinyGen_jade_need_sync {
		s.JadeSlotMax = gamedata.JadePartCount
		ids, jades := p.Profile.GetDestGeneralJades().CurrAll2Client()
		s.DGIdJade = ids
		s.DGJade = make([][]byte, 0, len(jades))
		for _, j := range jades {
			s.DGJade = append(s.DGJade, encode(j))
		}
	}
}
