package gamedata

import (
	"math/rand"
	"time"

	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gDDroidDatasSimplePvp  []DroidAccountData
	gDDroidDatasTeamPvp    map[uint32]DroidAccountData
	gDDroidDatasExpedition map[uint32]DroidAccountData
	gdDroidDatasGVG        map[uint32]DroidAccountData
	gDDroidDatasWsPvp      map[uint32]DroidAccountData
)

var (
	gdMaxLvlGVGGuard uint32
	gdMinLvlGVGGuard uint32
)

func loadDroidDatas(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.BSCPVPBOT_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	gDDroidDatasSimplePvp = make([]DroidAccountData, 0, len(ar.Items))
	gDDroidDatasTeamPvp = make(map[uint32]DroidAccountData, len(ar.Items))
	gDDroidDatasWsPvp = make(map[uint32]DroidAccountData)
	for _, data := range ar.Items {
		switch int(data.GetBottype()) {
		case 0:
			d := DroidAccountData{}
			d.FromData(data)
			gDDroidDatasSimplePvp = append(gDDroidDatasSimplePvp, d)
		case 1:
			d := DroidAccountData{}
			d.FromData(data)
			gDDroidDatasTeamPvp[data.GetBotID()] = d
		case 2:
			d := DroidAccountData{}
			d.FromData(data)
			gDDroidDatasWsPvp[data.GetBotID()] = d
		default:
			panic(fmt.Errorf("loadDroidDatas Bottype %d not define", data.GetBottype()))
		}
	}

	logs.Trace("gDDroidDatas %v", gDDroidDatasSimplePvp)

}

func loadExpeditionDroidDatas(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.EXPEDITIONBOT_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	gDDroidDatasExpedition = make(map[uint32]DroidAccountData, len(ar.Items))
	for _, data := range ar.Items {
		d := DroidAccountData{}
		d.FromDataExpedition(data)
		gDDroidDatasExpedition[data.GetCLv()] = d

	}

	logs.Trace("gDDroidDatas %v", gDDroidDatasExpedition)
}

func loadGVGDroidDatas(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.GVGGUARD_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)
	gdMaxLvlGVGGuard = 0
	gdMinLvlGVGGuard = 1000000 // 很大了,不会超过这个级数的
	gdDroidDatasGVG = make(map[uint32]DroidAccountData, len(ar.Items))
	for _, data := range ar.Items {
		d := DroidAccountData{}
		d.FromDataGVG(data)
		lv := data.GetCLv()
		gdDroidDatasGVG[lv] = d
		if lv > gdMaxLvlGVGGuard {
			gdMaxLvlGVGGuard = lv
		}
		if lv < gdMinLvlGVGGuard {
			gdMinLvlGVGGuard = lv
		}
	}
}

func GetRandDroidForSimplePvp() *DroidAccountData {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &gDDroidDatasSimplePvp[r.Intn(len(gDDroidDatasSimplePvp))]
}

func GetDroidForTeamPvp(droidId uint32) *DroidAccountData {
	if res, ok := gDDroidDatasTeamPvp[droidId]; ok {
		return &res
	}
	return nil
}

func GetDroidForExpedition(droidLvl uint32) *DroidAccountData {
	if res, ok := gDDroidDatasExpedition[droidLvl]; ok {
		return &res
	}
	return nil
}

func GetDroidForGVG(droidLvl uint32) *DroidAccountData {
	if droidLvl > gdMaxLvlGVGGuard {
		droidLvl = gdMaxLvlGVGGuard
	} else if droidLvl < gdMinLvlGVGGuard {
		droidLvl = gdMinLvlGVGGuard
	}
	if res, ok := gdDroidDatasGVG[droidLvl]; ok {
		return &res
	}
	logs.Error("Fatal Error, no Droid Data for LV: %d", droidLvl)
	res := gdDroidDatasGVG[gdMaxLvlGVGGuard]
	return &res
}

func GetDroidForWspvp(droidId uint32) *DroidAccountData {
	if res, ok := gDDroidDatasWsPvp[droidId]; ok {
		return &res
	}
	return nil
}
