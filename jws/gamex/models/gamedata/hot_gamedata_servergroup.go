package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type hotServerGroupMng struct {
}

func (sg *hotServerGroupMng) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.SEVERGROUP_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdServerGroup = make(map[uint32]uint32, len(dataList.GetItems()))
	datas.Activity.gdServerGroupSbatch = make(map[uint32]*ProtobufGen.SEVERGROUP, len(dataList.GetItems()))
	datas.Activity.gdWuShuangGroup = make(map[uint32][]uint32, len(dataList.GetItems()))
	datas.Activity.gdCSRobGroup = make(map[uint32][]uint32, len(dataList.GetItems()))
	datas.Activity.gdWBGroup = make(map[uint32][]uint32, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		datas.Activity.gdServerGroup[r.GetSID()] = r.GetGroupID()
		datas.Activity.gdServerGroupSbatch[r.GetSID()] = r
		datas.Activity.gdWuShuangGroup[r.GetWspvpGroupID()] = append(datas.Activity.gdWuShuangGroup[r.GetWspvpGroupID()],
			r.GetSID())
		datas.Activity.gdCSRobGroup[r.GetRobCropsGroupID()] = append(datas.Activity.gdCSRobGroup[r.GetRobCropsGroupID()], r.GetSID())
		datas.Activity.gdWBGroup[r.GetWorldBossGroupID()] = append(datas.Activity.gdWBGroup[r.GetWorldBossGroupID()], r.GetSID())
	}

	logs.Debug("Load Hot Data ServerGroup Success")
	return nil
}

type hotServerGroupActivity struct {
}

func (sga *hotServerGroupActivity) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.SGACTIVITY_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdServerGroupActivity = make(map[uint32][]uint32, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		if r.GetHotActivityID() == nil {
			return (fmt.Errorf("ServeGroupActivity HotActivityID is nil, %d", r.GetGroupID()))
		}
		datas.Activity.gdServerGroupActivity[r.GetGroupID()] = r.GetHotActivityID()
	}
	logs.Debug("Load Hot Data ServerGroupActivity Success")
	return nil
}

func (d hotActivityData) GetShardActivities(shard uint32) []uint32 {
	groupId, ok := d.gdServerGroup[shard]
	if !ok {
		return []uint32{}
	}
	ret, ok := d.gdServerGroupActivity[groupId]
	if !ok {
		return []uint32{}
	}
	return ret
}

func (d hotActivityData) GetShardGroup(shard uint32) uint32 {
	return d.gdServerGroup[shard]
}

func (d hotActivityData) GetServerGroupSbatch(shard uint32) *ProtobufGen.SEVERGROUP {
	return d.gdServerGroupSbatch[shard]
}

func GetWSPVPGroupId(sid uint32) uint32 {
	return GetHotDatas().Activity.GetServerGroupSbatch(sid).GetWspvpGroupID()
}

func GetWSPVPGroupCfg(sid uint32) *ProtobufGen.SEVERGROUP {
	return GetHotDatas().Activity.GetServerGroupSbatch(sid)
}

func GetWSPVPSids(groupId uint32) []uint32 {
	return GetHotDatas().Activity.gdWuShuangGroup[groupId]
}

func GetCSRobGroupId(sid uint32) uint32 {
	return GetHotDatas().Activity.GetServerGroupSbatch(sid).GetRobCropsGroupID()
}

func GetCSRobSids(groupId uint32) []uint32 {
	return GetHotDatas().Activity.gdCSRobGroup[groupId]
}

func GetWBGroupId(sid uint32) uint32 {
	return GetHotDatas().Activity.GetServerGroupSbatch(sid).GetWorldBossGroupID()
}

func GetTBGroupId(sid uint32) uint32 {
	return GetHotDatas().Activity.GetServerGroupSbatch(sid).GetTeamBossGroupID()
}

func GetWBGSids(groupId uint32) []uint32 {
	return GetHotDatas().Activity.gdWBGroup[groupId]
}

//GetWBGroupIDRange ..
func GetWBGroupIDRange(sidMin uint32, sidMax uint32) []int {
	if sidMax < sidMin {
		return []int{}
	}
	list := make([]int, 0)
	for i := sidMin; i <= sidMax; i++ {
		cfg := GetHotDatas().Activity.GetServerGroupSbatch(i)
		if nil != cfg {
			list = append(list, int(cfg.GetWorldBossGroupID()))
		}
	}
	return list
}

//GetWBGroupIDRange ..
func GetTBGroupIDRange(sidMin uint32, sidMax uint32) []int {
	if sidMax < sidMin {
		return []int{}
	}
	list := make([]int, 0)
	for i := sidMin; i <= sidMax; i++ {
		cfg := GetHotDatas().Activity.GetServerGroupSbatch(i)
		if nil != cfg {
			list = append(list, int(cfg.GetTeamBossGroupID()))
		}
	}
	return list
}


func (d hotActivityData) GetHGRHotID(sid uint32) uint32 {
	return d.GetServerGroupSbatch(sid).GetHGRHotID()
}
