package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	RecoverTyp_GameMode = 0
	RecoverTyp_Quest    = 1
)

var (
	gdRecoverInfo    map[uint32]*RecoverCfg
	gdRecoverSetting *ProtobufGen.RECOVERSETTINGS
)

type RecoverCfg struct {
	Recover *ProtobufGen.RECOVER
	Retails []*ProtobufGen.RECOVERDETAIL
}

func loadRecover(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.RECOVER_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdRecoverInfo = make(map[uint32]*RecoverCfg, len(ar.GetItems()))
	for _, cfg := range ar.GetItems() {
		gdRecoverInfo[cfg.GetRecoverID()] = &RecoverCfg{
			Recover: cfg,
			Retails: make([]*ProtobufGen.RECOVERDETAIL, 0, 5),
		}
	}
}

func loadRecoverRetail(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.RECOVERDETAIL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	for _, cfg := range ar.GetItems() {
		rec, ok := gdRecoverInfo[cfg.GetRecoverID()]
		if !ok {
			continue
		}
		rec.Retails = append(rec.Retails, cfg)
		gdRecoverInfo[cfg.GetRecoverID()] = rec
	}
}

func loadRecoverSetting(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.RECOVERSETTINGS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	ss := ar.GetItems()
	gdRecoverSetting = ss[0]
}

func GetRecoverCfg(recoverId uint32) *RecoverCfg {
	return gdRecoverInfo[recoverId]
}

func GetAllRecoverCfgs() map[uint32]*RecoverCfg {
	return gdRecoverInfo
}

func GetAllRecoverIds() []uint32 {
	res := make([]uint32, 0, len(gdRecoverInfo))
	for k, _ := range gdRecoverInfo {
		res = append(res, k)
	}
	return res
}

func GetRecoverSetting() *ProtobufGen.RECOVERSETTINGS {
	return gdRecoverSetting
}
