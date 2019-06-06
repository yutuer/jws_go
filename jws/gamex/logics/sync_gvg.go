package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GVGCityData struct {
	CityID    int    `codec:"GVGCityID"`
	LeadGuild string `codec:"GVGCityOwnedGuildName"`
}

func (s *SyncResp) mkGVGInfo(p *Account) {
	t, _ := gamedata.GetHotDatas().GvgConfig.GetGVGTime(p.AccountID.ShardId,
		gvg.GetModule(p.AccountID.ShardId).GetNowTime())
	s.SyncGVGTime = encode(t)
	if s.SyncGVGNeed {
		ret := gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
			Typ: gvg.Cmd_Typ_Get_CityLeader,
		})
		if ret.SortItem == nil {
			logs.Error("Fatal Error No GVG City Leader Data")
			return
		}
		s.SyncGVGCityData = make([][]byte, 0, len(ret.SortItem))
		for _, item := range ret.SortItem {
			data := GVGCityData{
				CityID:    item.IntKey,
				LeadGuild: item.StrVal,
			}
			s.SyncGVGCityData = append(s.SyncGVGCityData, encode(data))
		}
	}
}
