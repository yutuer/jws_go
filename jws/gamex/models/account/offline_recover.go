package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

// 离线找回资源功能
type OfflineRecoverInfo struct {
	Resources        []OfflineResource `json:"resources"`
	LastRecordTime   int64             `json:"last_record_time"`    // 记录这个时间， 以防万一
	LastClaimAllTime int64             `json:"last_claim_all_time"` // 记录每日全部领取后的时间，用于给客户端显示
}

type OfflineResource struct {
	ScId        string `json:"sc_id"`    // 资源ID
	OfflineDays int    `json:"off_days"` // 已经累计未领奖的天数
}

func (info *OfflineRecoverInfo) GetResource(id string) *OfflineResource {
	for i, res := range info.Resources {
		if res.ScId == id {
			return &info.Resources[i]
		}
	}
	return nil
}

func (info *OfflineRecoverInfo) HasRewards() bool {
	for _, res := range info.Resources {
		if res.OfflineDays > 0 {
			return true
		}
	}
	return false
}

// 客户端当日领完奖励后需要还能显示这个页签
func (info *OfflineRecoverInfo) IsClientShow(nowTime int64) bool {
	return gamedata.IsSameDayCommon(nowTime, info.LastClaimAllTime)
}

func (info *OfflineRecoverInfo) OnAfterLogin(lastLogoutTime, nowTime int64) {
	info.tryInitOrUpdate()
	if lastLogoutTime == 0 {
		// 没有登出记录的不处理, 一般是新建的角色
		return
	}
	startTime := lastLogoutTime
	if info.LastRecordTime > lastLogoutTime {
		startTime = lastLogoutTime // 理论上这种情况并不会发生
	}
	info.LastRecordTime = nowTime
	intervalDay := gamedata.GetIntervalDayByCommon(startTime, nowTime)
	maxDay := int(gamedata.GetCommonCfg().GetRecoverDayLimit())
	for i, res := range info.Resources {
		if (res.OfflineDays + intervalDay) > maxDay {
			info.Resources[i].OfflineDays = maxDay
		} else {
			info.Resources[i].OfflineDays += intervalDay
		}
	}
}

func (info *OfflineRecoverInfo) tryInitOrUpdate() {
	if info.Resources == nil {
		info.Resources = getInitResources()
	} else {
		// 表有增加或者删除的时候， 直接重新设置
		newResource := getInitResources()
		for i, newRes := range newResource {
			for _, oldRes := range info.Resources {
				if newRes.ScId == oldRes.ScId {
					newResource[i].OfflineDays = oldRes.OfflineDays
				}
			}
		}
		info.Resources = newResource
	}
}

func getInitResources() []OfflineResource {
	res := make([]OfflineResource, 0)
	configs := gamedata.GetAllOfflineRecoverConfigs()
	for _, cfg := range configs {
		res = append(res, OfflineResource{
			ScId: cfg.GetResourcesID(),
		})
	}
	return res
}
