package guild

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	RP_IPA_NONE = iota
	RP_IPA_HAS_PAY
	RP_IPA_HAS_CLAIM
)

type GuildRedPacketInfo struct {
	IpaStatus      int64              `json:"grp_ipa"`   // 状态
	TodayGrabList  []string           `json:"grp_grab"`  // 今日抢的红包ID
	TodayClaimList []int64            `json:"grp_claim"` // 今日临朐宝箱的进度
	GrabLogList    []GrabRedPacketLog `json:"grp_log"`   // 今日抢红包的记录
	LastResetTime  int64              `json:"grp_reset"` // 上次重置时间
	Sync           helper.SyncObj
}

type GrabRedPacketLog struct {
	RedPacketId string       `json:"grp_rp_id"`
	SenderName  string       `json:"grp_send_name"`
	ItemList    []RewardItem `json:"grp_reward"`
}

type RewardItem struct {
	Id    string `json:"grp_item_id"`
	Count int64  `json:"grp_item_count"`
}

func (rp *GuildRedPacketInfo) CheckDailyReset(now int64) bool {
	logs.Debug("guild red packet CheckDailyReset, %d, %d", now, rp.LastResetTime)
	if !gamedata.IsSameDayCommon(rp.LastResetTime, now) {
		rp.DailyReset(now)
		logs.Debug("redpacket daily reset")
		return true
	}
	return false
}

func (rp *GuildRedPacketInfo) DailyReset(now int64) {
	rp.IpaStatus = RP_IPA_NONE
	rp.TodayClaimList = nil
	rp.TodayGrabList = nil
	rp.GrabLogList = nil
	rp.LastResetTime = now
}

func (rp *GuildRedPacketInfo) IsClaimed(boxId int64) bool {
	if rp.TodayClaimList == nil {
		return false
	}
	for _, id := range rp.TodayClaimList {
		if id == boxId {
			return true
		}
	}
	return false
}

func (rp *GuildRedPacketInfo) OnGrab(rpId, senderName string, items map[string]uint32) {
	newLog := GrabRedPacketLog{RedPacketId: rpId, SenderName: senderName}
	newLog.ItemList = make([]RewardItem, len(items))
	index := 0
	for itemId, count := range items {
		newLog.ItemList[index] = RewardItem{Id: itemId, Count: int64(count)}
		index++
	}
	rp.addLog(newLog)
	rp.addGrab(rpId)
}

func (rp *GuildRedPacketInfo) addLog(newLog GrabRedPacketLog) {
	if rp.GrabLogList == nil {
		rp.GrabLogList = make([]GrabRedPacketLog, 1)
		rp.GrabLogList[0] = newLog
	} else {
		rp.GrabLogList = append(rp.GrabLogList, newLog)
	}
}

func (rp *GuildRedPacketInfo) addGrab(newRpId string) {
	if rp.TodayGrabList == nil {
		rp.TodayGrabList = make([]string, 1)
		rp.TodayGrabList[0] = newRpId
	} else {
		rp.TodayGrabList = append(rp.TodayGrabList, newRpId)
	}
}

func (rp *GuildRedPacketInfo) CanClaimIpa() bool {
	return rp.IpaStatus == RP_IPA_HAS_PAY
}

func (rp *GuildRedPacketInfo) OnClaim(boxId int64) {
	if rp.TodayClaimList == nil {
		rp.TodayClaimList = make([]int64, 1)
		rp.TodayClaimList[0] = boxId
	} else {
		rp.TodayClaimList = append(rp.TodayClaimList, boxId)
	}
}
