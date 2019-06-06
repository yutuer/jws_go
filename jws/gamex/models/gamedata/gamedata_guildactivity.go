package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdGuildActivityUnlockLvl map[uint32]uint32
)

const (
	GuildActivity_Null = iota
	GuildActivity_TERM_SYS_GVE
	GuildActivity_GUILDBOSS_TITLE
	GuildActivity_GUILDWAR_TITLE
	GuildActivity_GUILD_CONTRIBUTION_NAME
	GuildActivity_GUILD_WORSHIPCRIT_NAME
)

func loadGuildActivity(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDACTIVITY_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGuildActivityUnlockLvl = make(map[uint32]uint32, len(data))
	for _, d := range data {
		gdGuildActivityUnlockLvl[d.GetGuildActivityId()] = d.GetGuildLevelReq()
	}

}

func GetGuildActivityLvl(lvl uint32) uint32 {
	return gdGuildActivityUnlockLvl[lvl]
}
