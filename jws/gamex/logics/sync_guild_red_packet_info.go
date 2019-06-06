package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/jws/gamex/uutil"
)

type GuildRedPacket2Client struct {
	TodayGrabCount     int64    `codec:"grp_grab"`       // 今日领过的红包数
	TodayClaimList     []int64  `codec:"grp_claim_list"` // 今日领取过的宝箱进度
	GuildRedPacketList [][]byte `codec:"grp_guild"`      // 当前军团红包
}

// GuildRedPacket 用于生成sync部分代码
type GuildRedPacketDetail2Client struct {
	Id         string `codec:"grp_id"`        // 红包ID
	SendName   string `codec:"grp_send_name"` // 红包发放者名字
	HasClaimed bool   `codec:"grp_is_claim"`  // 自己是否领过
}

// RedPacketInfo 红包RedPacket信息
type redPacketInfo struct {
	ActivityID int64    `codec:"ac_id"`      // 活动ID
	RewardType int64    `codec:"r_type"`     // 奖励类型
	FCValue1   int64    `codec:"fc_one"`     // 完成条件参数1
	FCValue2   int64    `codec:"fc_two"`     // 完成条件参数2
	ItemID     []string `codec:"item_id"`    // 奖励物品ID
	ItemCount  []int64  `codec:"item_count"` // 奖励物品数量
}

func (s *SyncResp) makeGuildPacketItemCondition(now_t int64, channelId string) {
	hotRedPacketInfo := gamedata.GetHotDatas().RedPacketConfig.RewardConfigs
	if len(gamedata.GetHotDatas().Activity.GetActivityInfoFilterTime(gamedata.ActRedPacket, now_t)) == 0 {
		return
	}
	activeId := gamedata.GetHotDatas().Activity.GetActivityInfoFilterTime(gamedata.ActRedPacket, now_t)[0].ActivityId
	s.RedPacketInfo = make([][]byte, 0)
	for _, info := range hotRedPacketInfo {
		if info.GetActivityID() == activeId {
			s.RedPacketInfo = append(s.RedPacketInfo, encode(convertGpItemCondition2Client(info, channelId)))
		}
	}
	logs.Debug("hotRedPacketINfo %v", s.RedPacketInfo)

}

func convertGpItemCondition2Client(value *ProtobufGen.REDPACKET, channelId string) redPacketInfo {
	boxItem := make([]string, 0)
	boxItemNum := make([]int64, 0)
	for _, key := range value.GetItem_Table() {
		boxItem = append(boxItem, key.GetItemID())
		boxItemNum = append(boxItemNum, int64(key.GetItemCount()))
	}
	if channelId == util.Android_Enjoy_Korea_OneStore_Channel {
		return redPacketInfo{
			ActivityID: int64(value.GetActivityID()),
			RewardType: int64(value.GetRewardType()),
			FCValue1:   int64(value.GetFCValue1()),
			FCValue2:   int64(value.GetFCValue2() + uutil.IAPID_ONESTORE_2_GOOGLE),
			ItemID:     boxItem,
			ItemCount:  boxItemNum,
		}
	} else {
		return redPacketInfo{
			ActivityID: int64(value.GetActivityID()),
			RewardType: int64(value.GetRewardType()),
			FCValue1:   int64(value.GetFCValue1()),
			FCValue2:   int64(value.GetFCValue2()),
			ItemID:     boxItem,
			ItemCount:  boxItemNum,
		}
	}

}

// g may be nil
func (s *SyncRespNotify) mkGuildRedPacket(p *Account, g *guild.GuildInfo) {
	infoForClient := buildRedPacket2Client(p, g)
	s.SyncGuildRedPacket = encode(infoForClient)
	s.SyncRedPacketIpaStatus = int(p.GuildProfile.RedPacketInfo.IpaStatus)
	p.GuildProfile.RedPacketInfo.Sync.SetHadSync()
	s.SyncGuildRedPacketNeed = true // 由sync引起的同步需要设置这个值
}

func buildRedPacket2Client(p *Account, g *guild.GuildInfo) GuildRedPacket2Client {
	retInfo := GuildRedPacket2Client{}
	redPacketInfo := &p.GuildProfile.RedPacketInfo
	retInfo.TodayGrabCount = int64(len(redPacketInfo.GrabLogList))
	retInfo.TodayClaimList = redPacketInfo.TodayClaimList
	retInfo.GuildRedPacketList = buildRPDetailList2Client(p, g)
	return retInfo
}

func buildRPDetailList2Client(p *Account, g *guild.GuildInfo) [][]byte {
	if g == nil {
		return make([][]byte, 0)
	}
	retBytes := make([][]byte, len(g.GuildRedPacket.GuildRpList))
	for i, rp := range g.GuildRedPacket.GuildRpList {
		rpDetail := GuildRedPacketDetail2Client{}
		rpDetail.Id = rp.Id
		rpDetail.SendName = rp.SenderName
		rpDetail.HasClaimed = hasClaimed(p, rpDetail.Id)
		retBytes[i] = encode(rpDetail)
	}
	return retBytes
}

func hasClaimed(p *Account, rpId string) bool {
	for _, grabLog := range p.GuildProfile.RedPacketInfo.GrabLogList {
		if grabLog.RedPacketId == rpId {
			return true
		}
	}
	return false
}
