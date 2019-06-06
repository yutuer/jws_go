package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	ChiefMaxAbsentTime   int64  // 会长最大离线时间, 超过时间可能会自动退位
	GuildSafeTimeOnAwake int64  // 公会从睡眠状态被唤醒的时候保护时间
	gdGuildWorshipReward string //被膜拜的人的奖励
	gdGuildWorshipCount  uint32
)

func loadGuildConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	gdGuildRankingData = make([][Guild_Pos_Count]string, len(data), len(data))
	for _, d := range data {
		ChiefMaxAbsentTime = int64(int(d.GetAutoResignHour()) * 60 * 60)
		GuildSafeTimeOnAwake = int64(int(d.GetAutoCheckResignTime()) * 60 * 60)
		gdGuildWorshipReward = d.GetWorshipRewardID()
		gdGuildWorshipCount = d.GetWorshipRewardNum()
	}

}

func GetWorshipReward(num uint32) map[string]uint32 {
	reward := map[string]uint32{
		gdGuildWorshipReward: gdGuildWorshipCount * num,
	}
	return reward
}
