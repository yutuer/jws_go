package gamedata

import (
	"math/rand"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdGankConf        *ProtobufGen.GANKCONFIG
	gdGankSysNotice   []int
	gdGankLogIDSCount int
)

func loadGankData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GANKCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGankConf = data[0]

	gdGankSysNotice = []int{
		IDS_GANK_WIN_1,
		IDS_GANK_WIN_2,
		IDS_GANK_WIN_3}
}

func loadGankIDS(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GANKIDS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	for _, d := range ar.GetItems() {
		if d.GetType() == 0 {
			gdGankLogIDSCount++
		}
	}
}

func GetGankConf() *ProtobufGen.GANKCONFIG {
	return gdGankConf
}

func RankSysNotice(r *rand.Rand) int {
	i := r.Intn(len(gdGankSysNotice))
	return gdGankSysNotice[i]
}

func RandGankLogIDS(r *rand.Rand) int {
	return r.Intn(gdGankLogIDSCount)
}

func GankIDSCount() int {
	return gdGankLogIDSCount
}
