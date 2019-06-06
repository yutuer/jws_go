package guild_info

import (
	"fmt"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 公会所有红包信息
type GuildRedPacketInfo struct {
	GuildRpList        []GuildRedPacket
	RedPacketIndex     int64 // 公会红包ID自增索引
	LastDailyResetTime int64 // 每日重置时间
}

// 公会红包
type GuildRedPacket struct {
	Id           string //
	SenderName   string // 发送者姓名
	TimeStamp    int64  // 红包产生的时间戳
	GrabRPRecord []GrabRPRecord
}

// 红包领取记录
type GrabRPRecord struct {
	Acid       string // 领红包的角色
	PlayerName string // 领红包的角色
	RewardList []RewardItem
}

// 奖励物品
type RewardItem struct {
	ItemId string
	Count  uint32
}

func (rpInfo *GuildRedPacketInfo) NewGuildRedPacket(guildUuid string, senderName string) *GuildRedPacket {
	rpInfo.RedPacketIndex++
	rp := &GuildRedPacket{
		Id:         fmt.Sprintf("%s:%d", guildUuid, rpInfo.RedPacketIndex),
		SenderName: senderName,
		TimeStamp:  time.Now().Unix(),
	}
	rpInfo.add(rp)
	return rp
}

func (rpInfo *GuildRedPacketInfo) add(rp *GuildRedPacket) {
	if rpInfo == nil {
		rpInfo.GuildRpList = make([]GuildRedPacket, 1)
		rpInfo.GuildRpList[0] = *rp
	} else {
		rpInfo.GuildRpList = append(rpInfo.GuildRpList, *rp)
	}
}

func (rpInfo *GuildRedPacketInfo) CheckDailyReset(now int64) {
	//logs.Debug("guild red packet CheckDailyReset, %d, %d", now, rpInfo.LastDailyResetTime)
	if !gamedata.IsSameDayCommon(now, rpInfo.LastDailyResetTime) {
		rpInfo.DailyReset(now)
		logs.Debug("guild daily reset red pakcet")
	}
}

func (rpInfo *GuildRedPacketInfo) DailyReset(nowTime int64) {
	rpInfo.LastDailyResetTime = nowTime
	rpInfo.GuildRpList = nil
}

func (rpInfo *GuildRedPacketInfo) Get(id string) (*GuildRedPacket, bool) {
	if rpInfo.GuildRpList == nil {
		return nil, false
	}
	for i, rp := range rpInfo.GuildRpList {
		if rp.Id == id {
			return &rpInfo.GuildRpList[i], true
		}
	}
	return nil, false
}

func (rpInfo *GuildRedPacketInfo) GetRpCount() int {
	return len(rpInfo.GuildRpList)
}

//
func (rpInfo *GuildRedPacketInfo) DebugCleanPlayer(name string) {
	// 删除个人的红包
	newList := make([]GuildRedPacket, 0, len(rpInfo.GuildRpList))
	for _, rp := range rpInfo.GuildRpList {
		if rp.SenderName != name {
			newList = append(newList, rp)
		}
	}
	rpInfo.GuildRpList = newList
	// 删除每个红包的领取记录
	for i, rp := range rpInfo.GuildRpList {
		findIndex := -1
		for i, log := range rp.GrabRPRecord {
			if log.PlayerName == name {
				findIndex = i
				break
			}
		}
		if findIndex != -1 {
			rpInfo.GuildRpList[i].GrabRPRecord = append(rp.GrabRPRecord[0:findIndex], rp.GrabRPRecord[findIndex+1:]...)
		}
	}
}

func (rpInfo *GuildRedPacketInfo) ContainsBySenderName(name string) bool {
	for _, rp := range rpInfo.GuildRpList {
		if rp.SenderName == name {
			return true
		}
	}
	return false
}

func (rp *GuildRedPacket) Contains(name string) bool {
	for _, r := range rp.GrabRPRecord {
		if r.PlayerName == name {
			return true
		}
	}
	return false
}

func (rp *GuildRedPacket) AddGrab(act uint32, acid, name string) GrabRPRecord {
	record := GrabRPRecord{Acid: acid, PlayerName: name}
	config := gamedata.GetHotDatas().RedPacketConfig.GetRandomGrabConfig(act)
	record.RewardList = make([]RewardItem, len(config.GetItem_Table()))
	for i, item := range config.GetItem_Table() {
		record.RewardList[i] = RewardItem{ItemId: item.GetItemID(), Count: item.GetItemCount()}
	}
	if rp.GrabRPRecord == nil {
		rp.GrabRPRecord = make([]GrabRPRecord, 1)
		rp.GrabRPRecord[0] = record
	} else {
		rp.GrabRPRecord = append(rp.GrabRPRecord, record)
	}
	return record
}
