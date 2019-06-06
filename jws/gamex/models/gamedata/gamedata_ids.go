package gamedata

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	gdServerIds map[string]*ProtobufGen.COMMONIDS
)

func loadIDSData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.COMMONIDS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdServerIds = make(map[string]*ProtobufGen.COMMONIDS, len(data))

	for _, d := range data {
		gdServerIds[d.GetIDS()] = d
	}
}

func GetCommonIdsStr(ids string, param ...string) string {
	switch game.Cfg.Lang {
	case uutil.Lang_HANS:
		config, ok := gdServerIds[ids]
		if ok {
			res := config.GetZhHans()
			for i, p := range param {
				res = strings.Replace(res, fmt.Sprintf("{%d}", i), p, 1)
			}
			return res
		}
		return ""
	default:
		return ""
	}
}
