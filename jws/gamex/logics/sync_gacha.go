package logics

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

type GachaToClient struct {
	GachaCategory     uint32 `codec"gc"`

	TicketTyp         string `codec:"t_t"`
	TicketCount       uint32 `codec:"t_c"`

	TicketTenTyp      string `codec:"tt_t"`
	TicketTenCount    uint32 `codec:"tt_c"`
	PriceTyp          string `codec:"p_t"`
	PriceCount        uint32 `codec:"p_c"`
	PriceTenTyp       string `codec:"pt_t"`
	PriceTenCount     uint32 `codec:"pt_c"`
	IsFree            bool   `codec:"is_f"`
	FreeGachaTime     int    `codec:"f_c"`
	FreeGachaAllTime  int    `codec:"allf_c"`
	FreeNext          int64  `codec:"n"`
	SerialRewardId    string `codec:"s_t"`
	SerialRewardCount uint32 `codec:"s_c"`
	ExtNeedCount      int64  `codec:"extc"`
}

func (s *SyncResp) mkGachaAllInfo(p *Account) {
	if s.gacha_all_sync {
		now_time := p.Profile.GetProfileNowTime()
		p.Profile.Gacha.Update(now_time)
		gacha_len := len(p.Profile.Gacha.Gacha)
		s.SyncGacha = make([][]byte, gacha_len, gacha_len)
		for i := 0; i < len(s.SyncGacha); i++ {
			s.SyncGacha[i] = []byte{}
		}

		corp_lv, _ := p.Profile.GetCorp().GetXpInfo()

		for idx, gacha := range p.Profile.Gacha.Gacha {
			data := gamedata.GetGachaData(corp_lv, idx)

			if data == nil {
				//				logs.Error("GetGachaData Err By %d", idx)
				continue
			}

			if data.GachaID == 0 {
				continue
			}

			sid, sc := gacha.GetSerialRewardInfo(corp_lv, idx, p.Profile.GetCurrAvatar())
			etc := gacha.GetRewardExtRewardCount(corp_lv, idx)
			if data.ExtraSpace > 0 && etc >= 0 {
				etc = int64(data.ExtraSpace) - (gacha.GetRewardExtRewardCount(corp_lv, idx) % int64(data.ExtraSpace))
			} else {
				etc = 0
			}
			s.SyncGacha[idx] = encode(GachaToClient{
				GachaCategory:     data.GachaCategory,
				TicketTyp:         data.CostForOne_TTyp,
				TicketCount:       data.CostForOne_TCount,
				TicketTenTyp:      data.CostForTen_TTyp,
				TicketTenCount:    data.CostForTen_TCount,
				PriceTyp:          data.CostForOne_Typ,
				PriceCount:        data.CostForOne_Count,
				PriceTenTyp:       data.CostForTen_Typ,
				PriceTenCount:     data.CostForTen_Count,
				IsFree:            gacha.IsCanFree(corp_lv, now_time, idx),
				FreeGachaTime:     data.FreeCountEveryOneDay - gacha.TodayFreeCount,
				FreeGachaAllTime:  data.FreeCountEveryOneDay,
				FreeNext:          data.FreeCoolTime + gacha.LastFreeTime,
				SerialRewardId:    sid,
				SerialRewardCount: sc,
				ExtNeedCount:      etc,
			})
		}
	}
}
