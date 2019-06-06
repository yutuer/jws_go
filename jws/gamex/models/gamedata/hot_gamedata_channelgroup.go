package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	ALL_TYPE     = 0
	SPECIAL_TYPE = 1
)

type hotActivityChannelGroupData struct {
}

func (sg *hotActivityChannelGroupData) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.CHANNELGROUP_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.gdChannelGroup = make(map[uint32][]string, len(dataList.GetItems()))
	for _, r := range dataList.GetItems() {
		if r.GetChannelGroupType() == ALL_TYPE {
			// ALL_TYPE: nil
		} else if r.GetChannelGroupType() == SPECIAL_TYPE {
			// SPECIAL_TYPE: []
			ids := make([]string, 0)
			for _, item := range r.GetAccCon_Table() {
				ids = append(ids, item.GetChannelGroupValue())
			}
			datas.Activity.gdChannelGroup[r.GetChannelGroupID()] = ids
		}

	}
	logs.Debug("Load Hot Data ChannelGroup Success: %v", datas.Activity.gdChannelGroup)
	return nil
}
