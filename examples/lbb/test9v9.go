package main

import (
	"fmt"
	"os"
	"vcs.taiyouxi.net/jws/gamex/models/codec"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

/*
	// 个人基本信息 最大值 10000
	hashMap WarInfo
	排行榜信息
	实际战斗信息

	// 排名信息
	zset
	key=acid, score=rank

	// 被锁角色信息
	hashmap LockInfo
	key=acid, value=endtime

	// 战报信息 10000
	list
*/

// TODO 删除无用信息
type WarInfo struct {
	// 角色基础属性
	Acid      string `redis:"acid"`  // account id
	Name      string `redis:"name"`  // name
	AvatarId  int    `redis:"aid"`   // 当前使用形象ID
	CorpLv    uint32 `redis:"clv"`   // 战队等级
	GuildName string `redis:"gname"` // 公会名字

	// 9个武将信息
	Attr  [9]HeroInfo `redis:"attr"`  // 9个武将的战斗属性
	Group [9]int      `redis:"group"` // 9个位置的武将ID 如果更改阵型产生新武将需要同步attr
}

type HeroInfo struct {
	Idx       int    // id
	Attr      []byte // 战斗属性
	StarLevel int    // 星级
	Gs        int    // 战力

	AvatarSkills   [4]uint32 `json:"skills"`        // 武将技能等级
	AvatarFashion  [4]string `json:"avatar_equips"` // 武将时装
	HeroSwing      int       `json:"swing"`         // 翅膀
	PassiveSkillId [4]string `json:"pskillid"`      // 被动技能
	CounterSkillId [4]string `json:"cskillid"`
	TriggerSkillId [4]string `json:"tskillid"`
}

func main() {
	conn, err2 := redis.Dial("tcp", "127.0.0.1:6379", redis.DialDatabase(8), redis.DialPassword(""))
	if err2 != nil {
		fmt.Println(err2.Error())
		os.Exit(1)
	}
	cb := redis.NewCmdBuffer()
	for i := 0; i < 10000; i++ {
		warInfo := buildWarInfo(i)
		driver.DumpToHashDBCmcBuffer(cb, fmt.Sprintf("WarInfo:%d", i), warInfo)
	}
	conn.DoCmdBuffer(cb, true)
}

func buildWarInfo(i int) *WarInfo {
	warInfo := new(WarInfo)
	warInfo.Acid = fmt.Sprintf("warinfo:0:%d:6c5119b9-0dd6-445d-b4da-a35c4ffed42b", i)
	warInfo.Name = fmt.Sprintf("robot%d", i)
	warInfo.GuildName = "forfreedom"
	for j := 0; j < 9; j++ {
		warInfo.Attr[j] = buildHeroInfo(j)
	}
	return warInfo
}

func buildHeroInfo(i int) HeroInfo {
	heroInfo := HeroInfo{}
	heroInfo.Idx = i
	heroInfo.Attr = codec.Encode(gamedata.AvatarAttr{})
	heroInfo.PassiveSkillId = [4]string{"111111", "2222222", "333333", "444444"}
	heroInfo.CounterSkillId = [4]string{"111111", "2222222", "333333", "444444"}
	heroInfo.TriggerSkillId = [4]string{"111111", "2222222", "333333", "444444"}
	return heroInfo
}
