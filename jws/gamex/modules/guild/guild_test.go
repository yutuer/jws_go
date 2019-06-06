package guild

import (
	"fmt"
	"testing"
)

func TestGuildId(t *testing.T) {
	guildId := genGuildIdByShard(10, 1)
	t.Logf("guildId %v", guildId)
	sgid := GuildItoa(guildId)
	t.Logf("str guildId %v", sgid)
	igid := GuildIdAtoi(sgid)
	t.Logf("int guildId %v", igid)

}

func TestDeleteGuild(t *testing.T) {
	gmw := &GuildMgrWorker{
		guildSlice:    make([]string, 0),
		guildUuid2Idx: make(map[string]int, 0),
	}

	for i := 0; i < 10; i++ {
		gName := fmt.Sprintf("guild:%d", i)
		gmw.guildSlice = append(gmw.guildSlice, gName)
		gmw.guildUuid2Idx[gName] = i
	}

	gmw.delGuildUuid("guild:2")

	fmt.Println(gmw.guildSlice)
	fmt.Println(gmw.guildUuid2Idx)

	for i, gName := range gmw.guildSlice {
		if gmw.guildUuid2Idx[gName] != i {
			t.FailNow()
		}
	}
}
