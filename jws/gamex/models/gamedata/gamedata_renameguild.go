package gamedata

import "vcs.taiyouxi.net/jws/gamex/protogen"

var (
	gdGuildRenameData []int
)

func loadGuildRenameData(filepath string) {
	ar := &ProtobufGen.GUILDRENAMECHASE_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	gdGuildRenameData = make([]int, len(data)+1, len(data)+1)
	for _, v := range data {
		gdGuildRenameData[v.GetGuildReNameTime()] = int(v.GetGuildReNamePrice())
	}
}

func GetRenameGuildCost(times int) int64 {
	if times < 1 {
		times = 1
	}
	if times > len(gdGuildRenameData)-1 {
		times = len(gdGuildRenameData) - 1
	}
	return int64(gdGuildRenameData[times])
}
