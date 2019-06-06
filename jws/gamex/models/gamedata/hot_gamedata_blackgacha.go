package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 运营活动定向宝箱(黑盒宝箱)
type BlackGachaData struct {
	BlackGachaSettings []*ProtobufGen.BOXSETTINGS
	BlackGachaShow     []*ProtobufGen.BOXSHOW
	BlackGachaLowest   []*ProtobufGen.BOXLOWEST
}

// settings表
type hotBlackGachaSettingsMng struct {
}

func (act *hotBlackGachaSettingsMng) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.BOXSETTINGS_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		logs.Error("load hot gacha boxsettings err", err)
		return err
	}
	datas.HotBlackGachaData.BlackGachaSettings = dataList.Items
	return nil
}

// show表
type hotBlackGachaShowMng struct {
}

func (act *hotBlackGachaShowMng) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.BOXSHOW_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		logs.Error("load hot gacha show err", err)
		return err
	}
	datas.HotBlackGachaData.BlackGachaShow = dataList.Items
	return nil
}

// lowest表
type hotBlackGachaLowestMng struct {
}

func (act *hotBlackGachaLowestMng) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.BOXLOWEST_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		logs.Error("load hot gacha show err", err)
		return err
	}
	datas.HotBlackGachaData.BlackGachaLowest = dataList.Items
	return nil
}

func (act *BlackGachaData) GetBlackGachaSettingsCfg(actId, subId uint32) *ProtobufGen.BOXSETTINGS {
	for _, data := range act.BlackGachaSettings {
		if data.GetActivityID() == actId && data.GetActivitySubID() == subId {
			return data
		}
	}
	return nil
}

func (act *BlackGachaData) GetAllSubBlackGachaLowest(subId uint32) []*ProtobufGen.BOXLOWEST {
	ret := make([]*ProtobufGen.BOXLOWEST, 0)
	for _, data := range act.BlackGachaLowest {
		if data.GetActivitySubID() == subId {
			ret = append(ret, data)
		}
	}
	return ret
}

func (act *BlackGachaData) GetBlackGachaLowest(subId uint32, count uint32) *ProtobufGen.BOXLOWEST {
	for _, data := range act.BlackGachaLowest {
		if data.GetActivitySubID() == subId && data.GetLowestTimes() == count {
			return data
		}
	}
	return nil
}

func (act *BlackGachaData) GetAllSubGachaSettingsCfg(actId uint32) []*ProtobufGen.BOXSETTINGS {
	ret := make([]*ProtobufGen.BOXSETTINGS, 0)
	for _, data := range act.BlackGachaSettings {
		if data.GetActivityID() == actId {
			ret = append(ret, data)
		}
	}
	return ret
}
