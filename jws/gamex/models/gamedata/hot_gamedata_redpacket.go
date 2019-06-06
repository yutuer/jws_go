package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Red_Packet_Reward_Iap = iota
	Red_Packet_Reward_Box
	Red_Packet_Reward_Grab
)

type RedPacketConfig []*ProtobufGen.REDPACKET

func (rpCfg RedPacketConfig) GetWeight(index int) int {
	return int(rpCfg[index].GetFCValue1())
}

func (rpCfg RedPacketConfig) Len() int {
	return len(rpCfg)
}

type hotRedPacketData struct {
	IapSet        map[uint32]bool
	RewardConfigs []*ProtobufGen.REDPACKET
}

type hotRedPacketMng struct {
}

func (act *hotRedPacketMng) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.REDPACKET_ARRAY{}
	datas.RedPacketConfig.IapSet = make(map[uint32]bool)
	datas.RedPacketConfig.RewardConfigs = make([]*ProtobufGen.REDPACKET, 0)
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		logs.Error("load red packet error, %v", err)
		return err
	}
	for _, item := range dataList.GetItems() {
		datas.RedPacketConfig.RewardConfigs = append(datas.RedPacketConfig.RewardConfigs, item)
		if item.GetRewardType() == Red_Packet_Reward_Iap {
			iapIndex1, iapIndex2 := item.GetFCValue1(), item.GetFCValue2()
			if iapIndex1 != 0 {
				datas.RedPacketConfig.IapSet[iapIndex1] = true
			}
			if iapIndex2 != 0 {
				datas.RedPacketConfig.IapSet[iapIndex2] = true
			}
		}
	}
	logs.Info("datas.RedPacketConfig.IapSet: %v", datas.RedPacketConfig.IapSet)
	return nil
}

func (act *hotRedPacketData) GetRpBoxConfig(index int, ActivityID uint32) *ProtobufGen.REDPACKET {
	for _, config := range act.RewardConfigs {
		if config.GetActivityID() == ActivityID && config.GetRewardType() == Red_Packet_Reward_Box &&
			config.GetFCValue1() == uint32(index) {
			return config
		}
	}
	return nil
}

func (act *hotRedPacketData) GetRandomGrabConfig(ActivityID uint32) *ProtobufGen.REDPACKET {
	var tempList []*ProtobufGen.REDPACKET
	tempList = make([]*ProtobufGen.REDPACKET, 0)
	for _, config := range act.RewardConfigs {
		if config.GetActivityID() == ActivityID && config.GetRewardType() == Red_Packet_Reward_Grab {
			tempList = append(tempList, config)
		}
	}
	configList := RedPacketConfig(tempList)
	index := util.RandomItem(configList)
	return configList[index]
}

func (act *hotRedPacketData) GetIpaConfig(acid uint32) *ProtobufGen.REDPACKET {
	for _, config := range act.RewardConfigs {
		if config.GetActivityID() == acid && config.GetRewardType() == Red_Packet_Reward_Iap {
			return config
		}
	}
	return nil
}
